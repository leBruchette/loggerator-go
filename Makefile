.PHONY: run-server test test-coverage

# Task to run the server
run-server:
	go run main.go

# Task to run tests
test:
	go test ./...

# Task to run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out