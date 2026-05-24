package store

import (
	"testing"
	"time"
)

func TestMarkMessageSeenDedupe(t *testing.T) {
	store := NewMemoryStore("salt")
	if duplicate := store.MarkMessageSeen("wamid-1", time.Hour); duplicate {
		t.Fatal("first message should not be duplicate")
	}
	if duplicate := store.MarkMessageSeen("wamid-1", time.Hour); !duplicate {
		t.Fatal("second message should be duplicate")
	}
}

func TestAllowUserEventRateLimit(t *testing.T) {
	store := NewMemoryStore("salt")
	if !store.AllowUserEvent("user-1", 2, time.Minute) {
		t.Fatal("first event should be allowed")
	}
	if !store.AllowUserEvent("user-1", 2, time.Minute) {
		t.Fatal("second event should be allowed")
	}
	if store.AllowUserEvent("user-1", 2, time.Minute) {
		t.Fatal("third event should be rate limited")
	}
}
