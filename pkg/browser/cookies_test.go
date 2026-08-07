package browser

import (
	"os"
	"strings"
	"testing"
)

func TestResolveCookieStorePath(t *testing.T) {
	// We will override GOOS for the resolution test
	originalGOOS := os.Getenv("GOOS")
	defer func() {
		_ = os.Setenv("GOOS", originalGOOS)
	}()

	tests := []struct {
		name        string
		goos        string
		browserName string
		profile     string
		expectSub   string
		expectErr   bool
	}{
		{
			name:        "macOS Chrome Default profile",
			goos:        "darwin",
			browserName: "chrome",
			profile:     "",
			expectSub:   "Library/Application Support/Google/Chrome/Default/Cookies",
			expectErr:   false,
		},
		{
			name:        "macOS Vivaldi Custom profile",
			goos:        "darwin",
			browserName: "vivaldi",
			profile:     "Profile 2",
			expectSub:   "Library/Application Support/Vivaldi/Profile 2/Cookies",
			expectErr:   false,
		},
		{
			name:        "macOS Safari",
			goos:        "darwin",
			browserName: "safari",
			profile:     "",
			expectSub:   "Library/Containers/com.apple.Safari/Data/Library/Cookies/Cookies.binarycookies",
			expectErr:   false,
		},
		{
			name:        "Windows Safari (unsupported)",
			goos:        "windows",
			browserName: "safari",
			profile:     "",
			expectSub:   "",
			expectErr:   true,
		},
		{
			name:        "Windows Chrome Default profile",
			goos:        "windows",
			browserName: "chrome",
			profile:     "Default",
			expectSub:   "Chrome\\User Data\\Default\\Cookies",
			expectErr:   false,
		},
		{
			name:        "Unsupported Browser",
			goos:        "darwin",
			browserName: "netscape",
			profile:     "",
			expectSub:   "",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("GOOS", tt.goos)
			path, err := ResolveCookieStorePath(tt.browserName, tt.profile)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %t, got error: %v", tt.expectErr, err)
			}

			if err == nil {
				// Normalize path separators to forward slash for comparison
				normalizedPath := strings.ReplaceAll(path, "\\", "/")
				normalizedExpectSub := strings.ReplaceAll(tt.expectSub, "\\", "/")
				if !strings.Contains(normalizedPath, normalizedExpectSub) {
					t.Errorf("expected path %q to contain %q", normalizedPath, normalizedExpectSub)
				}
			}
		})
	}
}
