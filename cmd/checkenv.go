package main

import (
	"log"
	"os"

	"queue-bite/internal/config"
	_ "queue-bite/pkg/env/autoload"
)

func main() {
	_, err := config.LoadEnvConfig(os.Getenv)
	if err != nil {
		log.Fatalf("Invalid server configuration, check environment variables:\n%s", err)
	}
}
