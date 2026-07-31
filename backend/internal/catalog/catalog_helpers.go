package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/api"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
)

type dependentCounts struct {
	parameters int
	responses  int
	usage      int
}

func (c dependentCounts) total() int { return c.parameters + c.responses + c.usage }

func (c dependentCounts) describe() string {
	parts := make([]string, 0, 3)
	if c.parameters > 0 {
		parts = append(parts, fmt.Sprintf("%d parameters", c.parameters))
	}
	if c.responses > 0 {
		parts = append(parts, fmt.Sprintf("%d responses", c.responses))
	}
	if c.usage > 0 {
		parts = append(parts, fmt.Sprintf("%d usage records", c.usage))
	}
	return strings.Join(parts, ", ")
}

func dependentAPICounts(ctx context.Context, tx *ent.Tx, id string) (dependentCounts, error) {
	parameters, err := tx.Api.Query().Where(api.ID(id)).QueryParameters().Count(ctx)
	if err != nil {
		return dependentCounts{}, err
	}
	responses, err := tx.Api.Query().Where(api.ID(id)).QueryResponses().Count(ctx)
	if err != nil {
		return dependentCounts{}, err
	}
	usage, err := tx.Api.Query().Where(api.ID(id)).QueryUsageRecords().Count(ctx)
	if err != nil {
		return dependentCounts{}, err
	}
	return dependentCounts{parameters: parameters, responses: responses, usage: usage}, nil
}

func classifyNotFound(err error, resource string) error {
	if ent.IsNotFound(err) {
		return runtime.NewAPIError(http.StatusNotFound, "not_found", resource+" not found", fmt.Errorf("%s: %w", resource, runtime.ErrNotFound))
	}
	return fmt.Errorf("query %s: %w", resource, err)
}

func validationError(field, reason string) error {
	return runtime.NewValidationError(runtime.FieldError{Field: field, Reason: reason})
}

func classifyMutationError(err error, conflictMessage string) error {
	switch {
	case ent.IsNotFound(err):
		return classifyNotFound(err, "catalog resource")
	case ent.IsConstraintError(err):
		return runtime.NewAPIError(http.StatusConflict, "catalog_conflict", conflictMessage, fmt.Errorf("%w: %v", runtime.ErrConflict, err))
	default:
		return fmt.Errorf("catalog mutation: %w", err)
	}
}

func dependencyConflict(resource, name, details string) error {
	return runtime.NewAPIError(http.StatusConflict, "dependent_records", fmt.Sprintf("cannot delete %s %q: dependent records exist (%s)", resource, name, details), runtime.ErrConflict)
}

func abortTx(tx *ent.Tx, err error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return errors.Join(err, fmt.Errorf("rollback catalog mutation: %w", rollbackErr))
	}
	return err
}

func writeAudit(ctx context.Context, tx *ent.Tx, action, resource, details, level, status string) error {
	builder := tx.AuditLog.Create().SetAction(action).SetResource(resource).SetDetails(details).SetLevel(level).SetStatus(status)
	if principal, ok := authz.PrincipalFromContext(ctx); ok && principal.UserID != "" {
		builder.SetUserID(principal.UserID)
	}
	if _, err := builder.Save(ctx); err != nil {
		return err
	}
	return nil
}

func normalizeListOptions(options APIListOptions) APIListOptions {
	if options.Page < 1 {
		options.Page = 1
	}
	if options.Limit < 1 {
		options.Limit = 10
	}
	if options.Limit > 100 {
		options.Limit = 100
	}
	return options
}
