package payment

import (
	"encoding/json"

	pay "github.com/shuTwT/nex-api/internal/infra/pay"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type createPaymentInput struct {
	UserID    string
	Amount    float64
	Currency  string
	Method    pay.PaymentMethod
	PlanID    string
	NotifyURL string
	Metadata  map[string]json.RawMessage
}

type PaymentView = model.PaymentResp
type CreatePaymentResult = model.PaymentCreateResp
type PaymentSettings = model.PaymentSettingsResp
