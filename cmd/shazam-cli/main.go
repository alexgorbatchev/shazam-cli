package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexgorbatchev/shazam-cli/pkg/browser"
	"github.com/alexgorbatchev/shazam-cli/pkg/shazam"
)

var (
	version = "dev"
)

type cliOptions struct {
	cookie      string
	browserName string
	profile     string
	noCache     bool
	noSave      bool
	output      string
	suffix      string
}

func NewRootCmd() *cobra.Command {
	opts := &cliOptions{}

	rootCmd := &cobra.Command{
		Use:     "shazam-cli [command|flags]",
		Short:   "A Go CLI to fetch, list, and cache your Shazam library tracks",
		Version: version,
		Long: `shazam-cli is a fast Go command-line utility that extracts your Shazam library
either by reading cookies automatically from local browser profiles (Chrome/Vivaldi)
or using a manually provided session cookie. It paginates and caches tracks locally.

Subcommands:
  shazam-cli list              List all tracks in your library (default action)
  shazam-cli export <file>     Export all tracks to a CSV or JSON file
  shazam-cli cache status      Display information about the local JSON cache
  shazam-cli cache clear       Clear the local JSON cache
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to 'list'
			return runList(cmd, opts)
		},
	}

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&opts.cookie, "cookie", "c", "", "Manual Shazam session cookie string")
	flags.StringVarP(&opts.browserName, "browser", "b", "", "Browser to auto-extract cookie from (e.g. chrome, vivaldi)")
	flags.StringVarP(&opts.profile, "profile", "p", "Default", "Browser profile name to read cookies from")
	flags.BoolVar(&opts.noCache, "no-cache", false, "Force fetching fresh data from Shazam API (bypasses cache)")
	flags.BoolVar(&opts.noSave, "no-save", false, "Disable saving fetched tracks to local JSON cache")
	flags.StringVarP(&opts.output, "output", "o", "table", "Output format for console (table, list, csv, query)")
	flags.StringVar(&opts.suffix, "suffix", "", "Custom suffix to append to each track in query output (e.g. 'Extended Mix')")

	// Add subcommands
	rootCmd.AddCommand(
		newListCmd(opts),
		newExportCmd(opts),
		newCacheCmd(),
	)

	return rootCmd
}

func newListCmd(opts *cliOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tracks in your Shazam library",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}
}

func newExportCmd(opts *cliOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "export <file>",
		Short: "Export all tracks to a file (CSV or JSON)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			return runExport(cmd, opts, filePath)
		},
	}
}

func newCacheCmd() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage local library cache",
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show local cache file statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := shazam.NewCache()
			if err != nil {
				return err
			}
			path := c.GetCachePath()
			cmd.Printf("Cache File Path: %s\n", path)

			info, err := os.Stat(path)
			if os.IsNotExist(err) {
				cmd.Println("Cache status: Empty / Not initialized")
				return nil
			} else if err != nil {
				return err
			}

			tracks, err := c.LoadTracks()
			if err != nil {
				cmd.Printf("Cache Status: Corrupted (%v)\n", err)
				return nil
			}

			cmd.Printf("Cached tracks:   %d songs\n", len(tracks))
			cmd.Printf("Last modified:   %s\n", info.ModTime().Format("2006-01-02 15:04:05 MST"))
			cmd.Printf("File size:       %.2f KB\n", float64(info.Size())/1024.0)

			return nil
		},
	}

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the local JSON cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := shazam.NewCache()
			if err != nil {
				return err
			}
			if err := c.Clear(); err != nil {
				return err
			}
			cmd.Println("Local library cache cleared successfully.")
			return nil
		},
	}

	cacheCmd.AddCommand(statusCmd, clearCmd)
	return cacheCmd
}

func resolveCookie(cmd *cobra.Command, opts *cliOptions) (string, error) {
	// 1. Check manual cookie flag
	if opts.cookie != "" {
		return opts.cookie, nil
	}

	// 2. Check SHAZAM_COOKIE env var
	if envCookie := os.Getenv("SHAZAM_COOKIE"); envCookie != "" {
		return envCookie, nil
	}

	// 3. Check browser flag
	if opts.browserName != "" {
		cmd.PrintErrf("Extracting shazam.com session cookie from %s (%s profile)...\n", opts.browserName, opts.profile)
		cookie, err := browser.GetShazamCookieFromBrowser(opts.browserName, opts.profile)
		if err != nil {
			return "", fmt.Errorf("getting cookie from %s browser: %w", opts.browserName, err)
		}
		return cookie, nil
	}

	// 4. Try automatic browser extraction fallback
	cmd.PrintErrln("No cookie provided. Attempting automatic cookie lookup across default browsers...")
	cookie, err := browser.GetShazamCookieAuto()
	if err == nil {
		cmd.PrintErrln("Successfully retrieved decrypted cookie from browser profile!")
		return cookie, nil
	}

	return "", fmt.Errorf("no cookie specified and auto-extraction failed. Use -c, -b, or set SHAZAM_COOKIE environment variable")
}

func fetchOrLoadTracks(cmd *cobra.Command, opts *cliOptions) ([]shazam.Track, bool, error) {
	cache, err := shazam.NewCache()
	if err != nil {
		return nil, false, err
	}

	// 1. Try Cache First unless bypassed
	if !opts.noCache {
		tracks, err := cache.LoadTracks()
		if err == nil {
			return tracks, true, nil
		}
	}

	// 2. Resolve Cookie and fetch fresh from API
	cookie, err := resolveCookie(cmd, opts)
	if err != nil {
		return nil, false, err
	}

	client := shazam.NewClient()
	cmd.PrintErrln("Fetching library tracks from Shazam API...")

	tracks, err := client.FetchAllTracks(cmd.Context(), cookie, func(fetchedThisPage, totalFetched int) {
		cmd.PrintErrf("  • Retrieved %d tracks (Total: %d)...\n", fetchedThisPage, totalFetched)
	})
	if err != nil {
		return nil, false, fmt.Errorf("fetching tracks from Shazam: %w", err)
	}

	// 3. Save to cache unless disabled
	if !opts.noSave && len(tracks) > 0 {
		if err := cache.SaveTracks(tracks); err != nil {
			cmd.PrintErrf("Warning: failed to write fetched tracks to cache: %v\n", err)
		} else {
			cmd.PrintErrf("Saved %d tracks to local JSON cache.\n", len(tracks))
		}
	}

	return tracks, false, nil
}

func runList(cmd *cobra.Command, opts *cliOptions) error {
	tracks, isFromCache, err := fetchOrLoadTracks(cmd, opts)
	if err != nil {
		return err
	}

	if len(tracks) == 0 {
		cmd.Println("No tracks found in your Shazam library.")
		return nil
	}

	if isFromCache {
		cmd.PrintErrf("Using cached Shazam library (%d tracks, bypass with --no-cache)\n\n", len(tracks))
	} else {
		cmd.PrintErrf("Successfully retrieved %d tracks\n\n", len(tracks))
	}

	return printTracks(cmd.OutOrStdout(), tracks, opts)
}

func printTracks(out io.Writer, tracks []shazam.Track, opts *cliOptions) error {
	switch opts.output {
	case "query":
		for _, t := range tracks {
			q := fmt.Sprintf("%s - %s", t.Artist, t.Title)
			if opts.suffix != "" {
				q = fmt.Sprintf("%s %s", q, opts.suffix)
			}
			fmt.Fprintln(out, q)
		}
		return nil

	case "csv":
		writer := csv.NewWriter(out)
		_ = writer.Write([]string{"Index", "Date", "Title", "Artist", "URL"})
		for i, t := range tracks {
			_ = writer.Write([]string{
				fmt.Sprintf("%d", i+1),
				t.Timestamp.Format("2006-01-02 15:04:05"),
				t.Title,
				t.Artist,
				t.URL,
			})
		}
		writer.Flush()
		return nil

	case "list":
		for i, t := range tracks {
			dateStr := "N/A"
			if !t.Timestamp.IsZero() {
				dateStr = t.Timestamp.Format("2006-01-02")
			}
			fmt.Fprintf(out, "[%d] %s - %s (%s)\n", i+1, t.Title, t.Artist, dateStr)
		}
		return nil

	default: // "table"
		w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "INDEX\tDATE\tTITLE\tARTIST\tSHAZAM URL")
		fmt.Fprintln(w, "-----\t----\t-----\t------\t----------")
		for i, t := range tracks {
			dateStr := "N/A"
			if !t.Timestamp.IsZero() {
				dateStr = t.Timestamp.Format("2006-01-02")
			}
			title := t.Title
			if len(title) > 35 {
				title = title[:32] + "..."
			}
			artist := t.Artist
			if len(artist) > 25 {
				artist = artist[:22] + "..."
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", i+1, dateStr, title, artist, t.URL)
		}
		w.Flush()
		return nil
	}
}

func runExport(cmd *cobra.Command, opts *cliOptions, filePath string) error {
	tracks, _, err := fetchOrLoadTracks(cmd, opts)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("creating directories for export: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating export file: %w", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".json" {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(tracks); err != nil {
			return fmt.Errorf("exporting to JSON: %w", err)
		}
	} else {
		// Default to CSV
		writer := csv.NewWriter(file)
		_ = writer.Write([]string{"Index", "Date", "Title", "Artist", "URL", "CoverArt", "Key"})
		for i, t := range tracks {
			_ = writer.Write([]string{
				fmt.Sprintf("%d", i+1),
				t.Timestamp.Format("2006-01-02 15:04:05"),
				t.Title,
				t.Artist,
				t.URL,
				t.CoverArt,
				t.Key,
			})
		}
		writer.Flush()
	}

	cmd.Printf("Exported %d tracks to: %s\n", len(tracks), filePath)
	return nil
}

func main() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
