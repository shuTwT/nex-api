package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Advertisement holds the schema definition for the Advertisement entity.
type Advertisement struct {
	ent.Schema
}

func (Advertisement) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Advertisement"}}
}

func (Advertisement) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("image").StorageKey("image"),
		field.Int("imageWidth").StorageKey("imageWidth"),
		field.Int("imageHeight").StorageKey("imageHeight"),
		field.String("link").StorageKey("link"),
		field.String("title").StorageKey("title"),
		field.String("position").StorageKey("position"),
		field.Bool("isActive").StorageKey("isActive").Default(true),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").UpdateDefault(time.Now),
	}
}

func (Advertisement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("position").Unique().StorageKey("Advertisement_position_key"),
		index.Fields("position").StorageKey("Advertisement_position_idx"),
		index.Fields("isActive").StorageKey("Advertisement_isActive_idx"),
	}
}
