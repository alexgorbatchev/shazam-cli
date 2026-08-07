# shazam-cli Agent Guidelines

Workspace for `shazam-cli` — a Go-based CLI utility (`shazam-cli`) that fetches, displays, exports, and locally caches your Shazam library tracks. It supports manual session cookies or auto-extracting and decrypting cookies natively from local browser profiles (Chrome/Vivaldi/Safari).

## Repository & GitHub Releases
- Public GitHub Repository: [`github.com/alexgorbatchev/shazam-cli`](https://github.com/alexgorbatchev/shazam-cli)
- Binary name: `shazam-cli`

## Installation
- **Via Go**:
  ```bash
  go install github.com/alexgorbatchev/shazam-cli/cmd/shazam-cli@latest
  ```

## Shared Commands
- Build binary: `just build` (outputs `shazam-cli`)
- Run development: `just run [args]`
- Tidy dependencies: `just tidy`
- Run unit tests: `go test -v ./...`
- Run coverage report: `go test -coverprofile=coverage.out ./...`

## Architecture Map
- `cmd/shazam-cli/main.go`: Cobra CLI entrypoint, argument parsing, subcommands (`list`, `export`, `cache`), console printing, formatting (table, list, CSV), and execution flow.
- `pkg/browser/cookies.go`: Locates Chrome, Vivaldi, and Safari profiles across macOS, Windows, and Linux. Uses `github.com/browserutils/kooky` (via replace) to traverse and decrypt cookie databases on the `shazam.com` domain.
- `pkg/shazam/client.go`: High-performance HTTP client that calls Shazam's private web endpoint `https://www.shazam.com/services/profile/v1/en-US/US/web/profile` with appropriate cookies, user-agent, and pagination.
- `pkg/shazam/types.go`: Structs representing Shazam API responses and tags. Includes a robust recursive JSON parser `ParseTracksFromJSON` that dynamically searches for track list nodes to guard against API drift.
- `pkg/shazam/cache.go`: Manages local JSON persistence under the user's cache folder (`~/.cache/shazam-cli/library.json`), enabling super-fast offline listing.

## CLI Usage (Assuming `shazam-cli` is on PATH)
- `shazam-cli list`: List all tracks (loads from local cache if initialized).
- `shazam-cli list -b chrome`: Extract cookie from Chrome, fetch fresh tracks, update cache, and display.
- `shazam-cli list -b vivaldi`: Extract cookie from Vivaldi, fetch, cache, and display.
- `shazam-cli list -b safari`: Extract cookie from Safari (macOS only), fetch, cache, and display.
- `shazam-cli list -o query --suffix "Extended Mix"`: Output list as raw search strings with custom suffixes for downloading.
- `shazam-cli export <file>`: Export your Shazam library as CSV or JSON (e.g., `export my_lib.csv` or `export my_lib.json`).
- `shazam-cli cache [status|clear]`: View cache statistics or purge the cached library.

## Mandatory Maintenance Boundaries
1. **RESIENT PARSING REQUIREMENT**: Never assume a rigid schema for Shazam's private endpoints. If JSON structure changes, the recursive parser (`types.go:findArrayInMap`) must be maintained to dynamically seek tags/tracks/items arrays in the map.
2. **SECURE COOKIE HANDLING**: Decrypted cookies must only be kept in-memory and never logged, exported, or persisted to disk. Only the track list meta details may be cached.
3. **NO BINARIES IN GIT**: Compiled binaries (`shazam-cli`, `shazam-cli.exe`) must always be excluded via `.gitignore`.
