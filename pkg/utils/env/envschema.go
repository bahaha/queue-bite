package env

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type ConfigError struct {
	errors []error
}

func (e *ConfigError) addError(err error) {
	if e.errors == nil {
		e.errors = make([]error, 0)
	}
	e.errors = append(e.errors, err)
}

func (e *ConfigError) Error() string {
	if e.errors == nil || len(e.errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e.errors {
		messages = append(messages, fmt.Sprintf("\t\t> %s", err.Error()))
	}

	return fmt.Sprintf("\tConfiguration Errors:\n%s", strings.Join(messages, "\n"))
}

func (e *ConfigError) hasErrors() bool {
	return len(e.errors) > 0
}

type configLoader struct {
	*ConfigError
	getenv func(string) string
}

func newConfigLoader(getenv func(string) string) *configLoader {
	return &configLoader{getenv: getenv, ConfigError: &ConfigError{errors: make([]error, 0)}}
}

func (l *configLoader) loadField(field reflect.Value, fieldType reflect.StructField) error {
	envKey := fieldType.Tag.Get("env")
	if envKey == "" {
		return nil
	}

	required, _ := strconv.ParseBool(fieldType.Tag.Get("required"))
	envVal := l.getenv(envKey)
	if envVal == "" {
		envVal = fieldType.Tag.Get("default")
	}

	if envVal == "" && required {
		return fmt.Errorf("missing required environment variable of key %s", envKey)
	}

	if envVal == "" {
		return nil
	}

	parser, ok := defaultRegistry.FindParser(field.Type())
	if !ok {
		return fmt.Errorf("unsupported type for %s: %s", envKey, field.Type())
	}

	if err := parser.Parse(envVal, field); err != nil {
		return fmt.Errorf("invalid value for %s: %w", envKey, err)
	}
	return nil
}

func LoadConfig[T any](getenv func(string) string, schema T) (T, error) {
	cfg := reflect.ValueOf(schema)
	if cfg.Kind() == reflect.Ptr {
		cfg = cfg.Elem()
	}

	if cfg.Kind() != reflect.Struct {
		return schema, fmt.Errorf("schema must be a struct or pointer to struct")
	}

	loader := newConfigLoader(getenv)
	typ := cfg.Type()
	for i := 0; i < cfg.NumField(); i++ {
		field := cfg.Field(i)
		if !field.CanSet() {
			continue
		}

		if err := loader.loadField(field, typ.Field(i)); err != nil {
			loader.addError(err)
		}
	}

	if loader.hasErrors() {
		return schema, fmt.Errorf("%s", loader.Error())
	}

	return schema, nil
}
