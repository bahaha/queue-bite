package env

import (
	"fmt"
	"strings"
)

// ConfigError represents a collection of configuration errors that occurred
// during loading or validation. It groups errors by field for better error reporting.
type ConfigError struct {
	errors []error
	fields map[string][]error
}

// RequiredFieldError indicates that a required environment variable was not set.
// This error is returned when a field marked as required:"T" has no value and no default.
type RequiredFieldError struct {
	Field string
}

func (e *RequiredFieldError) Error() string {
	return fmt.Sprintf("requried field %q is not set", e.Field)
}

// PraseError indicates a failure to convert an environment variable string
// to the target Go type. It includes the field name, the invalid value,
// and the underlying type conversion error.
type ParseError struct {
	Field     string
	Value     any
	TypeError error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse value %q for field %q: %v", e.Value, e.Field, e.TypeError)
}

func (e *ConfigError) AddError(err error) {
	if e.errors == nil {
		e.errors = make([]error, 0)
	}
	e.errors = append(e.errors, err)
}

func (e *ConfigError) AddFieldError(field string, err error) {
	if e.fields == nil {
		e.fields = make(map[string][]error)
	}
	e.fields[field] = append(e.fields[field], err)
	e.AddError(err)
}

// ConfigError#Error returns a formatted string containing all configuration errors.
//
// Example output:
//
// configuration errors:
//
//	PORT:
//	    - failed to parse value "invalid" for field "PORT": invalid integer
//	DATABASE_URL:
//	    - required field "DATABASE_URL" is not set
func (e *ConfigError) Error() string {
	if len(e.errors) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("configuration errors:")
	for field, errs := range e.fields {
		b.WriteString(fmt.Sprintf("\n\t%s:", field))
		for _, err := range errs {
			b.WriteString(fmt.Sprintf("\n\t\t- %v", err))
		}
	}
	return b.String()
}
