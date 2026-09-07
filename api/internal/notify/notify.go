// Package notify sends operational email (docs/PRD.md F8).
//
// When SMTP is not configured the sender logs instead of failing: a missing
// mail server must never block an attendance correction from being filed.
package notify

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type Sender struct {
	cfg Config
}

func New(cfg Config) *Sender { return &Sender{cfg: cfg} }

func (s *Sender) enabled() bool {
	return s.cfg.Host != "" && s.cfg.From != ""
}

// Send delivers one message. Errors are returned but callers are expected to
// treat them as non-fatal; the caller's own work has already been committed.
func (s *Sender) Send(to []string, subject, body string) error {
	if len(to) == 0 {
		return nil
	}
	if !s.enabled() {
		log.Printf("notify (SMTP nonaktif) -> %s: %s", strings.Join(to, ","), subject)
		return nil
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		s.cfg.From, strings.Join(to, ", "), subject, body)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}
	addr := s.cfg.Host + ":" + s.cfg.Port
	if err := smtp.SendMail(addr, auth, s.cfg.From, to, []byte(msg)); err != nil {
		log.Printf("notify gagal kirim ke %s: %v", strings.Join(to, ","), err)
		return err
	}
	return nil
}
