package auth

import (
	"strings"
	"sync"
	"time"
)

// throttle counts FAILED login attempts per account.
//
// Keying this by IP instead would be wrong for this product: every employee in
// the office shares one NAT address, so an IP limit locks out the whole company
// at 08:00 while doing nothing to slow an attacker who rotates IPs. The thing
// worth protecting is the account.
type throttle struct {
	mu       sync.Mutex
	attempts map[string]*attemptRecord
	limit    int
	window   time.Duration
}

type attemptRecord struct {
	count int
	reset time.Time
}

func newThrottle(limit int, window time.Duration) *throttle {
	t := &throttle{attempts: map[string]*attemptRecord{}, limit: limit, window: window}
	go func() {
		for range time.Tick(window) {
			t.mu.Lock()
			now := time.Now()
			for k, r := range t.attempts {
				if now.After(r.reset) {
					delete(t.attempts, k)
				}
			}
			t.mu.Unlock()
		}
	}()
	return t
}

func (t *throttle) key(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// blocked reports whether this account has burned through its failed attempts.
func (t *throttle) blocked(email string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	r, ok := t.attempts[t.key(email)]
	if !ok || time.Now().After(r.reset) {
		return false
	}
	return r.count >= t.limit
}

func (t *throttle) recordFailure(email string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	k := t.key(email)
	r, ok := t.attempts[k]
	if !ok || time.Now().After(r.reset) {
		r = &attemptRecord{reset: time.Now().Add(t.window)}
		t.attempts[k] = r
	}
	r.count++
}

// reset clears the counter after a successful login, so a user who mistyped
// twice and then got it right starts clean.
func (t *throttle) reset(email string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, t.key(email))
}
