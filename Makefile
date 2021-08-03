.PHONY: all build deps image lint migrate test vet
CHECK_FILES?=$$(go list ./... | grep -v /vendor/)
FLAGS?=-ldflags "-X github.com/supabase/supabase-admin-api/cmd.Version=`git rev-parse HEAD`"

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: lint vet test build ## Run the tests and build the binary.

build: ## Build the binary.
	go build $(FLAGS)
	GOOS=linux GOARCH=arm64 go build $(FLAGS) -o supabase-admin-api-arm64

deps: ## Install dependencies.
	@go get -u golang.org/x/lint/golint
	@go mod download

lint: ## Lint the code.
	golint $(CHECK_FILES)

test: ## Run tests.
	go test -p 1 -v $(CHECK_FILES)

vet: # Vet the code
	go vet $(CHECK_FILES)
