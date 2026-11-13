# Keep recipes alphabetized.
# Build all Go packages.
build:
    go build ./...

# Install the project tools pinned in mise.toml.
setup:
    mise install
