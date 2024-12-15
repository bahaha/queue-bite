package env

import (
	"testing"
	"time"
)

func TestParser(t *testing.T) {
	kvStore := make(map[string]string)
	setEnv := func(vars map[string]string) {
		for k, v := range vars {
			kvStore[k] = v
		}
	}

	cleanEnv := func() {
		kvStore = make(map[string]string)
	}

	getenv := func(key string) string {
		return kvStore[key]
	}

	t.Run("string validation", func(t *testing.T) {
		type StrConf struct {
			DbHost       string `env:"DB_HOST" required:"T"`
			DatabaseName string `env:"DATABASE_NAME" default:"0"`
			Password     string `env:"PASSWORD"`
		}

		cfg, err := LoadConfig(getenv, &StrConf{})
		if err == nil {
			t.Error("expected error for missing required string, got nil")
		}

		setEnv(map[string]string{"DB_HOST": "localhost"})
		t.Cleanup(cleanEnv)

		cfg, err = LoadConfig(getenv, &StrConf{})
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

		setEnv(map[string]string{"DATABASE_NAME": "QUEUEBiTE"})
		cfg, err = LoadConfig(getenv, &StrConf{})
		if cfg.DatabaseName != "QUEUEBiTE" {
			t.Errorf("expected DATABASE_NAME to be its override value 'QUEUEBiTE', got %v", cfg.DatabaseName)
		}
	})

	t.Run("int validation", func(t *testing.T) {
		type IntConf struct {
			ServerPort int `env:"SERVER_PORT"`
			RedisPort  int `env:"REDIS_PORT" default:"6379"`
		}
		setEnv(map[string]string{"SERVER_PORT": "55688"})
		t.Cleanup(cleanEnv)

		cfg, err := LoadConfig(getenv, &IntConf{})
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
		type FloatConf struct {
			Pi float64 `env:"PI"`
		}
		setEnv(map[string]string{"PI": "3.14159"})
		t.Cleanup(cleanEnv)

		cfg, err := LoadConfig(getenv, &FloatConf{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.Pi != 3.14159 {
			t.Errorf("expected PI to be `3.14159` from the env var, got %v", cfg.Pi)
		}
	})

	t.Run("bool validation", func(t *testing.T) {
		type BoolConf struct {
			BoolVal bool `env:"BOOL_V"`
		}
		setEnv(map[string]string{"BOOL_V": "T"})
		t.Cleanup(cleanEnv)

		cfg, err := LoadConfig(getenv, &BoolConf{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if cfg.BoolVal != true {
			t.Errorf("expected BOOL_V to be `true`, got %v", cfg.BoolVal)
		}
	})

	t.Run("duration validation", func(t *testing.T) {
		type DurationConf struct {
			HealthCheckDuration time.Duration `env:"HEALTH_CHECK_DURATION"`
			ShutdownTimeout     time.Duration `env:"SHUTDOWN_TIMEOUT" default:"5s"`
		}
		setEnv(map[string]string{"HEALTH_CHECK_DURATION": "5s"})
		t.Cleanup(cleanEnv)

		cfg, err := LoadConfig(getenv, &DurationConf{})
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
