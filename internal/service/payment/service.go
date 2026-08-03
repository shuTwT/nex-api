package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	entSQL "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/payment"
	pay "github.com/shuTwT/nex-api/internal/infra/pay"
	appRuntime "github.com/shuTwT/nex-api/internal/service/apierror"
)

type Service struct {
	client          *ent.Client
	clock           pay.Clock
	providers       map[pay.PaymentMethod]pay.Provider
	providerFactory func(pay.PaymentMethod, pay.PaymentConfiguration) pay.Provider
	schedule        func(time.Duration, func())
	appURL          string
}

func NewService(client *ent.Client, appURL string, clocks ...pay.Clock) (*Service, error) {
	if client == nil {
		return nil, errors.New("payment: ent client is nil")
	}
	return &Service{client: client, clock: selectClock(clocks), schedule: scheduleAfter, appURL: strings.TrimSpace(appURL)}, nil
}

func NewPaymentService(client *ent.Client, appURL string, clocks ...pay.Clock) (*Service, error) {
	return NewService(client, appURL, clocks...)
}

func NewServiceWithProviders(client *ent.Client, providers []pay.Provider, clocks ...pay.Clock) (*Service, error) {
	if client == nil {
		return nil, errors.New("payment: ent client is nil")
	}
	if len(providers) == 0 {
		return nil, errors.New("payment: at least one provider is required")
	}
	providerMap := make(map[pay.PaymentMethod]pay.Provider, len(providers))
	for _, provider := range providers {
		if provider == nil {
			return nil, errors.New("payment: provider is nil")
		}
		providerMap[provider.Method()] = provider
	}
	return &Service{client: client, clock: selectClock(clocks), providers: providerMap, schedule: scheduleAfter}, nil
}

func scheduleAfter(delay time.Duration, callback func()) {
	time.AfterFunc(delay, callback)
}

func selectClock(clocks []pay.Clock) pay.Clock {
	if len(clocks) > 0 && clocks[0] != nil {
		return clocks[0]
	}
	return pay.RealClock{}
}

// createPayment is an internal provider primitive. HTTP callers must enter
// through a business-specific flow that validates pricing and builds metadata
// on the server.
func (s *Service) createPayment(ctx context.Context, input createPaymentInput) (CreatePaymentResult, error) {
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
	provider, notifyURL, err := s.resolveProvider(ctx, input.Method)
	if err != nil {
		return CreatePaymentResult{}, err
	}
	if strings.TrimSpace(input.NotifyURL) == "" {
		input.NotifyURL = notifyURL
	}
	outTradeNo := "PAY-" + strings.ToUpper(uuid.NewString())
	createdAt := s.clock.Now().UTC()
	expiredAt := createdAt.Add(2 * time.Hour)
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("marshal payment metadata: %w", err)
	}
	providerResult, err := provider.Create(ctx, pay.ProviderCreateRequest{OutTradeNo: outTradeNo, Amount: input.Amount, Currency: input.Currency, NotifyURL: input.NotifyURL})
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("create %s payment: %w", input.Method, err)
	}
	created, err := s.client.Payment.Create().SetUserId(input.UserID).SetOutTradeNo(outTradeNo).
		SetMethod(string(input.Method)).SetAmount(input.Amount).SetCurrency(input.Currency).
		SetStatus(string(pay.PaymentStatePending)).SetCreatedAt(createdAt).SetUpdatedAt(createdAt).
		SetExpiredAt(expiredAt).SetMetadata(string(metadata)).SetNotifyUrl(input.NotifyURL).
		SetQrcodeUrl(providerResult.QRCodeURL).SetPayUrl(providerResult.PayURL).Save(ctx)
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("store payment order: %w", err)
	}
	s.scheduleMockSuccess(provider, created)
	return CreatePaymentResult{Success: true, PaymentID: created.ID, OutTradeNo: created.OutTradeNo, QRCodeURL: created.QrcodeUrl, PayURL: created.PayUrl}, nil
}

