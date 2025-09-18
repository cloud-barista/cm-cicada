SHELL = /bin/bash

MODULE_NAME := "cm-cicada"
PROJECT_NAME := "github.com/cloud-barista/${MODULE_NAME}"
PKG_LIST := $(shell go list ${PROJECT_NAME}/... 2>&1)

GOPROXY_OPTION := GOPROXY=direct
GO_COMMAND := ${GOPROXY_OPTION} go
GOPATH := $(shell go env GOPATH)

.PHONY: all dependency lint test race coverage coverhtml gofmt update swag swagger build run_airflow stop_airflow run run_docker stop stop_docker clean help

all: build

dependency: ## Get dependencies
	@echo Checking dependencies...
	@${GO_COMMAND} mod tidy

lint: dependency ## Lint the files
	@echo "Running linter..."
	@if [ ! -f "${GOPATH}/bin/golangci-lint" ] && [ ! -f "$(GOROOT)/bin/golangci-lint" ]; then \
	  ${GO_COMMAND} install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@main; \
	fi
	@golangci-lint run --timeout 30m -E contextcheck -D unused

test: dependency ## Run unittests
	@echo "Running tests..."
	@${GO_COMMAND} test -v ${PKG_LIST}

race: dependency ## Run data race detector
	@echo "Checking races..."
	@${GO_COMMAND} test -race -v ${PKG_LIST}

coverage: dependency ## Generate global code coverage report
	@echo "Generating coverage report..."
	@${GO_COMMAND} test -v -coverprofile=coverage.out ${PKG_LIST}
	@${GO_COMMAND} tool cover -func=coverage.out

coverhtml: coverage ## Generate global code coverage report in HTML
	@echo "Generating coverage report in HTML..."
	@${GO_COMMAND} tool cover -html=coverage.out

gofmt: ## Run gofmt for go files
	@echo "Running gofmt..."
	@find . -\( -path "./_airflow" -o -path "./vendor" \) -prune -o -type f -name '*.go' -exec $(GOROOT)/bin/gofmt -s -w {} \;

update: ## Update all of module dependencies
	@echo Updating dependencies...
	@cd cmd/${MODULE_NAME} && ${GO_COMMAND} get -u
	@echo Checking dependencies...
	@${GO_COMMAND} mod tidy

swag swagger: ## Generate Swagger Documentation
	@echo "Running swag..."
	@if [ ! -f "${GOPATH}/bin/swag" ] && [ ! -f "$(GOROOT)/bin/swag" ]; then \
	  ${GO_COMMAND} install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g ./pkg/api/rest/server/server.go --pd -o ./pkg/api/rest/docs/ > /dev/null

build: lint swag ## Build the binary file
	@echo Building...
	@kernel_name=`uname -s` && \
	  if [[ $$kernel_name == "Linux" ]]; then \
	    cd cmd/${MODULE_NAME} && ${GO_COMMAND} build -o ${MODULE_NAME} main.go; \
	  elif [[ $$kernel_name == "CYGWIN"* ]] || [[ $$kernel_name == "MINGW"* ]]; then \
	    cd cmd/${MODULE_NAME} && GOOS=windows ${GO_COMMAND} build -o ${MODULE_NAME}.exe main.go; \
	  else \
	    echo $$kernel_name; \
	    echo "Not supported Operating System. ($$kernel_name)"; \
	  fi
	@git diff > .diff_last_build
	@git rev-parse HEAD > .git_hash_last_build
	@echo Build finished!

build-only: swag ## Build the binary file without running linter
	@echo Building...
	@kernel_name=`uname -s` && \
	  if [[ $$kernel_name == "Linux" ]]; then \
	    cd cmd/${MODULE_NAME} && ${GO_COMMAND} build -o ${MODULE_NAME} main.go; \
	  elif [[ $$kernel_name == "CYGWIN"* ]] || [[ $$kernel_name == "MINGW"* ]]; then \
	    cd cmd/${MODULE_NAME} && GOOS=windows ${GO_COMMAND} build -o ${MODULE_NAME}.exe main.go; \
	  else \
	    echo $$kernel_name; \
	    echo "Not supported Operating System. ($$kernel_name)"; \
	  fi
	@git diff > .diff_last_build
	@git rev-parse HEAD > .git_hash_last_build
	@echo Build finished!

run_airflow: ## Run Airflow server
	@mkdir -p _airflow/airflow-home/dags
	@cd _airflow/ && docker compose up -d && cd ..

stop_airflow: ## Stop Airflow server
	@cd _airflow/ && docker compose down && cd ..

run: run_airflow ## Run Airflow server and the built binary
	@sudo killall ${MODULE_NAME} | true
	@git diff > .diff_current
	@STATUS=`diff .diff_last_build .diff_current 2>&1 > /dev/null; echo $$?` && \
	  GIT_HASH_MINE=`git rev-parse HEAD` && \
	  GIT_HASH_LAST_BUILD=`cat .git_hash_last_build 2>&1 > /dev/null | true` && \
	  if [ "$$STATUS" != "0" ] || [ "$$GIT_HASH_MINE" != "$$GIT_HASH_LAST_BUILD" ]; then \
	    $(MAKE) build; \
	  fi
	@cp -RpPf conf cmd/${MODULE_NAME}/ && ./cmd/${MODULE_NAME}/${MODULE_NAME}* || echo "Trying with sudo..." && sudo ./cmd/${MODULE_NAME}/${MODULE_NAME}* &

run_docker: run_airflow ## Run Airflow server and the built binary within Docker
	@docker compose up -d

stop: stop_airflow ## Stop Airflow server and the built binary
	@sudo killall ${MODULE_NAME} | true

stop_docker: stop_airflow ## Stop the Docker containers
	@docker compose down

clean: ## Remove previous build
	@echo Cleaning build...
	@rm -f coverage.out
	@rm -rf cmd/${MODULE_NAME}/conf
	@cd cmd/${MODULE_NAME} && ${GO_COMMAND} clean

	# Run only the Go binary (without Airflow)
run_go: ## Run the built binary only
	@sudo killall ${MODULE_NAME} | true
	@git diff > .diff_current
	@STATUS=`diff .diff_last_build .diff_current 2>&1 > /dev/null; echo $$?` && \
	  GIT_HASH_MINE=`git rev-parse HEAD` && \
	  GIT_HASH_LAST_BUILD=`cat .git_hash_last_build 2>&1 > /dev/null | true` && \
	  if [ "$$STATUS" != "0" ] || [ "$$GIT_HASH_MINE" != "$$GIT_HASH_LAST_BUILD" ]; then \
	    $(MAKE) build; \
	  fi
	@cp -RpPf conf cmd/${MODULE_NAME}/ && ./cmd/${MODULE_NAME}/${MODULE_NAME}* || echo "Trying with sudo..." && sudo ./cmd/${MODULE_NAME}/${MODULE_NAME}* &

# Stop only the Go binary
stop_go: ## Stop the running Go binary only
	@sudo pkill -f ${MODULE_NAME} || true

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
