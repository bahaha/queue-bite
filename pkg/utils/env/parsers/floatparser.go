package env

import (
	"reflect"
	"strconv"
)

type FloatParser struct{}

func (p *FloatParser) Parse(value string, field reflect.Value) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}

	field.SetFloat(v)
	return nil
}

func (p *FloatParser) Type() reflect.Type {
	return reflect.TypeOf(float64(0))
}
