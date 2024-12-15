package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ConfigError struct {
	errors []error
}

func (e *ConfigError) addError(err error) {
	e.errors = append(e.errors, err)
}

func (e *ConfigError) Error() string {
	if len(e.errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e.errors {
		messages = append(messages, fmt.Sprintf("\t\t> %s", err.Error()))
	}

	return fmt.Sprintf("\tConfiguration Errors:\n%s", strings.Join(messages, "\n"))
}

func LoadConfig[T any](getenv func(string) string, schema T) (T, error) {
	cfg := reflect.ValueOf(schema)
	if cfg.Kind() == reflect.Ptr {
		cfg = cfg.Elem()
	}

	if cfg.Kind() != reflect.Struct {
		return schema, fmt.Errorf("schema must be a struct or pointer to struct")
	}

	var badCfg ConfigError
	typ := cfg.Type()
	for i := 0; i < cfg.NumField(); i++ {
		field := cfg.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		envKey := fieldType.Tag.Get("env")
		if envKey == "" {
			continue
		}

		required, _ := strconv.ParseBool(fieldType.Tag.Get("required"))
		envVal := getenv(envKey)
		if envVal == "" {
			envVal = fieldType.Tag.Get("default")
		}

		if envVal == "" && required {
			badCfg.addError(fmt.Errorf("missing required environment variable of key %s", envKey))
			continue
		}

		if envVal == "" {
			continue
		}

		if err := setField(field, envVal, envKey); err != nil {
			badCfg.addError(err)
		}

	}
	if len(badCfg.errors) > 0 {
		return schema, fmt.Errorf("%s", badCfg.Error())
	}
	return schema, nil
}

func setField(field reflect.Value, value, key string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return nil
	case reflect.Int, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration expression for %s: %w", key, err)
			}
			field.Set(reflect.ValueOf(duration))
			return nil
		}

		num, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int for %s: %w", key, err)
		}
		field.SetInt(num)
		return nil
	case reflect.Bool:
		boolean, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool for %s: %w", key, err)
		}
		field.SetBool(boolean)
		return nil
	case reflect.Float64:
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float for %s: %w", key, err)
		}
		field.SetFloat(num)
		return nil
	default:
		return fmt.Errorf("unsupported type for %s: %s", key, field.Type())
	}
}
