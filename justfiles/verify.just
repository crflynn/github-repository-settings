# Keep recipes alphabetized.
# Run formatting, tests, vet, and build checks.
check: build format-check test vet

# Format Go source files.
format:
    find . -type f -name '*.go' -not -path './.git/*' -print0 | xargs -0 gofmt -w

# Check Go source formatting without modifying files.
format-check:
    test -z "$(find . -type f -name '*.go' -not -path './.git/*' -print0 | xargs -0 gofmt -l)"

# Run the Go test suite.
test:
    go test ./...

# Run Go static analysis.
vet:
    go vet ./...
