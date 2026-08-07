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

# Print raw queries (Artist - Title)
shazam-cli list -o query

# Print raw queries with a custom DJ suffix
shazam-cli list -o query --suffix "Extended Mix"
```

### 5. DJ Download Pipeline (Piping into `yt-dlp`)

As a DJ, you can pipe your shazam queries directly into `yt-dlp` to automatically search YouTube, find the best quality tracks/remixes, and download them as high-quality MP3s!

```bash
# Search and download each track as high-quality MP3 (320kbps equivalent)
shazam-cli list -o query | xargs -I {} yt-dlp --extract-audio --audio-format mp3 --audio-quality 0 "ytsearch1:{}"

# Force search for specific DJ mixes or extended versions by adding a suffix
shazam-cli list -o query --suffix "Extended Mix" | xargs -I {} yt-dlp --extract-audio --audio-format mp3 --audio-quality 0 "ytsearch1:{}"
```

### 6. Exporting Tracks

Export your Shazam library to a file:

```bash
# Export to CSV (with headers)
shazam-cli export my_shazams.csv

# Export to pretty-printed JSON
shazam-cli export my_shazams.json
```

### 7. Cache Management

To view stats about your local cache or clear it:

```bash
# Check cache statistics
shazam-cli cache status

# Clear the cached library
shazam-cli cache clear
```

## Output Examples

### Pretty Table Output (`-o table`, default)

```text
INDEX   DATE         TITLE                              ARTIST                   SHAZAM URL
-----   ----         -----                              ------                   ----------
1       2026-08-07   Around the World                   Daft Punk                https://www.shazam.com/track/549952578
2       2026-08-06   One More Time                      Daft Punk                https://www.shazam.com/track/549952579
3       2026-08-05   Harder, Better, Faster, Stronger   Daft Punk                https://www.shazam.com/track/549952580
```

### Simple List Output (`-o list`)

```text
[1] Around the World - Daft Punk (2026-08-07)
[2] One More Time - Daft Punk (2026-08-06)
[3] Harder, Better, Faster, Stronger - Daft Punk (2026-08-05)
```

### Standard CSV Output (`-o csv`)

```text
Index,Date,Title,Artist,URL
1,2026-08-07 19:26:18,Around the World,Daft Punk,https://www.shazam.com/track/549952578
2,2026-08-06 15:10:42,One More Time,Daft Punk,https://www.shazam.com/track/549952579
```

## License
MIT
