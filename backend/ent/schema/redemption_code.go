package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RedemptionCode holds the schema definition for the RedemptionCode entity.
type RedemptionCode struct {
	ent.Schema
}

func (RedemptionCode) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "RedemptionCode"}}
}

func (RedemptionCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("code").StorageKey("code"),
		field.String("type").StorageKey("type"),
		field.String("planId").StorageKey("planId").Optional(),
		field.String("planName").StorageKey("planName").Optional(),
		field.Int("credits").StorageKey("credits").Optional(),
		field.Time("expiresAt").StorageKey("expiresAt").Optional(),
		field.Bool("isUsed").StorageKey("isUsed").Default(false),
		field.String("usedBy").StorageKey("usedBy").Optional(),
		field.Time("usedAt").StorageKey("usedAt").Optional(),
		field.String("createdBy").StorageKey("createdBy"),
		field.String("batchId").StorageKey("batchId").Optional(),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (RedemptionCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique().StorageKey("RedemptionCode_code_key"),
		index.Fields("code").StorageKey("RedemptionCode_code_idx"),
		index.Fields("type").StorageKey("RedemptionCode_type_idx"),
		index.Fields("isUsed").StorageKey("RedemptionCode_isUsed_idx"),
		index.Fields("batchId").StorageKey("RedemptionCode_batchId_idx"),
		index.Fields("createdBy").StorageKey("RedemptionCode_createdBy_idx"),
	}
}
