package membership

import (
	"fmt"
	"strings"

	"github.com/shuTwT/nex-api/ent"
	appRuntime "github.com/shuTwT/nex-api/internal/service/apierror"
)

var validityUnits = map[string]bool{"day": true, "week": true, "month": true, "year": true}

func normalizeCreateInput(input PlanCreateInput) (PlanCreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.ValidityUnit = strings.ToLower(strings.TrimSpace(input.ValidityUnit))
	input.CreditResetCycle = strings.ToLower(strings.TrimSpace(input.CreditResetCycle))
	if input.CreditResetCycle == "" {
		input.CreditResetCycle = "month"
	}
	if input.ValidityUnit == "" {
		input.ValidityUnit = "day"
	}
	if err := validatePlanFields(input.Title, input.Price, input.TotalCredits, input.ValidityDuration, input.ValidityUnit); err != nil {
		return PlanCreateInput{}, err
	}
	return input, nil
}

func validateUpdateInput(input *PlanUpdateInput) error {
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "title", Reason: "required"})
	}
	if input.Price != nil && *input.Price < 0 {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "price", Reason: "must be non-negative"})
	}
	if input.TotalCredits != nil && *input.TotalCredits < 0 {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "totalCredits", Reason: "must be non-negative"})
	}
	if input.ValidityDuration != nil && *input.ValidityDuration < 1 {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "validityDuration", Reason: "must be positive"})
	}
	if input.ValidityUnit != nil && !validityUnits[strings.ToLower(strings.TrimSpace(*input.ValidityUnit))] {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "validityUnit", Reason: "must be day, week, month, or year"})
	}
	return nil
}

func validatePlanFields(title string, price float64, credits, duration int, unit string) error {
	fields := make([]appRuntime.FieldError, 0, 2)
	if title == "" {
		fields = append(fields, appRuntime.FieldError{Field: "title", Reason: "required"})
	}
	if price < 0 {
		fields = append(fields, appRuntime.FieldError{Field: "price", Reason: "must be non-negative"})
	}
	if credits < 0 {
		fields = append(fields, appRuntime.FieldError{Field: "totalCredits", Reason: "must be non-negative"})
	}
	if duration < 1 {
		fields = append(fields, appRuntime.FieldError{Field: "validityDuration", Reason: "must be positive"})
	}
	if !validityUnits[unit] {
		fields = append(fields, appRuntime.FieldError{Field: "validityUnit", Reason: "must be day, week, month, or year"})
	}
	if len(fields) > 0 {
		return appRuntime.NewValidationError(fields...)
	}
	return nil
}

func wrapPlanWriteError(action string, err error) error {
	switch {
	case ent.IsNotFound(err):
		return fmt.Errorf("%s: %w", action, appRuntime.ErrNotFound)
	case ent.IsConstraintError(err):
		return fmt.Errorf("%s: %w", action, appRuntime.ErrConflict)
	default:
		return fmt.Errorf("%s: %w", action, err)
	}
}

func normalizePageFilter(filter PlanListFilter) PlanListFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 10
	}
	return filter
}
