run:
	@echo Running server...
	@go run ./cmd/app/main.go

lint:
	golangci-lint run ./...

generate-mocks:
	@go install github.com/vektra/mockery/v2@latest
	@mockery --dir internal/repository --all --output internal/repository/mocks --outpkg mocks --with-expecter

test-unit:
	@echo Running unit-tests...
	@go test ./tests/unit/... -v

test-e2e:
	set ENV=test && go test ./tests/e2e/... -v -count=1
