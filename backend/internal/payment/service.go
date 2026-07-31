package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	entSQL "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/shuTwT/nex-api/backend/internal/config"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/payment"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/systemsetting"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/runtime"
)

type Service struct {
	client    *ent.Client
	clock     Clock
	providers map[PaymentMethod]Provider
	config    config.Payment
}

func NewService(client *ent.Client, paymentConfig config.Payment, clocks ...Clock) (*Service, error) {
	return NewServiceWithProviders(client, paymentConfig, []Provider{
		NewWeChatProvider(paymentConfig.WeChat, nil),
		NewAlipayProvider(paymentConfig.Alipay, nil),
		NewMockProvider(paymentConfig.Mock),
	}, clocks...)
}

func NewPaymentService(client *ent.Client, paymentConfig config.Payment, clocks ...Clock) (*Service, error) {
	return NewService(client, paymentConfig, clocks...)
}

func NewServiceWithProviders(client *ent.Client, paymentConfig config.Payment, providers []Provider, clocks ...Clock) (*Service, error) {
	if client == nil {
		return nil, errors.New("payment: ent client is nil")
	}
	if len(providers) == 0 {
		return nil, errors.New("payment: at least one provider is required")
	}
	providerMap := make(map[PaymentMethod]Provider, len(providers))
	for _, provider := range providers {
		if provider == nil {
			return nil, errors.New("payment: provider is nil")
		}
		providerMap[provider.Method()] = provider
	}
	clock := Clock(realClock{})
	if len(clocks) > 0 && clocks[0] != nil {
		clock = clocks[0]
	}
	return &Service{client: client, clock: clock, providers: providerMap, config: paymentConfig}, nil
}

func (s *Service) CreatePayment(ctx context.Context, input CreatePaymentInput) (CreatePaymentResult, error) {
	if err := validateContext(ctx); err != nil {
		return CreatePaymentResult{}, err
	}
	if strings.TrimSpace(input.UserID) == "" {
		return CreatePaymentResult{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "userId", Reason: "required"})
	}
	if input.Amount <= 0 {
		return CreatePaymentResult{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "amount", Reason: "must be positive"})
	}
	if input.Currency == "" {
		input.Currency = "CNY"
	}
	if strings.ToUpper(input.Currency) != "CNY" {
		return CreatePaymentResult{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "currency", Reason: "only CNY is supported"})
	}
	provider, ok := s.providers[input.Method]
	if !ok {
		return CreatePaymentResult{}, fmt.Errorf("unsupported payment method %q", input.Method)
	}
	outTradeNo := "PAY-" + strings.ToUpper(uuid.NewString())
	createdAt := s.clock.Now().UTC()
	expiredAt := createdAt.Add(2 * time.Hour)
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("marshal payment metadata: %w", err)
	}
	providerResult, err := provider.Create(ctx, ProviderCreateRequest{OutTradeNo: outTradeNo, Amount: input.Amount, Currency: input.Currency, NotifyURL: input.NotifyURL})
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("create %s payment: %w", input.Method, err)
	}
	created, err := s.client.Payment.Create().SetUserId(input.UserID).SetOutTradeNo(outTradeNo).
		SetMethod(string(input.Method)).SetAmount(input.Amount).SetCurrency(input.Currency).
		SetStatus(string(PaymentStatePending)).SetCreatedAt(createdAt).SetUpdatedAt(createdAt).
		SetExpiredAt(expiredAt).SetMetadata(string(metadata)).SetNotifyUrl(input.NotifyURL).
		SetQrcodeUrl(providerResult.QRCodeURL).SetPayUrl(providerResult.PayURL).Save(ctx)
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("store payment order: %w", err)
	}
	return CreatePaymentResult{Success: true, PaymentID: created.ID, OutTradeNo: created.OutTradeNo, QRCodeURL: created.QrcodeUrl, PayURL: created.PayUrl}, nil
}

func (s *Service) CreateSubscriptionPayment(ctx context.Context, userID, planID string, method PaymentMethod) (CreatePaymentResult, error) {
	plan, err := s.client.SubscriptionPlan.Get(ctx, planID)
	if ent.IsNotFound(err) {
		return CreatePaymentResult{}, fmt.Errorf("subscription plan %q: %w", planID, appRuntime.ErrNotFound)
	}
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("get subscription plan: %w", err)
	}
	if !plan.IsActive {
		return CreatePaymentResult{}, appRuntime.NewAPIError(http.StatusBadRequest, "plan_inactive", "subscription plan is inactive", appRuntime.ErrConflict)
	}
	return s.CreatePayment(ctx, CreatePaymentInput{UserID: userID, Amount: plan.Price, Currency: "CNY", Method: method, PlanID: plan.ID, Metadata: map[string]json.RawMessage{"type": json.RawMessage(`"subscription"`), "planId": json.RawMessage(fmt.Sprintf("%q", plan.ID))}})
}

func (s *Service) GetPayment(ctx context.Context, outTradeNo string) (PaymentView, error) {
	paymentRecord, err := s.client.Payment.Query().Where(payment.OutTradeNo(outTradeNo)).Only(ctx)
	if ent.IsNotFound(err) {
		return PaymentView{}, fmt.Errorf("payment %q: %w", outTradeNo, appRuntime.ErrNotFound)
	}
	if err != nil {
		return PaymentView{}, fmt.Errorf("get payment: %w", err)
	}
	return paymentView(paymentRecord), nil
}

