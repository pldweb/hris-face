package attendance

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// PhotoStore keeps the frame each attendance was verified from (docs/PRD.md 8:
// optional, retained 30 days, then deleted).
//
// Disabled by default. Storing a face photo of every employee twice a day is a
// meaningful privacy cost under UU PDP, so it is opt-in and self-expiring
// rather than something a deployment accumulates by accident.
type PhotoStore struct {
	dir       string
	retention time.Duration
}

func NewPhotoStore(dir string, retentionDays int) *PhotoStore {
	if dir == "" {
		return nil
	}
	if retentionDays <= 0 {
		retentionDays = 30
	}
	return &PhotoStore{dir: dir, retention: time.Duration(retentionDays) * 24 * time.Hour}
}

// Save writes the JPEG and returns the path recorded on the attendance row.
// A failure here must not fail the check-in: the attendance is the record that
// matters, the photo is supporting evidence.
func (p *PhotoStore) Save(employeeID string, at time.Time, jpeg []byte) string {
	if p == nil {
		return ""
	}
	day := at.Format("2006-01-02")
	dir := filepath.Join(p.dir, day)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		log.Printf("photostore: mkdir %s: %v", dir, err)
		return ""
	}
	name := employeeID + "-" + at.Format("150405.000") + ".jpg"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, jpeg, 0o640); err != nil {
		log.Printf("photostore: write %s: %v", path, err)
		return ""
	}
	return path
}

// StartCleanup removes day-directories older than the retention window. Runs
// on a ticker rather than as a cron entry so the retention promise travels with
// the binary instead of depending on a deploy step someone can forget.
func (p *PhotoStore) StartCleanup() {
	if p == nil {
		return
	}
	go func() {
		p.sweep()
		for range time.Tick(6 * time.Hour) {
			p.sweep()
		}
	}()
}

func (p *PhotoStore) sweep() {
	entries, err := os.ReadDir(p.dir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("photostore: baca %s: %v", p.dir, err)
		}
		return
	}

	cutoff := time.Now().Add(-p.retention)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		day, err := time.Parse("2006-01-02", e.Name())
		if err != nil {
			continue // not one of ours; leave it alone
		}
		if day.Before(cutoff) {
			path := filepath.Join(p.dir, e.Name())
			if err := os.RemoveAll(path); err != nil {
				log.Printf("photostore: hapus %s: %v", path, err)
				continue
			}
			log.Printf("photostore: retensi -- %s dihapus", e.Name())
		}
	}
}
