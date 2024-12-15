package env

import (
	"reflect"
	"strconv"
)

type IntParser struct{}

func (p *IntParser) Parse(value string, field reflect.Value) error {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}

	field.SetInt(v)
	return nil
}

func (p *IntParser) Type() reflect.Type {
	return reflect.TypeOf(int(0))
}