func (s *Service) ListHistory(ctx context.Context, userID string) ([]PaymentView, error) {
	records, err := s.client.Payment.Query().Where(payment.UserId(userID)).Order(payment.ByCreatedAt(entSQL.OrderDesc())).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list payment history: %w", err)
	}
	views := make([]PaymentView, len(records))
	for index, record := range records {
		views[index] = paymentView(record)
	}
	return views, nil
}

func (s *Service) QueryPayment(ctx context.Context, outTradeNo string) (PaymentView, error) {
	current, err := s.client.Payment.Query().Where(payment.OutTradeNo(outTradeNo)).Only(ctx)
	if ent.IsNotFound(err) {
		return PaymentView{}, fmt.Errorf("payment %q: %w", outTradeNo, appRuntime.ErrNotFound)
	}
	if err != nil {
		return PaymentView{}, fmt.Errorf("get payment before query: %w", err)
	}
	provider, ok := s.providers[PaymentMethod(current.Method)]
	if !ok {
		return PaymentView{}, fmt.Errorf("unsupported payment method %q", current.Method)
	}
	status, err := provider.Query(ctx, outTradeNo)
	if err != nil {
		if errors.Is(err, ErrProviderUnavailable) {
			return paymentView(current), nil
		}
		return PaymentView{}, fmt.Errorf("query payment provider: %w", err)
	}
	if status.Status == PaymentStatePending {
		return paymentView(current), nil
	}
	updated, err := s.applyCallback(ctx, PaymentMethod(current.Method), ProviderCallback{OutTradeNo: outTradeNo, TransactionID: status.TransactionID, Amount: status.Amount, HasAmount: true, Currency: status.Currency, Status: status.Status, PaidAt: status.PaidAt, CallbackKey: string(current.Method) + ":query:" + outTradeNo})
	if err != nil {
		return PaymentView{}, err
	}
	return paymentView(updated), nil
}

func (s *Service) CancelPayment(ctx context.Context, outTradeNo string) error {
	current, err := s.client.Payment.Query().Where(payment.OutTradeNo(outTradeNo)).Only(ctx)
	if ent.IsNotFound(err) {
		return fmt.Errorf("payment %q: %w", outTradeNo, appRuntime.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("get payment before cancel: %w", err)
	}
	if current.Status == string(PaymentStatePaid) {
		return fmt.Errorf("cancel paid payment: %w", ErrInvalidTransition)
	}
	provider, ok := s.providers[PaymentMethod(current.Method)]
	if !ok {
		return fmt.Errorf("unsupported payment method %q", current.Method)
	}
	if err := provider.Cancel(ctx, outTradeNo); err != nil {
		return fmt.Errorf("cancel payment provider: %w", err)
	}
	_, err = s.applyCallback(ctx, PaymentMethod(current.Method), ProviderCallback{OutTradeNo: outTradeNo, Status: PaymentStateCancelled, CallbackKey: string(current.Method) + ":cancel:" + outTradeNo})
	return err
}

func (s *Service) HandleCallback(ctx context.Context, method PaymentMethod, request ProviderCallbackRequest) (PaymentView, error) {
	provider, ok := s.providers[method]
	if !ok {
		return PaymentView{}, fmt.Errorf("unsupported payment method %q", method)
	}
	callback, err := provider.VerifyCallback(ctx, request)
	if err != nil {
		return PaymentView{}, err
	}
	updated, err := s.applyCallback(ctx, method, callback)
	if err != nil {
		return PaymentView{}, err
	}
	return paymentView(updated), nil
}

func (s *Service) AvailableMethods() []PaymentMethod {
	methods := make([]PaymentMethod, 0, len(s.providers))
	for _, method := range []PaymentMethod{PaymentMethodWeChat, PaymentMethodAlipay, PaymentMethodMock} {
		if _, ok := s.providers[method]; ok {
			methods = append(methods, method)
		}
	}
	return methods
}

func (s *Service) Settings(ctx context.Context) (PaymentSettings, error) {
	settings, err := s.client.SystemSetting.Query().Where(systemsetting.Category("payment")).All(ctx)
	if err != nil {
		return PaymentSettings{}, fmt.Errorf("get payment settings: %w", err)
	}
	values := make(map[string]string, len(settings))
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}
	return PaymentSettings{CreditPrice: parseSettingFloat(values, "creditPrice", 1), MinRecharge: parseSettingFloat(values, "minRecharge", 10), AlipayEnabled: values["alipayEnabled"] == "true", WeChatEnabled: values["wechatEnabled"] == "true", MockEnabled: s.config.Mock.Enabled}, nil
}

func paymentView(record *ent.Payment) PaymentView {
	view := PaymentView{ID: record.ID, UserID: record.UserId, OutTradeNo: record.OutTradeNo, TransactionID: record.TransactionId, Method: PaymentMethod(record.Method), Amount: record.Amount, Currency: record.Currency, Status: PaymentState(record.Status), QRCodeURL: record.QrcodeUrl, PayURL: record.PayUrl, CreatedAt: record.CreatedAt}
	if !record.PaidAt.IsZero() {
		value := record.PaidAt
		view.PaidAt = &value
	}
	if !record.ExpiredAt.IsZero() {
		value := record.ExpiredAt
		view.ExpiredAt = &value
	}
	if !record.CancelledAt.IsZero() {
		value := record.CancelledAt
		view.CancelledAt = &value
	}
	return view
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("payment: context is nil")
	}
	return nil
}
