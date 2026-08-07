package shazam

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Cache struct {
	cacheFilePath string
}

func NewCache() (*Cache, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting user home dir: %w", err)
		}
		cacheDir = filepath.Join(home, ".cache")
	}

	appCacheDir := filepath.Join(cacheDir, "shazam-cli")
	return &Cache{
		cacheFilePath: filepath.Join(appCacheDir, "library.json"),
	}, nil
}

func (c *Cache) GetCachePath() string {
	return c.cacheFilePath
}

func (c *Cache) SaveTracks(tracks []Track) error {
	dir := filepath.Dir(c.cacheFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	data, err := json.MarshalIndent(tracks, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing tracks to JSON: %w", err)
	}

	if err := os.WriteFile(c.cacheFilePath, data, 0644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	return nil
}

func (c *Cache) LoadTracks() ([]Track, error) {
	if _, err := os.Stat(c.cacheFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache file does not exist")
	}

	data, err := os.ReadFile(c.cacheFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading cache file: %w", err)
	}

	var tracks []Track
	if err := json.Unmarshal(data, &tracks); err != nil {
		return nil, fmt.Errorf("deserializing tracks from JSON: %w", err)
	}

	return tracks, nil
}

func (c *Cache) Clear() error {
	if _, err := os.Stat(c.cacheFilePath); os.IsNotExist(err) {
		return nil // Already cleared
	}
	return os.Remove(c.cacheFilePath)
}
