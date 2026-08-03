package catalog

import (
	"regexp"
	"strings"

	"github.com/shuTwT/nex-api/ent"
)

var aliasPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)

var validMethods = map[string]struct{}{
	"DELETE": {}, "GET": {}, "HEAD": {}, "OPTIONS": {}, "PATCH": {}, "POST": {}, "PUT": {},
}

func validateAPIInput(input APIInput) error {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Alias) == "" || strings.TrimSpace(input.Endpoint) == "" || strings.TrimSpace(input.CategoryID) == "" {
		return ValidationError("request", "name, alias, endpoint, and categoryId are required")
	}
	if err := validateAlias(input.Alias); err != nil {
		return err
	}
	if err := validateMethod(input.Method); err != nil {
		return err
	}
	if input.Pricing < 0 {
		return ValidationError("pricing", "must be non-negative")
	}
	return nil
}

func validateAPIUpdate(input APIUpdateInput) error {
	if input.Alias != nil {
		if err := validateAlias(*input.Alias); err != nil {
			return err
		}
	}
	if input.Method != nil {
		if err := validateMethod(*input.Method); err != nil {
			return err
		}
	}
	if input.Pricing != nil && *input.Pricing < 0 {
		return ValidationError("pricing", "must be non-negative")
	}
	return nil
}

func validateAlias(value string) error {
	if !aliasPattern.MatchString(value) {
		return ValidationError("alias", "must start with a letter and contain only letters and numbers")
	}
	return nil
}

func validateMethod(value string) error {
	method := strings.ToUpper(strings.TrimSpace(value))
	if _, ok := validMethods[method]; !ok {
		return ValidationError("method", "must be a valid HTTP method")
	}
	return nil
}

func applyAPIUpdate(builder *ent.APIUpdateOne, input APIUpdateInput) {
	if input.Name != nil {
		builder.SetName(*input.Name)
	}
	if input.Alias != nil {
		builder.SetAlias(*input.Alias)
	}
	if input.Description != nil {
		builder.SetDescription(*input.Description)
	}
	if input.Endpoint != nil {
		builder.SetEndpoint(*input.Endpoint)
	}
	if input.Method != nil {
		builder.SetMethod(strings.ToUpper(strings.TrimSpace(*input.Method)))
	}
	if input.CategoryID != nil {
		builder.SetCategoryId(*input.CategoryID)
	}
	if input.Pricing != nil {
		builder.SetPricing(*input.Pricing)
	}
	if input.Documentation != nil {
		builder.SetDocumentation(*input.Documentation)
	}
	if input.PreScript != nil {
		builder.SetPreScript(*input.PreScript)
	}
	if input.PostScript != nil {
		builder.SetPostScript(*input.PostScript)
	}
	if input.IsActive != nil {
		builder.SetIsActive(*input.IsActive)
	}
}
