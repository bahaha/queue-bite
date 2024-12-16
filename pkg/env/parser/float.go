package parser

import (
	"reflect"
	"strconv"
)

type FloatParser struct{}

func (p *FloatParser) Parse(value string) (any, error) {
	v, err := strconv.ParseFloat(value, 64)
	return v, err
}

func (p *FloatParser) Type() reflect.Type {
	return reflect.TypeOf(float64(0))
}
