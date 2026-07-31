package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SystemSetting holds the schema definition for the SystemSetting entity.
type SystemSetting struct {
	ent.Schema
}

func (SystemSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "SystemSetting"}}
}

func (SystemSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("key").StorageKey("key"),
		field.String("value").StorageKey("value"),
		field.String("category").StorageKey("category").Default("general"),
		field.String("description").StorageKey("description").Optional(),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").UpdateDefault(time.Now),
	}
}

func (SystemSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key").Unique().StorageKey("SystemSetting_key_key"),
		index.Fields("key").StorageKey("SystemSetting_key_idx"),
		index.Fields("category").StorageKey("SystemSetting_category_idx"),
	}
}
