package mailer_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"intraclub/mailer"
)

// TestSend_Integration actually delivers a message via MX lookup.
// Skipped unless MAILER_TEST_TO is set, so `go test ./...` stays offline-safe.
//
// Run with:
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

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
