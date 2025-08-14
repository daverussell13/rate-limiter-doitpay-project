COVERAGE_FILE := coverage.out

.PHONY: test
test:
	go test -v -coverprofile=$(COVERAGE_FILE) ./...

.PHONY: coverage
coverage:
ifeq ($(OS),Windows_NT)
	if exist $(COVERAGE_FILE) (go tool cover -html=$(COVERAGE_FILE)) else (echo "Coverage file not found, run 'make test' first")
else
	go tool cover -html=$(COVERAGE_FILE)
endif

.PHONY: clean
clean:
	del $(COVERAGE_FILE) 2>nul || true