func (s *Service) CreateSubscriptionPayment(ctx context.Context, userID, planID string, method pay.PaymentMethod) (CreatePaymentResult, error) {
	plan, err := s.client.SubscriptionPlan.Get(ctx, planID)
	if ent.IsNotFound(err) {
		return CreatePaymentResult{}, fmt.Errorf("subscription plan %q: %w", planID, appRuntime.ErrNotFound)
	}
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("get subscription plan: %w", err)
	}
	if !plan.IsActive {
		return CreatePaymentResult{}, appRuntime.NewError(appRuntime.KindInvalidInput, "plan_inactive", "subscription plan is inactive", appRuntime.ErrConflict)
	}
	return s.createPayment(ctx, createPaymentInput{UserID: userID, Amount: plan.Price, Currency: "CNY", Method: method, PlanID: plan.ID, Metadata: map[string]json.RawMessage{"type": json.RawMessage(`"subscription"`), "planId": json.RawMessage(fmt.Sprintf("%q", plan.ID))}})
}

func (s *Service) CreateSubscriptionPaymentByMethod(ctx context.Context, userID, planID, method string) (CreatePaymentResult, error) {
	return s.CreateSubscriptionPayment(ctx, userID, planID, pay.PaymentMethod(method))
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
	if current.Method == string(pay.PaymentMethodMock) {
		return paymentView(current), nil
	}
	provider, _, err := s.resolveProvider(ctx, pay.PaymentMethod(current.Method))
	if err != nil {
		return PaymentView{}, err
	}
	status, err := provider.Query(ctx, outTradeNo)
	if err != nil {
		if errors.Is(err, pay.ErrProviderUnavailable) {
			return paymentView(current), nil
		}
		return PaymentView{}, fmt.Errorf("query payment provider: %w", err)
	}
	if status.Status == pay.PaymentStatePending {
		return paymentView(current), nil
	}
	updated, err := s.applyCallback(ctx, pay.PaymentMethod(current.Method), pay.ProviderCallback{OutTradeNo: outTradeNo, TransactionID: status.TransactionID, Amount: status.Amount, HasAmount: true, Currency: status.Currency, Status: status.Status, PaidAt: status.PaidAt, CallbackKey: string(current.Method) + ":query:" + outTradeNo})
	if err != nil {
		return PaymentView{}, err
	}
	return paymentView(updated), nil
}

func (s *Service) scheduleMockSuccess(provider pay.Provider, record *ent.Payment) {
	mock, ok := provider.(*pay.MockProvider)
	if !ok || !mock.AutoSuccess() || s.schedule == nil {
		return
	}
	s.schedule(mock.Delay(), func() {
		_, _ = s.applyCallback(context.Background(), pay.PaymentMethodMock, pay.ProviderCallback{
			OutTradeNo: record.OutTradeNo, TransactionID: "MOCK-" + record.OutTradeNo,
			Amount: record.Amount, HasAmount: true, Currency: record.Currency,
			Status: pay.PaymentStatePaid, PaidAt: s.clock.Now().UTC(),
			CallbackKey: "mock:" + record.OutTradeNo + ":auto-success",
		})
	})
}

