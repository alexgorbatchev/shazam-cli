# Build shazam-cli
build:
    go build -o shazam-cli cmd/shazam-cli/main.go

# Run development version
run *args:
    go run cmd/shazam-cli/main.go {{args}}

# Tidy dependencies
tidy:
    go mod tidy
