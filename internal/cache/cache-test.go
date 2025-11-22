// internal/cache/cache_test.go
package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_SaveAndLoad(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	cache.BaseDir = t.TempDir() // Use temp dir for testing

	type testData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// Test saving
	original := testData{Name: "test", Value: 42}
	err := cache.SaveTo("test_key", original, "test_app")
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Test loading
	var loaded testData
	if !cache.LoadFrom("test_key", &loaded, "test_app") {
		t.Fatal("failed to load cached data")
	}

	if loaded.Name != original.Name || loaded.Value != original.Value {
		t.Errorf("loaded data doesn't match: got %+v, want %+v", loaded, original)
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(100 * time.Millisecond) // Very short TTL
	cache.BaseDir = t.TempDir()

	// Save data
	data := map[string]string{"key": "value"}
	err := cache.Save("expiring_key", data)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Should load immediately
	var loaded map[string]string
	if !cache.Load("expiring_key", &loaded) {
		t.Fatal("should have loaded fresh data")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not load after expiration
	if cache.Load("expiring_key", &loaded) {
		t.Error("should not have loaded expired data")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	cache.BaseDir = t.TempDir()

	// Save multiple items
	cache.SaveTo("key1", "value1", "app1")
	cache.SaveTo("key2", "value2", "app1")
	cache.SaveTo("key3", "value3", "app2")

	// Clear app1
	err := cache.Clear("app1")
	if err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	// app1 items should be gone
	var val string
	if cache.LoadFrom("key1", &val, "app1") {
		t.Error("key1 should have been cleared")
	}
	if cache.LoadFrom("key2", &val, "app1") {
		t.Error("key2 should have been cleared")
	}

	// app2 items should remain
	if !cache.LoadFrom("key3", &val, "app2") {
		t.Error("key3 should still exist")
	}
}

func TestCache_ClearKey(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	cache.BaseDir = t.TempDir()

	// Save items
	cache.SaveTo("key1", "value1", "app1")
	cache.SaveTo("key2", "value2", "app1")

	// Clear specific key
	err := cache.ClearKey("key1", "app1")
	if err != nil {
		t.Fatalf("failed to clear key: %v", err)
	}

	// key1 should be gone
	var val string
	if cache.LoadFrom("key1", &val, "app1") {
		t.Error("key1 should have been cleared")
	}

	// key2 should remain
	if !cache.LoadFrom("key2", &val, "app1") {
		t.Error("key2 should still exist")
	}
}

func TestCache_IsExpired(t *testing.T) {
	cache := NewCache(200 * time.Millisecond)
	cache.BaseDir = t.TempDir()

	// Save data
	err := cache.SaveTo("test_key", "data", "app")
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Should not be expired immediately
	if cache.IsExpired("test_key", "app") {
		t.Error("fresh cache should not be expired")
	}

	// Wait for expiration
	time.Sleep(250 * time.Millisecond)

	// Should be expired now
	if !cache.IsExpired("test_key", "app") {
		t.Error("cache should be expired")
	}
}

func TestCache_GetAge(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	cache.BaseDir = t.TempDir()

	// Save data
	err := cache.SaveTo("test_key", "data", "app")
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Check age
	age, err := cache.GetAge("test_key", "app")
	if err != nil {
		t.Fatalf("failed to get age: %v", err)
	}

	if age < 100*time.Millisecond {
		t.Errorf("age should be at least 100ms, got %v", age)
	}
	if age > 200*time.Millisecond {
		t.Errorf("age should be less than 200ms, got %v", age)
	}
}

func TestCache_GetCacheDir(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	cache.BaseDir = t.TempDir()

	dir := cache.GetCacheDir("test_app")
	expectedDir := filepath.Join(cache.BaseDir, "test_app")

	if dir != expectedDir {
		t.Errorf("expected cache dir %s, got %s", expectedDir, dir)
	}

	// Directory should be created
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("cache dir should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("cache dir should be a directory")
	}
}

func TestCache_NonExistentKey(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	cache.BaseDir = t.TempDir()

	var data string
	if cache.Load("nonexistent", &data) {
		t.Error("should not load non-existent key")
	}
}

func TestCache_SetTTL(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	cache.BaseDir = t.TempDir()

	// Save with initial TTL
	err := cache.Save("key", "value")
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Change TTL to very short
	cache.SetTTL(100 * time.Millisecond)

	// Wait for new TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should not load with new TTL
	var val string
	if cache.Load("key", &val) {
		t.Error("should not load after TTL change and expiration")
	}
}