package attendance

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The retention promise in docs/PRD.md section 8 is a legal commitment under
// UU PDP, not a nice-to-have. It gets a real test.
func TestPhotoStoreRetention(t *testing.T) {
	dir := t.TempDir()
	store := NewPhotoStore(dir, 30)

	now := time.Now()
	recent := store.Save("emp-1", now, []byte("jpeg-recent"))
	old := store.Save("emp-1", now.AddDate(0, 0, -45), []byte("jpeg-old"))
	edge := store.Save("emp-1", now.AddDate(0, 0, -29), []byte("jpeg-edge"))

	for _, p := range []string{recent, old, edge} {
		if p == "" {
			t.Fatal("Save returned an empty path")
		}
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("file not written: %v", err)
		}
	}

	store.sweep()

	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("a 45-day-old photo survived the sweep")
	}
	if _, err := os.Stat(recent); err != nil {
		t.Error("today's photo was deleted")
	}
	if _, err := os.Stat(edge); err != nil {
		t.Error("a 29-day-old photo was deleted before the 30-day window closed")
	}

	// Directories that are not ours must be left alone: the store may share a
	// volume, and deleting a stranger's data would be worse than keeping ours.
	foreign := filepath.Join(dir, "not-a-date")
	if err := os.MkdirAll(foreign, 0o750); err != nil {
		t.Fatal(err)
	}
	store.sweep()
	if _, err := os.Stat(foreign); err != nil {
		t.Error("sweep deleted a directory it does not own")
	}
}

// A nil store is the disabled configuration and must stay silent, not panic.
func TestPhotoStoreDisabled(t *testing.T) {
	var store *PhotoStore = NewPhotoStore("", 30)
	if store != nil {
		t.Fatal("empty dir should disable the store")
	}
	if path := store.Save("emp-1", time.Now(), []byte("x")); path != "" {
		t.Errorf("disabled store returned a path: %q", path)
	}
	store.StartCleanup() // must not panic
}
