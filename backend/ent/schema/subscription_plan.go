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

// SubscriptionPlan holds the schema definition for the SubscriptionPlan entity.
type SubscriptionPlan struct {
	ent.Schema
}

func (SubscriptionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "SubscriptionPlan"}}
}

func (SubscriptionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("title").StorageKey("title"),
		field.Float("price").StorageKey("price"),
		field.Int("totalCredits").StorageKey("totalCredits"),
		field.Int("sortOrder").StorageKey("sortOrder").Default(0),
		field.Int("validityDuration").StorageKey("validityDuration"),
		field.String("validityUnit").StorageKey("validityUnit").Default("day"),
		field.String("creditResetCycle").StorageKey("creditResetCycle").Default("month"),
		field.Bool("isActive").StorageKey("isActive").Default(true),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (SubscriptionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("subscriptions", Subscription.Type).
			StorageKey(edge.Symbol("Subscription_planId_fkey")).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (SubscriptionPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title").Unique().StorageKey("SubscriptionPlan_title_key"),
		index.Fields("sortOrder").StorageKey("SubscriptionPlan_sortOrder_idx"),
		index.Fields("isActive").StorageKey("SubscriptionPlan_isActive_idx"),
	}
}
