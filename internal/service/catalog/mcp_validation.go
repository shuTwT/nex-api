package catalog

import (
	"regexp"
	"strings"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*$`)

var validMCPTypes = map[string]struct{}{
	"stdio": {}, "sse": {}, "streamableHttp": {},
}

func validateMCPInput(input MCPInput) error {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Identifier) == "" || strings.TrimSpace(input.Type) == "" {
		return ValidationError("request", "name, identifier, and type are required")
	}
	if err := validateIdentifier(input.Identifier); err != nil {
		return err
	}
	if err := validateMCPType(input.Type); err != nil {
		return err
	}
	if input.Pricing < 0 {
		return ValidationError("pricing", "must be non-negative")
	}
	return nil
}

func validateMCPUpdate(input MCPUpdateInput) error {
	if input.Identifier != nil {
		if err := validateIdentifier(*input.Identifier); err != nil {
			return err
		}
	}
	if input.Type != nil {
		if err := validateMCPType(*input.Type); err != nil {
			return err
		}
	}
	if input.Pricing != nil && *input.Pricing < 0 {
		return ValidationError("pricing", "must be non-negative")
	}
	return nil
}

func validateIdentifier(value string) error {
	if !identifierPattern.MatchString(value) {
		return ValidationError("identifier", "must start with a letter and contain only letters, numbers, and hyphens")
	}
	return nil
}

func validateMCPType(value string) error {
	if _, ok := validMCPTypes[value]; !ok {
		return ValidationError("type", "must be stdio, sse, or streamableHttp")
	}
	return nil
}

func normalizeMCPListOptions(options MCPListOptions) MCPListOptions {
	options.Type = strings.TrimSpace(options.Type)
	options.Search = strings.TrimSpace(options.Search)
	options.Status = strings.ToLower(strings.TrimSpace(options.Status))
	if strings.EqualFold(options.Type, "all") {
		options.Type = ""
	}
	if options.Status == "all" {
		options.Status = ""
	}
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
