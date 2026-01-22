// Package validation - Unit tests for validators
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package validation

import (
	"regexp"
	"testing"
)

func TestStringValidator_Required(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "required field present",
			field:   "name",
			data:    map[string]interface{}{"name": "test"},
			wantErr: false,
		},
		{
			name:    "required field missing",
			field:   "name",
			data:    map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "required field nil",
			field:   "name",
			data:    map[string]interface{}{"name": nil},
			wantErr: true,
		},
		{
			name:    "required field empty string",
			field:   "name",
			data:    map[string]interface{}{"name": ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &StringValidator{
				Field:    tt.field,
				Required:  true,
				AllowEmpty: false,
			}
			err := v.Validate(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Length(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		minLen   int
		maxLen   int
		wantErr  bool
	}{
		{
			name:    "valid length",
			field:   "name",
			value:   "test",
			minLen:  1,
			maxLen:  10,
			wantErr: false,
		},
		{
			name:    "too short",
			field:   "name",
			value:   "t",
			minLen:  2,
			maxLen:  10,
			wantErr: true,
		},
		{
			name:    "too long",
			field:   "name",
			value:   "this is too long",
			minLen:  1,
			maxLen:  10,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &StringValidator{
				Field:     tt.field,
				MinLength: tt.minLen,
				MaxLength: tt.maxLen,
			}
			err := v.Validate(map[string]interface{}{tt.field: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("StringValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Pattern(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		pattern *regexp.Regexp
		wantErr bool
	}{
		{
			name:    "matches pattern",
			field:   "code",
			value:   "ABC123",
			pattern: regexp.MustCompile(`^[A-Z0-9]+$`),
			wantErr: false,
		},
		{
			name:    "does not match pattern",
			field:   "code",
			value:   "abc123",
			pattern: regexp.MustCompile(`^[A-Z0-9]+$`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &StringValidator{
				Field:   tt.field,
				Pattern: tt.pattern,
			}
			err := v.Validate(map[string]interface{}{tt.field: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("StringValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Enum(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		enum    []string
		wantErr bool
	}{
		{
			name:    "valid enum value",
			field:   "status",
			value:   "pending",
			enum:    []string{"pending", "completed", "archived"},
			wantErr: false,
		},
		{
			name:    "invalid enum value",
			field:   "status",
			value:   "invalid",
			enum:    []string{"pending", "completed", "archived"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &StringValidator{
				Field: tt.field,
				Enum:  tt.enum,
			}
			err := v.Validate(map[string]interface{}{tt.field: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("StringValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumericValidator(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    interface{}
		min      *float64
		max      *float64
		wantErr  bool
	}{
		{
			name:    "valid number",
			field:   "age",
			value:   25,
			min:     floatPtr(0),
			max:     floatPtr(120),
			wantErr: false,
		},
		{
			name:    "below minimum",
			field:   "age",
			value:   -5,
			min:     floatPtr(0),
			max:     floatPtr(120),
			wantErr: true,
		},
		{
			name:    "above maximum",
			field:   "age",
			value:   150,
			min:     floatPtr(0),
			max:     floatPtr(120),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &NumericValidator{
				Field: tt.field,
				Min:   tt.min,
				Max:   tt.max,
				IsInt: true,
			}
			err := v.Validate(map[string]interface{}{tt.field: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("NumericValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailValidator(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"invalid email", "notanemail", true},
		{"invalid format", "@example.com", true},
		{"missing domain", "test@", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := EmailValidator("email", true)
			err := v.Validate(map[string]interface{}{"email": tt.email})
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDValidator(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid UUID", "not-a-uuid", true},
		{"wrong format", "550e8400e29b41d4a716446655440000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := UUIDValidator("id", true)
			err := v.Validate(map[string]interface{}{"id": tt.uuid})
			if (err != nil) != tt.wantErr {
				t.Errorf("UUIDValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompositeValidator(t *testing.T) {
	v := &CompositeValidator{
		Validators: []Validator{
			&StringValidator{
				Field:    "name",
				Required: true,
				MinLength: 1,
			},
			&StringValidator{
				Field:    "email",
				Required: true,
			},
		},
	}

	t.Run("all valid", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "Test",
			"email": "test@example.com",
		}
		if err := v.Validate(data); err != nil {
			t.Errorf("CompositeValidator.Validate() error = %v", err)
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "Test",
		}
		if err := v.Validate(data); err == nil {
			t.Error("CompositeValidator.Validate() expected error, got nil")
		}
	})
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal string", "hello world", "hello world"},
		{"with null bytes", "hello\x00world", "helloworld"},
		{"with control chars", "hello\nworld\t", "hello\nworld"},
		{"with whitespace", "  hello  ", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateNoSQLInjection(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"safe input", "normal text", false},
		{"SQL injection attempt", "'; DROP TABLE users; --", true},
		{"SELECT attempt", "SELECT * FROM users", true},
		{"UNION attempt", "UNION SELECT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNoSQLInjection("field", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNoSQLInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

