package parser

import "reflect"

type StringParser struct{}

func (p *StringParser) Parse(value string) (any, error) {
	return value, nil
}

func (p *StringParser) Type() reflect.Type {
	return reflect.TypeOf("")
}
