// Package env provides a flexible and type-safe way to load configuration from environment variables.
package env

import "reflect"

// Field represents a single configuration field and its metadata.
// Each field corresponds to an environment variable and includes
// parsing and validation information.
//
// Example:
//
// // Field representing a database port
//
// // > Port int `env:"DB_HOST" default:"3000"`
//
// Name: "Port", EnvKey: "DB_HOST", Type: <int>, Required: false, DefaultVal: 3000
type Field struct {
	Name       string
	EnvKey     string
	Type       reflect.Type
	Required   bool
	DefaultVal string
}

// Parser convert string environment variable to typed Go values.
// Each parser is responsible to converting strings to a specific Go type.
//
// Example:
//
// // Duration parser implementation
// type DurationParser struct {}
//
//	func (p *DurationParser) Parse(value string) (any, error) {
//	    return time.ParseDuration(value)
//	}
//
//	func (p *DurationParser) Type() reflect.Type {
//	    return reflect.TypeOf(time.Duration(0))
//	}
//
// Usage:
// parser := &DurationParser{}
// value, err := parser.Parse("5m")  // Returns 5 minutes duration
type Parser interface {
	// Parse converts a string value to the appropriate Go type
	Parse(value string) (any, error)
	// Type returns the Go type this parser handles
	Type() reflect.Type
}

// Option represents a configuration option for the loader
// Options are used to custimize the loader's behavior.
//
// Example:
//
// loader := NewLoader(
//
//	WithValidator(CustomValidator{}),
//
// )
type Option func(*Loader)

//	 Loader loads configuration from environment variables into a struct.
//	 It handle parsing, validation, and error collection.
//
//	 Example:
//
//		type Config struct {
//		    Server struct {
//		        Port    int           `env:"PORT" default:"8080"`
//		        Timeout time.Duration `env:"TIMEOUT" default:"30s"`
//		    }
//		    Database struct {
//		        URL      string `env:"DB_URL" required:"true"`
//		        MaxConns int    `env:"DB_MAX_CONNS" default:"10"`
//		    }
//		}
//
//	 cfg := &Config{}
//	 loader := NewLoader()
//	 if err := loader.Load(cfg); err != nil {
//	     log.Fatalf("Failed to load config: %v"), err)
//	 }
type Loader struct {
	source   func(key string) string
	registry Registry
}

// Registry manages the set of available parsers.
// It provides type-safe parsing for supported Go types.
type Registry interface {
	Register(Parser) error
	FindParser(reflect.Type) (Parser, bool)
}
