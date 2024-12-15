package env

import (
	"reflect"
	"time"
)

type DurationParser struct{}

func (p *DurationParser) Parse(value string, field reflect.Value) error {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(duration))
	return nil
}

func (p *DurationParser) Type() reflect.Type {
	return reflect.TypeOf(time.Duration(0))
}
