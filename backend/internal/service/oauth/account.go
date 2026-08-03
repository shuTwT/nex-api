package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/account"
	"github.com/shuTwT/nex-api/backend/ent/user"
	infraoauth "github.com/shuTwT/nex-api/backend/internal/infra/oauth"
)

func (s *Service) Provision(ctx context.Context, profile infraoauth.NormalizedProfile, tokens infraoauth.AccountTokens) (*ent.User, error) {
	for attempt := 0; attempt < 3; attempt++ {
		accountUser, err := s.provisionOnce(ctx, profile, tokens)
		if err == nil {
			return accountUser, nil
		}
		if !ent.IsConstraintError(err) {
			return nil, err
		}
		if existing, lookupErr := s.client.Account.Query().Where(
			account.Provider(profile.Provider),
			account.ProviderAccountId(profile.ProviderAccountID),
		).Only(ctx); lookupErr == nil {
			accountUser, lookupErr = existing.QueryUser().Only(ctx)
			if lookupErr == nil {
				return accountUser, nil
			}
		}
	}
	return nil, errors.New("oauth: account provisioning conflicted repeatedly")
}

func (s *Service) provisionOnce(ctx context.Context, profile infraoauth.NormalizedProfile, tokens infraoauth.AccountTokens) (*ent.User, error) {
	if err := validateProfile(profile); err != nil {
		return nil, err
	}
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("oauth: begin account transaction: %w", err)
	}
	rollback := true
	defer func() {
		if rollback {
			_ = tx.Rollback()
		}
	}()
	linked, err := tx.Account.Query().Where(
		account.Provider(profile.Provider),
		account.ProviderAccountId(profile.ProviderAccountID),
	).Only(ctx)
	if err == nil {
		accountUser, queryErr := linked.QueryUser().Only(ctx)
		if queryErr != nil {
			return nil, fmt.Errorf("oauth: query linked user: %w", queryErr)
		}
		if _, updateErr := updateAccount(ctx, linked, tokens, s.clock()); updateErr != nil {
			return nil, updateErr
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("oauth: commit linked account: %w", err)
		}
		rollback = false
		return accountUser, nil
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("oauth: query linked account: %w", err)
	}
	accountUser, err := tx.User.Query().Where(user.Email(profile.Email)).Only(ctx)
	if ent.IsNotFound(err) {
		username, usernameErr := deterministicUsername(ctx, tx, profile)
		if usernameErr != nil {
			return nil, usernameErr
		}
		dummyPassword, passwordErr := infraoauth.RandomToken(32)
		if passwordErr != nil {
			return nil, fmt.Errorf("oauth: generate password placeholder: %w", passwordErr)
		}
		accountUser, err = tx.User.Create().
			SetNillableName(nonEmpty(profile.Name)).
			SetEmail(profile.Email).
			SetNillableImage(nonEmpty(profile.Image)).
			SetUsername(username).
			SetPassword("oauth:" + dummyPassword).
			SetRole("user").
			SetCredits(1000).
			SetCreatedAt(s.clock().UTC()).
			SetUpdatedAt(s.clock().UTC()).
			Save(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("oauth: provision user: %w", err)
	}
	createdAccount := tx.Account.Create()
	createdAccount.SetUserID(accountUser.ID)
	createdAccount.SetProvider(profile.Provider)
	createdAccount.SetProviderAccountId(profile.ProviderAccountID)
	setAccountTokenFields(createdAccount, tokens)
	createdAccount.SetCreatedAt(s.clock().UTC()).SetUpdatedAt(s.clock().UTC())
	if _, err := createdAccount.Save(ctx); err != nil {
		return nil, fmt.Errorf("oauth: link account: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("oauth: commit account provisioning: %w", err)
	}
	rollback = false
	return accountUser, nil
}

func updateAccount(ctx context.Context, existing *ent.Account, tokens infraoauth.AccountTokens, now time.Time) (*ent.Account, error) {
	builder := existing.Update().SetUpdatedAt(now.UTC())
	setAccountUpdateFields(builder, tokens)
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("oauth: update account tokens: %w", err)
	}
	return updated, nil
}

func setAccountTokenFields(builder *ent.AccountCreate, tokens infraoauth.AccountTokens) {
	if tokens.AccessToken != "" {
		builder.SetAccessToken(tokens.AccessToken)
	}
	if tokens.RefreshToken != "" {
		builder.SetRefreshToken(tokens.RefreshToken)
	}
	if tokens.TokenType != "" {
		builder.SetTokenType(tokens.TokenType)
	}
	if tokens.Scope != "" {
		builder.SetScope(tokens.Scope)
	}
	if tokens.ExpiresAt > 0 {
		builder.SetExpiresAt(tokens.ExpiresAt)
	}
	if tokens.IDToken != "" {
		builder.SetIdToken(tokens.IDToken)
	}
}

func setAccountUpdateFields(builder *ent.AccountUpdateOne, tokens infraoauth.AccountTokens) {
	if tokens.AccessToken != "" {
		builder.SetAccessToken(tokens.AccessToken)
	}
	if tokens.RefreshToken != "" {
		builder.SetRefreshToken(tokens.RefreshToken)
	}
	if tokens.TokenType != "" {
		builder.SetTokenType(tokens.TokenType)
	}
	if tokens.Scope != "" {
		builder.SetScope(tokens.Scope)
	}
	if tokens.ExpiresAt > 0 {
		builder.SetExpiresAt(tokens.ExpiresAt)
	}
	if tokens.IDToken != "" {
		builder.SetIdToken(tokens.IDToken)
	}
}

func validateProfile(profile infraoauth.NormalizedProfile) error {
	parsed, err := mail.ParseAddress(profile.Email)
	if err != nil || parsed.Address != profile.Email || profile.Provider == "" || profile.ProviderAccountID == "" {
		return errors.New("oauth: invalid provider profile")
	}
	return nil
}

func deterministicUsername(ctx context.Context, client *ent.Tx, profile infraoauth.NormalizedProfile) (string, error) {
	base := profile.Username
	if base == "" {
		base = profile.Name
	}
	if base == "" {
		base = strings.Split(profile.Email, "@")[0]
	}
	base = slug(base)
	if base == "" {
		base = "user"
	}
	if _, err := client.User.Query().Where(user.Username(base)).Only(ctx); ent.IsNotFound(err) {
		return base, nil
	} else if err != nil {
		return "", fmt.Errorf("oauth: check username: %w", err)
	}
	digest := sha256.Sum256([]byte(strings.ToLower(profile.Email)))
	suffix := hex.EncodeToString(digest[:])[:8]
	return base + "-" + suffix, nil
}

func slug(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '-' || char == '_' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func nonEmpty(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	return &trimmed
}
