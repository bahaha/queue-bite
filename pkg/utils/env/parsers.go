package env

import (
	"reflect"

	parsers "queue-bite/pkg/utils/env/parsers"
)

type ValueParser interface {
	Parse(value string, field reflect.Value) error
	Type() reflect.Type
}

var defaultRegistry = newParserRegistry()

func RegisterParser(parser ValueParser) {
	defaultRegistry.Register(parser)
}

func (r *parserRegistry) Register(parser ValueParser) {
	r.parsers = append(r.parsers, parser)
}

type parserRegistry struct {
	parsers []ValueParser
}

func newParserRegistry() *parserRegistry {
	r := &parserRegistry{
		parsers: make([]ValueParser, 0),
	}
	r.registerBuiltins()
	return r
}

func (r *parserRegistry) registerBuiltins() {
	builtins := []ValueParser{
		&parsers.StringParser{},
		&parsers.BoolParser{},
		&parsers.IntParser{},
		&parsers.FloatParser{},
		&parsers.DurationParser{},
	}
	for _, parser := range builtins {
		r.Register(parser)
	}
}

func (r *parserRegistry) FindParser(t reflect.Type) (ValueParser, bool) {
	for _, parser := range r.parsers {
		if parser.Type() == t {
			return parser, true
		}
	}
	return nil, false
}
