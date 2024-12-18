# Simple Makefile for a Go project

# Build the application
all: build test
templ-install:
	@if ! command -v templ > /dev/null; then \
		read -p "Go's 'templ' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/a-h/templ/cmd/templ@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "templ installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install templ. Exiting..."; \
			exit 1; \
		fi; \
	fi
tailwind-install:
	
	@if [ ! -f tailwindcss ]; then curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64 -o tailwindcss; fi
	@chmod +x tailwindcss

build: tailwind-install templ-install
	@echo "Building..."
	@go run cmd/checkenv.go || exit 1
	@templ generate
	@./tailwindcss -i internal/server/assets/css/input.css -o internal/server/assets/css/output.css
	@go build -o main cmd/api/main.go

# Run the application
run:
	@APP_ENV=dev go run cmd/api/main.go

redis-run:
	@if command -v docker compose >/dev/null 2>&1; then \
		docker compose up redis_bp -d; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up redis_bp -d; \
	fi


# Create DB container
docker-run:
	@cp .env.docker .env
	@if command -v docker compose >/dev/null 2>&1; then \
		docker compose up --build; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if command -v docker compose >/dev/null 2>&1; then \
		docker compose down; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@APP_ENV=test go test ./... 

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch: redis-run
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch tailwind-install templ-install
