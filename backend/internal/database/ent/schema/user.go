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

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "User"}}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("name").StorageKey("name").Optional(),
		field.String("email").StorageKey("email"),
		field.Time("emailVerified").StorageKey("emailVerified").Optional(),
		field.String("image").StorageKey("image").Optional(),
		field.String("username").StorageKey("username"),
		field.String("password").StorageKey("password"),
		field.String("role").StorageKey("role").Default("user"),
		field.Int("credits").StorageKey("credits").Default(1000),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").UpdateDefault(time.Now),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique().StorageKey("User_email_key"),
		index.Fields("username").Unique().StorageKey("User_username_key"),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("subscriptions", Subscription.Type).
			StorageKey(edge.Symbol("Subscription_userId_fkey")).
			Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.To("apiUsage", ApiUsage.Type).
			StorageKey(edge.Symbol("ApiUsage_userId_fkey")).
			Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.To("accounts", Account.Type).
			StorageKey(edge.Symbol("Account_userId_fkey")).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("sessions", Session.Type).
			StorageKey(edge.Symbol("Session_userId_fkey")).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("apiTokens", ApiToken.Type).
			StorageKey(edge.Symbol("ApiToken_userId_fkey")).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("auditLogs", AuditLog.Type).
			StorageKey(edge.Symbol("AuditLog_userId_fkey")).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("payments", Payment.Type).
			StorageKey(edge.Symbol("Payment_userId_fkey")).
			Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.To("mcpUsage", McpUsage.Type).
			StorageKey(edge.Symbol("McpUsage_userId_fkey")).
			Annotations(entsql.OnDelete(entsql.Restrict)),
	}
}
