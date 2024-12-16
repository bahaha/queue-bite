package parser

import (
	"reflect"
	"time"
)

type DurationParser struct{}

func (p *DurationParser) Parse(value string) (any, error) {
	return time.ParseDuration(value)
}

func (p *DurationParser) Type() reflect.Type {
	return reflect.TypeOf(time.Duration(0))
}