func (s *Service) CancelPayment(ctx context.Context, outTradeNo string) error {
	current, err := s.client.Payment.Query().Where(payment.OutTradeNo(outTradeNo)).Only(ctx)
	if ent.IsNotFound(err) {
		return fmt.Errorf("payment %q: %w", outTradeNo, appRuntime.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("get payment before cancel: %w", err)
	}
	if current.Status == string(pay.PaymentStatePaid) {
		return fmt.Errorf("cancel paid payment: %w", pay.ErrInvalidTransition)
	}
	provider, _, err := s.resolveProvider(ctx, pay.PaymentMethod(current.Method))
	if err != nil {
		return err
	}
	if err := provider.Cancel(ctx, outTradeNo); err != nil {
		return fmt.Errorf("cancel payment provider: %w", err)
	}
	_, err = s.applyCallback(ctx, pay.PaymentMethod(current.Method), pay.ProviderCallback{OutTradeNo: outTradeNo, Status: pay.PaymentStateCancelled, CallbackKey: string(current.Method) + ":cancel:" + outTradeNo})
	return err
}

func (s *Service) HandleCallback(ctx context.Context, method pay.PaymentMethod, request pay.ProviderCallbackRequest) (PaymentView, error) {
	provider, _, err := s.resolveProvider(ctx, method)
	if err != nil {
		return PaymentView{}, err
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

func (s *Service) HandleRawCallback(ctx context.Context, method string, body []byte, headers, form map[string][]string) (PaymentView, error) {
	return s.HandleCallback(ctx, pay.PaymentMethod(method), pay.ProviderCallbackRequest{Body: body, Headers: headers, Form: form})
}

func (s *Service) AvailableMethods(ctx context.Context) ([]pay.PaymentMethod, error) {
	if s.providers != nil {
		methods := make([]pay.PaymentMethod, 0, len(s.providers))
		for _, method := range []pay.PaymentMethod{pay.PaymentMethodWeChat, pay.PaymentMethodAlipay, pay.PaymentMethodMock} {
			if _, ok := s.providers[method]; ok {
				methods = append(methods, method)
			}
		}
		return methods, nil
	}
	configuration, err := s.loadConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	methods := make([]pay.PaymentMethod, 0, 3)
	for _, method := range []pay.PaymentMethod{pay.PaymentMethodWeChat, pay.PaymentMethodAlipay, pay.PaymentMethodMock} {
		if method == pay.PaymentMethodWeChat && pay.WechatConfigured(configuration.WeChat) ||
			method == pay.PaymentMethodAlipay && pay.AlipayConfigured(configuration.Alipay) ||
			method == pay.PaymentMethodMock && configuration.Mock.Enabled {
			methods = append(methods, method)
		}
	}
	return methods, nil
}

func (s *Service) Settings(ctx context.Context) (PaymentSettings, error) {
	configuration, err := s.loadConfiguration(ctx)
	if err != nil {
		return PaymentSettings{}, err
	}
	return PaymentSettings{CreditPrice: configuration.CreditPrice, MinRecharge: configuration.MinRecharge, AlipayEnabled: configuration.AlipayEnabled, WeChatEnabled: configuration.WeChatEnabled, MockEnabled: configuration.Mock.Enabled}, nil
}

func paymentView(record *ent.Payment) PaymentView {
	view := PaymentView{ID: record.ID, UserID: record.UserId, OutTradeNo: record.OutTradeNo, TransactionID: record.TransactionId, Method: pay.PaymentMethod(record.Method), Amount: record.Amount, Currency: record.Currency, Status: pay.PaymentState(record.Status), QRCodeURL: record.QrcodeUrl, PayURL: record.PayUrl, CreatedAt: record.CreatedAt}
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

// CreateRechargePayment validates the recharge amount against the minimum and
// the credit conversion, then creates the payment order. Business rules for
// pricing live here; the handler only adapts the request.
func (s *Service) CreateRechargePayment(ctx context.Context, userID string, amount float64, credits int, method pay.PaymentMethod) (CreatePaymentResult, error) {
	settings, err := s.Settings(ctx)
	if err != nil {
		return CreatePaymentResult{}, err
	}
	if amount < settings.MinRecharge {
		return CreatePaymentResult{}, appRuntime.NewError(appRuntime.KindInvalidInput, "minimum_recharge", "recharge amount is below the minimum", nil)
	}
	expectedCredits := int(amount / settings.CreditPrice)
	if credits != expectedCredits {
		return CreatePaymentResult{}, appRuntime.NewError(appRuntime.KindInvalidInput, "invalid_credits", "credit calculation is invalid", nil)
	}
	metadata := map[string]json.RawMessage{
		"type":        json.RawMessage(`"recharge"`),
		"credits":     json.RawMessage(strconv.Itoa(credits)),
		"creditPrice": json.RawMessage(strconv.FormatFloat(settings.CreditPrice, 'f', 2, 64)),
	}
	return s.createPayment(ctx, createPaymentInput{UserID: userID, Amount: amount, Currency: "CNY", Method: method, Metadata: metadata})
}

func (s *Service) CreateRechargePaymentByMethod(ctx context.Context, userID string, amount float64, credits int, method string) (CreatePaymentResult, error) {
	return s.CreateRechargePayment(ctx, userID, amount, credits, pay.PaymentMethod(method))
}
