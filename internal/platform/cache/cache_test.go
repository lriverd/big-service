package cache_test

import (
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/platform/cache"
)

func TestCache_SetAndGet(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)

	c.Set("key1", "value1", 5*time.Minute)
	val, found := c.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if val.(string) != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestCache_GetMiss(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)

	_, found := c.Get("nonexistent")
	if found {
		t.Error("expected not found")
	}
}

func TestCache_Delete(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)

	c.Set("key1", "value1", 5*time.Minute)
	c.Delete("key1")
	_, found := c.Get("key1")
	if found {
		t.Error("expected key1 to be deleted")
	}
}

func TestCache_DeleteByPrefix(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)

	c.Set("species:list:1", "a", 5*time.Minute)
	c.Set("species:list:2", "b", 5*time.Minute)
	c.Set("species:123", "c", 5*time.Minute)
	c.Set("other:key", "d", 5*time.Minute)

	c.DeleteByPrefix("species:list")

	if _, found := c.Get("species:list:1"); found {
		t.Error("expected species:list:1 deleted")
	}
	if _, found := c.Get("species:list:2"); found {
		t.Error("expected species:list:2 deleted")
	}
	if _, found := c.Get("species:123"); !found {
		t.Error("expected species:123 to remain")
	}
	if _, found := c.Get("other:key"); !found {
		t.Error("expected other:key to remain")
	}
}

func TestCache_Flush(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)

	c.Set("key1", "a", 5*time.Minute)
	c.Set("key2", "b", 5*time.Minute)
	c.Flush()

	if c.ItemCount() != 0 {
		t.Errorf("expected 0 items, got %d", c.ItemCount())
	}
}

func TestCache_ItemCount(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)

	c.Set("a", 1, 5*time.Minute)
	c.Set("b", 2, 5*time.Minute)

	if c.ItemCount() != 2 {
		t.Errorf("expected 2, got %d", c.ItemCount())
	}
}

func TestCache_Expiration(t *testing.T) {
	c := cache.New(50*time.Millisecond, 100*time.Millisecond)

	c.Set("key1", "value1", 50*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	_, found := c.Get("key1")
	if found {
		t.Error("expected key to be expired")
	}
}

