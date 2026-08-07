package shazam

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultBaseURL = "https://www.shazam.com/services/profile/v1/en-US/US/web/profile"
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

type ClientOption func(*Client)

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    DefaultBaseURL,
		userAgent:  DefaultUserAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// FetchLibraryPage fetches a single page of Shazam tracks.
func (c *Client) FetchLibraryPage(ctx context.Context, cookie string, limit, offset int) ([]Track, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}

	q := u.Query()
	q.Set("limit", fmt.Sprintf("%d", limit))
	q.Set("offset", fmt.Sprintf("%d", offset))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Cookie", cookie)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-200 status: %d (body: %s)", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return ParseTracksFromJSON(bodyBytes)
}

// FetchAllTracks automatically paginates and fetches all tracks in the user's Shazam library.
func (c *Client) FetchAllTracks(ctx context.Context, cookie string, onPageProgress func(fetchedThisPage, totalFetched int)) ([]Track, error) {
	var allTracks []Track
	limit := 100
	offset := 0

	for {
		pageTracks, err := c.FetchLibraryPage(ctx, cookie, limit, offset)
		if err != nil {
			// If we already fetched some tracks, we can return what we have, or bubble up the error.
			if len(allTracks) > 0 {
				return allTracks, fmt.Errorf("error fetching page at offset %d (returning %d tracks): %w", offset, len(allTracks), err)
			}
			return nil, err
		}

		if len(pageTracks) == 0 {
			break
		}

		allTracks = append(allTracks, pageTracks...)
		if onPageProgress != nil {
			onPageProgress(len(pageTracks), len(allTracks))
		}

		// Pagination logic
		offset += len(pageTracks)
		
		// If we got fewer tracks than the limit, we've hit the end of the list.
		if len(pageTracks) < limit {
			break
		}

		// Pause briefly to be nice to Shazam's servers
		time.Sleep(300 * time.Millisecond)
	}

	return allTracks, nil
}
