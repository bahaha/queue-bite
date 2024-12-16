package autoload

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	env := os.Getenv("APP_ENV")
	if "" == env {
		env = "local"
	}

	if env == "production" {
		return
	}

	envFilename := fmt.Sprintf(".env.%s", env)
	godotenv.Load()
	godotenv.Overload(envFilename)

}
