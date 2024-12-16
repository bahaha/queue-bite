package parser

import (
	"reflect"
	"strconv"
)

type BoolParser struct{}

func (p *BoolParser) Parse(value string) (any, error) {
	v, err := strconv.ParseBool(value)
	return v, err
}

func (p *BoolParser) Type() reflect.Type {
	return reflect.TypeOf(true)
}
