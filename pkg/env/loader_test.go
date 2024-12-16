package env

import (
	"testing"
	"time"
)

type testEnv struct {
	kvStore map[string]string
}

func newTestEnv() *testEnv {
	return &testEnv{
		kvStore: make(map[string]string),
	}
}

func (e *testEnv) setEnv(vars map[string]string) {
	for k, v := range vars {
		e.kvStore[k] = v
	}
}

func (e *testEnv) getenv(key string) string {
	return e.kvStore[key]
}

func TestLoader(t *testing.T) {
	t.Parallel()

	t.Run("string validation", func(t *testing.T) {
		t.Parallel()
		env := newTestEnv()

		type StrConf struct {
			DbHost       string `env:"DB_HOST" required:"T"`
			DatabaseName string `env:"DATABASE_NAME" default:"0"`
			Password     string `env:"PASSWORD"`
		}

		cfg := &StrConf{}
		loader := NewEnvLoader(WithEnvSource(env.getenv))
		err := loader.Parse(cfg)
		if err == nil {
			t.Error("expected error for missing required string, got nil")
		}

		env.setEnv(map[string]string{"DB_HOST": "localhost"})

		err = loader.Parse(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.DbHost != "localhost" {
			t.Errorf("expected DB_HOST to be 'localhost', got %v", cfg.DbHost)
		}
		if cfg.DatabaseName != "0" {
			t.Errorf("expected DATABASE_NAME to be its default value '0', got %v", cfg.DatabaseName)
		}
		if cfg.Password != "" {
			t.Errorf("expected optional variable PASSWORD to be '', got %v", cfg.Password)
		}

		env.setEnv(map[string]string{"DATABASE_NAME": "QUEUEBiTE"})
		err = loader.Parse(cfg)
		if cfg.DatabaseName != "QUEUEBiTE" {
			t.Errorf("expected DATABASE_NAME to be its override value 'QUEUEBiTE', got %v", cfg.DatabaseName)
		}
	})

	t.Run("int validation", func(t *testing.T) {
		t.Parallel()
		env := newTestEnv()

		type IntConf struct {
			ServerPort int `env:"SERVER_PORT"`
			RedisPort  int `env:"REDIS_PORT" default:"6379"`
		}
		env.setEnv(map[string]string{"SERVER_PORT": "55688"})

		cfg := &IntConf{}
		loader := NewEnvLoader(WithEnvSource(env.getenv))
		err := loader.Parse(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.ServerPort != 55688 {
			t.Errorf("expected SERVER_PORT to be `55688` from the env var, got %v", cfg.ServerPort)
		}
		if cfg.RedisPort != 6379 {
			t.Errorf("expected REDIS_PORT to be its default value `6379`, got %v", cfg.RedisPort)
		}
	})

	t.Run("int validation", func(t *testing.T) {
		t.Parallel()
		env := newTestEnv()

		type IntConf struct {
			ServerPort int `env:"SERVER_PORT"`
			RedisPort  int `env:"REDIS_PORT" default:"6379"`
		}
		env.setEnv(map[string]string{"SERVER_PORT": "55688"})

		cfg := &IntConf{}
		loader := NewEnvLoader(WithEnvSource(env.getenv))
		err := loader.Parse(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.ServerPort != 55688 {
			t.Errorf("expected SERVER_PORT to be `55688` from the env var, got %v", cfg.ServerPort)
		}
		if cfg.RedisPort != 6379 {
			t.Errorf("expected REDIS_PORT to be its default value `6379`, got %v", cfg.RedisPort)
		}
	})

	t.Run("float validation", func(t *testing.T) {
		t.Parallel()
		env := newTestEnv()

		type FloatConf struct {
			Pi float64 `env:"PI"`
		}
		env.setEnv(map[string]string{"PI": "3.14159"})

		cfg := &FloatConf{}
		loader := NewEnvLoader(WithEnvSource(env.getenv))
		err := loader.Parse(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.Pi != 3.14159 {
			t.Errorf("expected PI to be `3.14159` from the env var, got %v", cfg.Pi)
		}
	})

	t.Run("bool validation", func(t *testing.T) {
		t.Parallel()
		env := newTestEnv()

		type BoolConf struct {
			BoolVal bool `env:"BOOL_V"`
		}
		env.setEnv(map[string]string{"BOOL_V": "T"})

		cfg := &BoolConf{}
		loader := NewEnvLoader(WithEnvSource(env.getenv))
		err := loader.Parse(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if cfg.BoolVal != true {
			t.Errorf("expected BOOL_V to be `true`, got %v", cfg.BoolVal)
		}
	})

	t.Run("duration validation", func(t *testing.T) {
		t.Parallel()
		env := newTestEnv()

		type DurationConf struct {
			HealthCheckDuration time.Duration `env:"HEALTH_CHECK_DURATION"`
			ShutdownTimeout     time.Duration `env:"SHUTDOWN_TIMEOUT" default:"5s"`
		}
		env.setEnv(map[string]string{"HEALTH_CHECK_DURATION": "5s"})

		cfg := &DurationConf{}
		loader := NewEnvLoader(WithEnvSource(env.getenv))
		err := loader.Parse(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if cfg.HealthCheckDuration != 5*time.Second {
			t.Errorf("expected HEALTH_CHECK_DURATION to be `5s`, got %v", cfg.HealthCheckDuration)
		}
		if cfg.ShutdownTimeout != 5*time.Second {
			t.Errorf("expected SHUTDOWN_TIMEOUT to be its default value `5s`, got %v", cfg.ShutdownTimeout)
		}
	})

}
