package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

// NewEnvLoader creates a new configuration loader that populates values from environment variables.
// It takes options to override the default configuration.
//
// - source: the environment variable source of default `os.Getenv`
// - registry: built-in parser registry with string, bool, int, float, duration, ...etc.
//
// The struct fields should be tagged with `env` to specify the
// environment variable names.
//
// The configuration struct supports the following field tags:
//   - env: the environment variable name (required for fields to be populated)
//   - required: if "T", returns an error when the environment variable is not set
//   - default: fallback value if the environment variable is not set
//
// Example:
//
//	type Config struct {
//	    Port     int           `env:"PORT" default:"8080"`
//	    Host     string        `env:"HOST" required:"T"`
//	    LogLevel string        `env:"LOG_LEVEL" default:"info"`
//	    Timeout  time.Duration `env:"TIMEOUT" default:"30s"`
//	}
//
//	cfg := &Config{}
//	loader := NewEnvLoader(WithEnvironmentSource(getenv))
//	if err := loader.Parse(cfg); err != nil {
//	    log.Fatal(err)
//	}
func NewEnvLoader(opts ...Option) *Loader {
	l := &Loader{
		source:   os.Getenv,
		registry: NewBuiltinRegistry(),
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

func WithEnvSource(source func(key string) string) Option {
	return func(l *Loader) {
		l.source = source
	}
}

// Parse loads environment variables into the configuration struct.
// It returns an error if:
//   - The config parameter passed to NewLoader is not a pointer to a struct
//   - A required field is not set in the environment
//   - A field value cannot be parsed into the target type
//   - The environment variable value is invalid according to field constraints
//
// If multiple errors occur during parsing, they are collected into a single ConfigError
// that contains details about all failures.
func (l *Loader) Parse(config interface{}) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("config must be a non-nil pointer to a struct")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	cfgErr := &ConfigError{}
	fields := l.getFields(v.Type())

	for _, field := range fields {
		value := v.FieldByName(field.Name)
		if !value.CanSet() {
			continue
		}

		if err := l.loadField(value, field); err != nil {
			cfgErr.AddFieldError(field.Name, err)
		}
	}

	if len(cfgErr.errors) > 0 {
		return cfgErr
	}
	return nil
}

func (l *Loader) getFields(t reflect.Type) []Field {
	var fields []Field
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		tag := sf.Tag.Get("env")
		if tag == "" {
			continue
		}

		required, _ := strconv.ParseBool(sf.Tag.Get("required"))

		fields = append(fields, Field{
			Name:       sf.Name,
			EnvKey:     tag,
			Type:       sf.Type,
			Required:   required,
			DefaultVal: sf.Tag.Get("default"),
		})
	}
	return fields
}

func (l *Loader) loadField(value reflect.Value, field Field) error {
	envKey := field.EnvKey
	envVal := l.source(envKey)

	if envVal == "" {
		if field.Required && field.DefaultVal == "" {
			return &RequiredFieldError{Field: envKey}
		}

		if field.DefaultVal != "" {
			envVal = field.DefaultVal
		} else {
			return nil
		}
	}

	parser, ok := l.registry.FindParser(field.Type)
	if !ok {
		return fmt.Errorf("no parser registered for type: %v", field.Type)
	}

	parsed, err := parser.Parse(envVal)
	if err != nil {
		return &ParseError{
			Field:     envKey,
			Value:     parsed,
			TypeError: err,
		}
	}

	value.Set(reflect.ValueOf(parsed))
	return nil
}
