package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Payment holds the schema definition for the Payment entity.
type Payment struct {
	ent.Schema
}

func (Payment) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Payment"}}
}

func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("userId").StorageKey("userId"),
		field.String("outTradeNo").StorageKey("outTradeNo"),
		field.String("transactionId").StorageKey("transactionId").Optional(),
		field.String("method").StorageKey("method"),
		field.Float("amount").StorageKey("amount"),
		field.String("currency").StorageKey("currency").Default("CNY"),
		field.String("status").StorageKey("status").Default("pending"),
		field.String("qrcodeUrl").StorageKey("qrcodeUrl").Optional(),
		field.String("payUrl").StorageKey("payUrl").Optional(),
		field.String("notifyUrl").StorageKey("notifyUrl").Optional(),
		field.Time("paidAt").StorageKey("paidAt").Optional(),
		field.Time("expiredAt").StorageKey("expiredAt").Optional(),
		field.Time("cancelledAt").StorageKey("cancelledAt").Optional(),
		field.String("metadata").StorageKey("metadata").Optional(),
		field.Int("callbackVersion").StorageKey("callbackVersion").Default(0),
		field.String("callbackKey").StorageKey("callbackKey").Optional().Unique(),
		field.Time("callbackProcessedAt").StorageKey("callbackProcessedAt").Optional(),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Payment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("payments").
			Field("userId").
			Unique().
			Required(),
		edge.To("subscription", Subscription.Type).
			Unique().
			StorageKey(edge.Symbol("Subscription_paymentId_fkey")).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("outTradeNo").Unique().StorageKey("Payment_outTradeNo_key"),
		index.Fields("userId").StorageKey("Payment_userId_idx"),
		index.Fields("status").StorageKey("Payment_status_idx"),
		index.Fields("method").StorageKey("Payment_method_idx"),
		index.Fields("outTradeNo").StorageKey("Payment_outTradeNo_idx"),
		index.Fields("transactionId").StorageKey("Payment_transactionId_idx"),
		index.Fields("callbackVersion").StorageKey("Payment_callbackVersion_idx"),
		index.Fields("createdAt").StorageKey("Payment_createdAt_idx"),
	}
}
