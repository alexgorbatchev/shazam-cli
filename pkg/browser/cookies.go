package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/browser/chrome"
)

// Default paths for Chrome and Vivaldi on different operating systems.
func ResolveCookieStorePath(browserName, profile string) (string, error) {
	if profile == "" {
		profile = "Default"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	configDir, _ := os.UserConfigDir()

	var relPath string
	browserName = strings.ToLower(browserName)

	switch os.Getenv("GOOS") {
	case "windows":
		// Windows paths
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		if browserName == "chrome" || browserName == "google-chrome" {
			relPath = filepath.Join(localAppData, "Google", "Chrome", "User Data", profile, "Cookies")
		} else if browserName == "vivaldi" {
			relPath = filepath.Join(localAppData, "Vivaldi", "User Data", profile, "Cookies")
		} else {
			return "", fmt.Errorf("unsupported browser for auto-resolution on Windows: %s", browserName)
		}
	case "darwin":
		// macOS paths
		if browserName == "chrome" || browserName == "google-chrome" {
			relPath = filepath.Join(home, "Library", "Application Support", "Google", "Chrome", profile, "Cookies")
		} else if browserName == "vivaldi" {
			relPath = filepath.Join(home, "Library", "Application Support", "Vivaldi", profile, "Cookies")
		} else {
			return "", fmt.Errorf("unsupported browser for auto-resolution on macOS: %s", browserName)
		}
	default:
		// Linux/other Unix paths
		if browserName == "chrome" || browserName == "google-chrome" {
			relPath = filepath.Join(configDir, "google-chrome", profile, "Cookies")
			if _, err := os.Stat(relPath); os.IsNotExist(err) {
				// Fallback to chromium path
				relPath = filepath.Join(configDir, "chromium", profile, "Cookies")
			}
		} else if browserName == "vivaldi" {
			relPath = filepath.Join(configDir, "vivaldi", profile, "Cookies")
		} else {
			return "", fmt.Errorf("unsupported browser for auto-resolution on Linux: %s", browserName)
		}
	}

	return relPath, nil
}

// GetShazamCookieFromBrowser extracts and decrypts the cookies for shazam.com from a browser's Cookie database.
func GetShazamCookieFromBrowser(browserName, profile string) (string, error) {
	path, err := ResolveCookieStorePath(browserName, profile)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("browser cookies file does not exist at: %s", path)
	}

	// Since both Chrome and Vivaldi are Chromium-based, they use the same cookie store schema
	// and system keychain/DPAPI decryption method.
	cookieStore := chrome.TraverseCookies(path)
	
	var shazamCookies []string
	for cookie := range cookieStore.OnlyCookies() {
		// Filter for shazam.com
		if strings.HasSuffix(cookie.Domain, "shazam.com") {
			shazamCookies = append(shazamCookies, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
		}
	}

	if len(shazamCookies) == 0 {
		return "", fmt.Errorf("no shazam.com cookies found in browser store at %s (are you logged into shazam.com in that browser?)", path)
	}

	return strings.Join(shazamCookies, "; "), nil
}

// GetShazamCookieAuto attempts to automatically find Shazam cookies across all known default browser stores using kooky's auto-find capability.
func GetShazamCookieAuto() (string, error) {
	// Query using kooky's traverse with filters
	cookiesSeq := kooky.TraverseCookies(
		context.Background(),
		kooky.Valid,
		kooky.DomainHasSuffix("shazam.com"),
	).OnlyCookies()

	var shazamCookies []string
	for cookie := range cookiesSeq {
		shazamCookies = append(shazamCookies, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}

	if len(shazamCookies) == 0 {
		return "", fmt.Errorf("could not automatically locate any decrypted shazam.com cookies")
	}

	return strings.Join(shazamCookies, "; "), nil
}
