package env

import (
	"fmt"
	"reflect"

	"queue-bite/pkg/env/parser"
)

type defaultRegistry struct {
	parsers map[reflect.Type]Parser
}

func NewBuiltinRegistry() Registry {
	r := &defaultRegistry{
		parsers: make(map[reflect.Type]Parser),
	}
	r.registerBuiltinParsers()
	return r
}

func (r *defaultRegistry) Register(p Parser) error {
	if existing, exists := r.parsers[p.Type()]; exists {
		return fmt.Errorf("parser already registered for type %v: %T", p.Type(), existing)
	}
	r.parsers[p.Type()] = p
	return nil
}

func (r *defaultRegistry) FindParser(t reflect.Type) (Parser, bool) {
	p, ok := r.parsers[t]
	return p, ok
}

func (r *defaultRegistry) registerBuiltinParsers() {
	builtins := []Parser{
		&parser.StringParser{},
		&parser.IntParser{},
		&parser.BoolParser{},
		&parser.FloatParser{},
		&parser.DurationParser{},
	}

	for _, p := range builtins {
		r.Register(p)
	}
}
