// internal/cache/cache.go
package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
)

// Cache provides both HTTP caching and general data caching
type Cache struct {
	TTL           time.Duration
	BaseDir       string
	HTTPCache     httpcache.Cache
	HTTPTransport *httpcache.Transport
}

// NewCache creates a new cache with the specified TTL
func NewCache(ttl time.Duration) *Cache {
	// Determine base cache directory using XDG standard
	baseDir := os.Getenv("XDG_CACHE_HOME")
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".cache")
	}

	// Create HTTP cache directory
	httpCacheDir := filepath.Join(baseDir, "http-cache")
	os.MkdirAll(httpCacheDir, 0755)

	// Create disk cache for HTTP
	diskCache := diskcache.New(httpCacheDir)

	// Create transport with caching
	transport := httpcache.NewTransport(diskCache)
	transport.MarkCachedResponses = true

	return &Cache{
		TTL:           ttl,
		BaseDir:       baseDir,
		HTTPCache:     diskCache,
		HTTPTransport: transport,
	}
}

// GetHTTPClient returns an HTTP client with caching enabled
func (c *Cache) GetHTTPClient() *http.Client {
	if c == nil {
		return nil
	}
	return &http.Client{
		Transport: c.HTTPTransport,
		Timeout:   30 * time.Second,
	}
}

// GetHTTPClientWithTransport wraps an existing transport with caching
func (c *Cache) GetHTTPClientWithTransport(baseTransport http.RoundTripper) *http.Client {
	if c == nil {
		return nil
	}
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	// Wrap the base transport
	c.HTTPTransport.Transport = baseTransport

	return &http.Client{
		Transport: c.HTTPTransport,
		Timeout:   30 * time.Second,
	}
}

// ClearHTTPCache clears the HTTP cache
func (c *Cache) ClearHTTPCache() {
	if c == nil {
		return
	}
	if _, ok := c.HTTPCache.(*diskcache.Cache); ok {
		// Clear all entries
		// Note: diskcache doesn't have a Clear method, so we recreate it
		httpCacheDir := filepath.Join(c.BaseDir, "http-cache")
		os.RemoveAll(httpCacheDir)
		os.MkdirAll(httpCacheDir, 0755)
		c.HTTPCache = diskcache.New(httpCacheDir)
		c.HTTPTransport.Cache = c.HTTPCache
	}
}

// GetCacheDir returns the cache directory for a specific app
func (c *Cache) GetCacheDir(appName string) string {
	if c == nil {
		return ""
	}
	dir := filepath.Join(c.BaseDir, appName)
	os.MkdirAll(dir, 0755)
	return dir
}

// --- JSON Data Caching (for non-HTTP data) ---

// Save saves data to cache
func (c *Cache) Save(key string, data any) error {
	return c.SaveTo(key, data, "")
}

// SaveTo saves data to a specific app's cache
func (c *Cache) SaveTo(key string, data any, appName string) error {
	if c == nil {
		return fmt.Errorf("cache is nil")
	}
	dir := c.BaseDir
	if appName != "" {
		dir = c.GetCacheDir(appName)
	}

	// Hash the key if it's too long or contains invalid characters
	safeKey := c.sanitizeKey(key)
	path := filepath.Join(dir, safeKey+".json")

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create wrapper with timestamp
	wrapper := struct {
		Timestamp time.Time `json:"timestamp"`
		Data      any       `json:"data"`
		Key       string    `json:"original_key"` // Store original key for reference
	}{
		Timestamp: time.Now(),
		Data:      data,
		Key:       key,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(wrapper)
}

// Load loads data from cache
func (c *Cache) Load(key string, target any) bool {
	return c.LoadFrom(key, target, "")
}

// LoadFrom loads data from a specific app's cache
func (c *Cache) LoadFrom(key string, target any, appName string) bool {
	if c == nil {
		return false
	}
	dir := c.BaseDir
	if appName != "" {
		dir = c.GetCacheDir(appName)
	}

	safeKey := c.sanitizeKey(key)
	path := filepath.Join(dir, safeKey+".json")

	// Check if file exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}

	// Check if cache is expired
	if time.Since(info.ModTime()) > c.TTL {
		return false
	}

	// Load file
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Decode wrapper
	var wrapper struct {
		Timestamp time.Time       `json:"timestamp"`
		Data      json.RawMessage `json:"data"`
		Key       string          `json:"original_key"`
	}

	if err := json.NewDecoder(file).Decode(&wrapper); err != nil {
		return false
	}

	// Check timestamp-based expiry
	if time.Since(wrapper.Timestamp) > c.TTL {
		return false
	}

	// Decode actual data
	if err := json.Unmarshal(wrapper.Data, target); err != nil {
		return false
	}

	return true
}

// Clear removes all cached files for an app
func (c *Cache) Clear(appName string) error {
	if c == nil {
		return fmt.Errorf("cache is nil")
	}
	dir := c.GetCacheDir(appName)

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".json" || filepath.Ext(path) == ".torrent" {
			return os.Remove(path)
		}
		return nil
	})
}

// ClearKey removes a specific cached item
func (c *Cache) ClearKey(key string, appName string) error {
	if c == nil {
		return fmt.Errorf("cache is nil")
	}
	dir := c.BaseDir
	if appName != "" {
		dir = c.GetCacheDir(appName)
	}

	safeKey := c.sanitizeKey(key)
	path := filepath.Join(dir, safeKey+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Also try to remove .torrent files with same key
	torrentPath := filepath.Join(dir, safeKey+".torrent")
	if err := os.Remove(torrentPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// SetTTL updates the TTL for future cache operations
func (c *Cache) SetTTL(ttl time.Duration) {
	c.TTL = ttl
}

// IsExpired checks if a cached item is expired without loading it
func (c *Cache) IsExpired(key string, appName string) bool {
	if c == nil {
		return true
	}
	dir := c.BaseDir
	if appName != "" {
		dir = c.GetCacheDir(appName)
	}

	safeKey := c.sanitizeKey(key)
	path := filepath.Join(dir, safeKey+".json")

	info, err := os.Stat(path)
	if err != nil {
		return true
	}

	return time.Since(info.ModTime()) > c.TTL
}

// GetAge returns how old a cached item is
func (c *Cache) GetAge(key string, appName string) (time.Duration, error) {
	if c == nil {
		return 0, fmt.Errorf("cache is nil")
	}
	dir := c.BaseDir
	if appName != "" {
		dir = c.GetCacheDir(appName)
	}

	safeKey := c.sanitizeKey(key)
	path := filepath.Join(dir, safeKey+".json")

	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("cache item not found: %w", err)
	}

	return time.Since(info.ModTime()), nil
}

// sanitizeKey creates a safe filename from a cache key
func (c *Cache) sanitizeKey(key string) string {
	// If key is already safe, use it directly
	if len(key) < 200 && isValidFilename(key) {
		return key
	}

	// Otherwise, hash it
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

// isValidFilename checks if a string is safe to use as a filename
func isValidFilename(s string) bool {
	// Check for invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r"}
	for _, char := range invalid {
		if strings.Contains(s, char) {
			return false
		}
	}
	return true
}
