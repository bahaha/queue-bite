package env

import (
	"reflect"
	"strconv"
)

type BoolParser struct{}

func (p *BoolParser) Parse(value string, field reflect.Value) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}

	field.SetBool(v)
	return nil
}

func (p *BoolParser) Type() reflect.Type {
	return reflect.TypeOf(true)
}
