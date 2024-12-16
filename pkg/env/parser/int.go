package parser

import (
	"reflect"
	"strconv"
)

type IntParser struct{}

func (p *IntParser) Parse(value string) (any, error) {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}

	return int(v), nil
}

func (p *IntParser) Type() reflect.Type {
	return reflect.TypeOf(int(0))
}
