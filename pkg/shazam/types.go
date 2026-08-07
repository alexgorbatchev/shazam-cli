package shazam

import (
	"encoding/json"
	"fmt"
	"time"
)

// Track represents a unified song record from Shazam.
type Track struct {
	Key       string    `json:"key"`
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	CoverArt  string    `json:"cover_art,omitempty"`
}

// ProfileResponse is our structured target for Shazam profile endpoint responses.
type ProfileResponse struct {
	User struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"user"`
	Tags []RawTag `json:"tags"`
}

// RawTag is a tag entry returned by the private web profile/tags endpoint.
type RawTag struct {
	Key       string `json:"key"`
	Timestamp int64  `json:"timestamp"` // epoch ms or sec
	Track     struct {
		Key      string            `json:"key"`
		Title    string            `json:"title"`
		Subtitle string            `json:"subtitle"` // Artist name is often here
		URL      string            `json:"url"`
		Images   map[string]string `json:"images"`
	} `json:"track"`
}

// ParseTracksFromJSON parses the JSON bytes of a profile response in a highly resilient way.
// If the structured ProfileResponse doesn't yield tracks, it recursively hunts for arrays of tracks.
func ParseTracksFromJSON(data []byte) ([]Track, error) {
	// 1. Try structured unmarshaling first
	var resp ProfileResponse
	if err := json.Unmarshal(data, &resp); err == nil && len(resp.Tags) > 0 {
		var tracks []Track
		for _, tag := range resp.Tags {
			t := Track{
				Key:    tag.Key,
				Title:  tag.Track.Title,
				Artist: tag.Track.Subtitle,
				URL:    tag.Track.URL,
			}
			if tag.Timestamp > 0 {
				// Handle both seconds and milliseconds
				if tag.Timestamp > 1e11 {
					t.Timestamp = time.UnixMilli(tag.Timestamp)
				} else {
					t.Timestamp = time.Unix(tag.Timestamp, 0)
				}
			}
			if tag.Track.Images != nil {
				if cover, ok := tag.Track.Images["coverart"]; ok {
					t.CoverArt = cover
				} else if cover, ok := tag.Track.Images["default"]; ok {
					t.CoverArt = cover
				}
			}
			if t.Title != "" {
				tracks = append(tracks, t)
			}
		}
		if len(tracks) > 0 {
			return tracks, nil
		}
	}

	// 2. Resilient fallback: parse into generic map and scan for tag arrays
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		// Try parsing as a flat list of items/tags
		var list []interface{}
		if errList := json.Unmarshal(data, &list); errList == nil {
			return parseRawList(list), nil
		}
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	foundList := findArrayInMap(raw, []string{"tags", "tracks", "library", "history", "items"})
	if foundList != nil {
		return parseRawList(foundList), nil
	}

	return nil, fmt.Errorf("no tracks found in JSON response")
}

func findArrayInMap(m map[string]interface{}, targetKeys []string) []interface{} {
	for _, tk := range targetKeys {
		if val, ok := m[tk]; ok {
			if arr, ok := val.([]interface{}); ok {
				return arr
			}
		}
	}
	// Recurse deeper if needed
	for _, val := range m {
		if nestedMap, ok := val.(map[string]interface{}); ok {
			if arr := findArrayInMap(nestedMap, targetKeys); arr != nil {
				return arr
			}
		} else if nestedArr, ok := val.([]interface{}); ok {
			for _, item := range nestedArr {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if arr := findArrayInMap(itemMap, targetKeys); arr != nil {
						return arr
					}
				}
			}
		}
	}
	return nil
}

func parseRawList(list []interface{}) []Track {
	var tracks []Track
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		t := Track{}
		// Look for key/id
		if key, ok := m["key"].(string); ok {
			t.Key = key
		} else if id, ok := m["id"].(string); ok {
			t.Key = id
		}

		// Look for timestamp
		if ts, ok := m["timestamp"].(float64); ok {
			if ts > 1e11 {
				t.Timestamp = time.UnixMilli(int64(ts))
			} else {
				t.Timestamp = time.Unix(int64(ts), 0)
			}
		} else if tsStr, ok := m["timestamp"].(string); ok {
			// Parse string format if needed
			if parsed, err := time.Parse(time.RFC3339, tsStr); err == nil {
				t.Timestamp = parsed
			}
		}

		// Look for inner track details or extract directly
		trackObj := m
		if innerTrack, ok := m["track"].(map[string]interface{}); ok {
			trackObj = innerTrack
		}

		if title, ok := trackObj["title"].(string); ok {
			t.Title = title
		} else if name, ok := trackObj["name"].(string); ok {
			t.Title = name
		}

		if subtitle, ok := trackObj["subtitle"].(string); ok {
			t.Artist = subtitle
		} else if artist, ok := trackObj["artist"].(string); ok {
			t.Artist = artist
		} else if artists, ok := trackObj["artists"].([]interface{}); ok && len(artists) > 0 {
			if artMap, ok := artists[0].(map[string]interface{}); ok {
				if name, ok := artMap["name"].(string); ok {
					t.Artist = name
				}
			}
		}

		if url, ok := trackObj["url"].(string); ok {
			t.URL = url
		}

		if images, ok := trackObj["images"].(map[string]interface{}); ok {
			if cover, ok := images["coverart"].(string); ok {
				t.CoverArt = cover
			} else if def, ok := images["default"].(string); ok {
				t.CoverArt = def
			}
		}

		if t.Title != "" {
			tracks = append(tracks, t)
		}
	}
	return tracks
}
