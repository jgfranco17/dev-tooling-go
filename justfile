# Default command
_default:
    @just --list --unsorted

# Execute unit tests
test:
    go clean -testcache
    go test -cover ./...

# Sync Go modules
tidy:
    @go mod tidy
    @echo "Go modules synced successfully!"
