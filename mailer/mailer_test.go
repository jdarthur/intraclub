package mailer_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"intraclub/mailer"
)

// TestSend_Integration performs a real end-to-end delivery: MX lookup,
// SMTP handshake, optional STARTTLS upgrade, DKIM signing, and message
// submission. It is skipped by default so that `go test ./...` remains
// safe to run without network access or credentials.
//
// Required environment variables:
//
//	MAILER_TEST_TO        Comma-separated recipient addresses (e.g. "me@gmail.com")
//	MAILER_DKIM_KEY       Path to the PEM-encoded DKIM private key
//	MAILER_FROM_DOMAIN    Sending domain (default: "rcintra.club")
//	MAILER_HOSTNAME       HELO hostname (default: "mail.rcintra.club")
//	MAILER_DKIM_SELECTOR  DKIM DNS selector (default: "default")
//
// Example:
//
//	MAILER_TEST_TO=you@gmail.com \
//	MAILER_DKIM_KEY=/etc/intraclub/dkim.key \
//	MAILER_FROM_DOMAIN=rcintra.club \
//	MAILER_HOSTNAME=mail.rcintra.club \
//	MAILER_DKIM_SELECTOR=default \
//	  go test ./mailer -run Integration -v
func TestSend_Integration(t *testing.T) {
	to := os.Getenv("MAILER_TEST_TO")
	if to == "" {
		t.Skip("set MAILER_TEST_TO to run this test")
	}

	m, err := mailer.New(mailer.Config{
		FromDomain:   envOr("MAILER_FROM_DOMAIN", "rcintra.club"),
		Hostname:     envOr("MAILER_HOSTNAME", "mail.rcintra.club"),
		DKIMSelector: envOr("MAILER_DKIM_SELECTOR", "default"),
		DKIMKeyPath:  os.Getenv("MAILER_DKIM_KEY"),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	from := "noreply@" + envOr("MAILER_FROM_DOMAIN", "rcintra.club")
	err = m.Send(ctx, mailer.Message{
		From:    from,
		To:      strings.Split(to, ","),
		Subject: "intraclub mailer test " + time.Now().Format(time.RFC3339),
		Text:    "If you can read this, MX lookup, STARTTLS, and DKIM signing all worked.\n",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
}

// envOr returns the value of environment variable k, or def if k is unset
// or empty.
func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
