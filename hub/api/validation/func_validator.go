// Package validation - Function-based validator wrapper
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package validation

// FuncValidator wraps a validation function to implement the Validator interface
type FuncValidator struct {
	ValidateFunc func(data map[string]interface{}) error
}

// Validate implements the Validator interface
func (f *FuncValidator) Validate(data map[string]interface{}) error {
	if f.ValidateFunc == nil {
		return nil // No validation function provided
	}
	return f.ValidateFunc(data)
}
