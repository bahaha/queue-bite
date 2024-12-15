package env

import "reflect"

type StringParser struct{}

func (p *StringParser) Parse(value string, field reflect.Value) error {
	field.SetString(value)
	return nil
}

func (p *StringParser) Type() reflect.Type {
	return reflect.TypeOf("")
}
