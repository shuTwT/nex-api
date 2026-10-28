// Package constant holds small shared definitions that carry no behavior and
// depend on nothing else in the module, so any layer can import them.
package constant

// FieldError describes one invalid request field.
type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}
