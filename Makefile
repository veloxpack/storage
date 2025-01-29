# Variables
PKG := ./...
TEST_FLAGS := -v

# Default target
all: test

# Run tests
test:
	go test $(TEST_FLAGS) $(PKG)

# Run tests with coverage
test-cover:
	go test -cover $(TEST_FLAGS) $(PKG)

# Clean test cache
test-clean:
	go clean -testcache
