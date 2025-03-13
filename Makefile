# Variables
GO = go
GOLANGCI_LINT = golangci-lint
AIR = air
COVERAGE_FILE = coverage.out
SHELL := /bin/bash
MIGRATIONS_PATH = ./internal/database/migrations
DB_ADDR = postgres://admin:adminpassword@localhost/url_shortener_db?sslmode=disable
EXCLUDE_DIRS = mocks|cmd|docs|dump

# Default target - run when only make is used with no args.
.PHONY: all
all: lint vet test coverage swag

# Lint your Go code
.PHONY: lint
lint:
	@echo "Running golangci-lint with config"
	@$(GOLANGCI_LINT) run --config .golangci.yml

# Run tests with coverage and ignore folders
# $(GO) test -coverprofile=$(COVERAGE_FILE) $$(go list ./... | grep -v '/mocks' | grep -v '/cmd/url-shortener')
# go tool cover -html=coverage.out -o coverage.html
.PHONY: coverage
coverage:
	$(GO) test -coverprofile=$(COVERAGE_FILE) $$(go list ./... | grep -vE '/($(EXCLUDE_DIRS))')
	$(GO) tool cover -html=$(COVERAGE_FILE)

# Notes: $$ to escape $
.PHONY: coverage-percentage
coverage-percentage:
	@echo "Running coverage tool..."
	$(GO) test -coverprofile=$(COVERAGE_FILE) $$(go list ./... | grep -vE '/($(EXCLUDE_DIRS))')
	@coverage=$$(go tool cover -func=coverage.out | tail -n 1 | awk '{print $$3}' | sed 's/%//' | tr -d '\n'); \
	echo "Coverage: $$coverage"; \
	if [ $$(echo "$$coverage < 70" | bc -l) -eq 1 ]; then \
		echo "Test coverage is below 70%. Current coverage: $$coverage%."; \
		exit 1; \
	fi

# Run tests without coverage reporting
.PHONY: test
test:
	$(GO) test -race ./...

# Run the app in development mode with air (live-reload)
.PHONY: dev
dev:
	docker-compose up -d
	make migrate-up
	docker logs -f go-url-shortener-dev

# Run the go outside of the docker container with live reload.
.PHONY: dev-local
dev-local:
	docker-compose up -d
	docker-compose stop go-url-shortener-dev
	make migrate-up
	$(AIR)

# Run the app manually with Go run without air
.PHONY: run
run:
	@$(GO) run ./cmd/url-shortener/main.go || echo "Shutdown Completed"

# Clean up generated files (e.g., coverage report)
.PHONY: clean
clean:
	rm -f $(COVERAGE_FILE)

# Vet your Go code for suspicious constructs
.PHONY: vet
vet:
	$(GO) vet ./...

# Build Go code.
.PHONY: build
build:
	@echo "Building Go application..."
	@go build -o url-shortener ./cmd/url-shortener/main.go


# Generate go code eg. Mocks
.PHONY: generate
generate:
	$(GO) generate ./...

# Check if generated files should be created.
.PHONY: check-generate
check-generate:
	go generate ./...
	@if [[ -n "$$(git status --porcelain)" ]]; then \
		echo "❌ Uncommitted changes after 'go generate'."; \
		git status; \
		git diff; \
		exit 1; \
	else \
		echo "✅ No changes after 'go generate'."; \
	fi

# Generate swagger docs.
.PHONY: swag
swag:
	swag init -d ./cmd/url-shortener --pdl 3

# Create new SQL migration.
.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

# Run SQL migrations.
# make clean-migrations; \
# echo "Re-running migrations..."; \
# migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" up; \
.PHONY: migrate-up
migrate-up:
	@if migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" up; then \
		echo "Migrations applied successfully."; \
	else \
    	echo "Please run the command below to resolve the issue:"; \
		echo -e "\033[1;31m\tmake clean-migrations\033[0m"; \
		exit 1; \
	fi

# Down SQL migrations.
.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" down $(filter-out $@,$(MAKECMDGOALS))

# Reset the migrations
.PHONY: clean-migrations
clean-migrations:
	@echo "Cleaning migration table ..."
	@migrate -path=$(MIGRATIONS_PATH) -database $(DB_ADDR) drop -f

# Install all required tools.
.PHONY: install-tools
install-tools:
	@echo "Installing dev tools..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
