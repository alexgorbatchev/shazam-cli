# shazam-cli

A Go command line tool to fetch, list, export, and locally cache your Shazam library tracks. It supports manually specifying your session cookie or automatically extracting and decrypting it from your Google Chrome or Vivaldi browser profiles.

## Features
- **Auto Cookie Extraction**: Reads and decrypts `shazam.com` session cookies natively from Google Chrome or Vivaldi profiles on macOS, Windows, and Linux using `kooky`.
- **Offline JSON Cache**: Locally caches retrieved tracks under `~/.cache/shazam-cli/library.json` so that viewing your library is instantaneous and avoids repeated API requests.
- **Paginated Fetching**: Automatically paginates through your entire Shazam history (no limits).
- **Flexible Exports**: Export your library to `.csv` or `.json` files.
- **Beautiful Outputs**: Prints cleanly structured text tables, simple lists, or raw CSV to stdout.

## Installation

### 1. Download Pre-compiled Binary (Recommended)

Download the latest release binary for your operating system and architecture directly from the [GitHub Releases](https://github.com/alexgorbatchev/shazam-cli/releases) page, or download it via the GitHub CLI (`gh`):

```bash
# Example: Download macOS Apple Silicon (Arm64) binary
gh release download --repo alexgorbatchev/shazam-cli --pattern "*Darwin_arm64.tar.gz"
tar -xzf shazam-cli_*_Darwin_arm64.tar.gz
mv shazam-cli /usr/local/bin/
```

### 2. Install from Source (Go Developers)

If you have Go installed (version 1.26.5 or later), you can build the utility manually:

```bash
go install github.com/alexgorbatchev/shazam-cli/cmd/shazam-cli@latest
```

Or clone and build:

```bash
git clone https://github.com/alexgorbatchev/shazam-cli.git
cd shazam-cli
go build -o shazam-cli cmd/shazam-cli/main.go
```

---

## Usage

### 1. Automatic Browser Cookie Retrieval (Recommended)

Make sure you have logged into `shazam.com` in either Google Chrome or Vivaldi. Run:

```bash
# Auto-extract from Chrome (Default profile)
shazam-cli list -b chrome

# Auto-extract from Vivaldi
shazam-cli list -b vivaldi

# Specify a custom profile
shazam-cli list -b vivaldi -p "Profile 2"
```

### 2. Manual Cookie Passing

If auto-extraction isn't supported on your platform/browser, log into `shazam.com`, copy your full request cookie header, and run:

```bash
# Via CLI flag
shazam-cli list -c "ini=...; token=...; shazam_session=..."

# Via Environment Variable
export SHAZAM_COOKIE="ini=...; token=..."
shazam-cli
```

### 3. Displaying and Offline Caching

Once fetched, tracks are saved locally in your user cache folder. Subsequent calls are instantaneous and run fully offline:

```bash
# Loads instantaneously from local cache!
shazam-cli list

# Force reload from Shazam API
shazam-cli list --no-cache
```

### 4. Custom Output Formats

Toggle output styles via `-o` or `--output`:

```bash
# Pretty text table (default)
shazam-cli list -o table

# Simple list (ideal for piping)
shazam-cli list -o list

# Print standard CSV to stdout
shazam-cli list -o csv
```

### 5. Exporting Tracks

Export your Shazam library to a file:

```bash
# Export to CSV (with headers)
shazam-cli export my_shazams.csv

# Export to pretty-printed JSON
shazam-cli export my_shazams.json
```

### 6. Cache Management

To view stats about your local cache or clear it:

```bash
# Check cache statistics
shazam-cli cache status

# Clear the cached library
shazam-cli cache clear
```

## License
MIT
