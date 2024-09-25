CHECKER_BIN=$(PWD)/tmp/bin
COMMIT ?= $(shell git describe --dirty --long --always --abbrev=15)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION ?= "latest"

$(CHECKER_BIN)/.installed: $(CHECKER_BIN)/.installed ## Install static checkers
	@echo "🚚 Downloading binaries.."
	@GOBIN=$(CHECKER_BIN) go install mvdan.cc/gofumpt@latest
	@GOBIN=$(CHECKER_BIN) go install golang.org/x/vuln/cmd/govulncheck@latest
	@GOBIN=$(CHECKER_BIN) go install github.com/securego/gosec/v2/cmd/gosec@latest
	@GOBIN=$(CHECKER_BIN) go install github.com/swaggo/swag/cmd/swag@latest
	@GOBIN=$(CHECKER_BIN) go install github.com/air-verse/air@latest
	@touch $(CHECKER_BIN)/.installed

lint: $(CHECKER_BIN)/.installed ## Lint the source code
	@echo "🧹 Cleaning go.mod.."
	@go mod tidy
	@echo "🧹 Formatting files.."
	@go fmt ./...
	@$(CHECKER_BIN)/gofumpt -l -w .
	@echo "🧹 Vetting go.mod.."
	@go vet ./...
	@echo "🧹 GoCI Lint.."
	@golangci-lint run ./...

vuln: $(CHECKER_BIN)/.installed ## Check for vulnerabilities
	@echo "🔍 Checking for vulnerabilities"
	@#$(CHECKER_BIN)/govulncheck -test ./...
	@$(CHECKER_BIN)/gosec -quiet -exclude=G104 ./...

run: $(CHECKER_BIN)/.installed ## Run Glide
	@go run semantic-cache -o ./dist;

build: ## Build Glide
	@echo "🔨Building Semantic Cache binary.."
	@echo "Build Date: $(BUILD_DATE)"
	@go build -o ./dist;

test: ## Run tests
	@go test -v -count=1 -race -shuffle=on -coverprofile=coverage.out ./...

test-cov: ## Run tests with coverage
	@go tool cover -func=coverage.out

docs-api: $(CHECKER_BIN)/.installed ## Generate OpenAPI API docs
	@$(CHECKER_BIN)/swag init

telemetry-up: ## Start observability services needed to receive Glides signals
	@docker-compose --profile telemetry up --wait
	@echo "Jaeger UI: http://localhost:16686/"
	@echo "Grafana UI: http://localhost:3000/"

telemetry-down: ## Shutdown observability services needed to receive Glides signals
	@docker-compose --profile telemetry down