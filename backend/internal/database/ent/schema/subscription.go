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

// Subscription holds the schema definition for the Subscription entity.
type Subscription struct {
	ent.Schema
}

func (Subscription) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Subscription"}}
}

func (Subscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("userId").StorageKey("userId"),
		field.String("planId").StorageKey("planId").Optional(),
		field.String("planName").StorageKey("planName"),
		field.Int("credits").StorageKey("credits"),
		field.Float("price").StorageKey("price"),
		field.Time("startDate").StorageKey("startDate").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("endDate").StorageKey("endDate"),
		field.Bool("isActive").StorageKey("isActive").Default(true),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").UpdateDefault(time.Now),
		field.String("paymentId").StorageKey("paymentId").Optional(),
	}
}

func (Subscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("subscriptions").
			Field("userId").
			Unique().
			Required(),
		edge.From("plan", SubscriptionPlan.Type).
			Ref("subscriptions").
			Field("planId").
			Unique(),
		edge.From("payment", Payment.Type).
			Ref("subscription").
			Field("paymentId").
			Unique(),
	}
}

func (Subscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paymentId").Unique().StorageKey("Subscription_paymentId_key"),
		index.Fields("userId").StorageKey("Subscription_userId_idx"),
		index.Fields("planId").StorageKey("Subscription_planId_idx"),
		index.Fields("isActive").StorageKey("Subscription_isActive_idx"),
		index.Fields("paymentId").StorageKey("Subscription_paymentId_idx"),
	}
}
