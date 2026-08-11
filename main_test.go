package main

import (
	"os"
	"testing"
	"time"

	"intraclub/api"
)

func TestResolveDBPath(t *testing.T) {
	// ensure env var is clean regardless of caller environment
	prev, hadPrev := os.LookupEnv("INTRACLUB_DB_PATH")
	t.Cleanup(func() {
		if hadPrev {
			os.Setenv("INTRACLUB_DB_PATH", prev)
		} else {
			os.Unsetenv("INTRACLUB_DB_PATH")
		}
	})

	t.Run("flag wins over env", func(t *testing.T) {
		os.Setenv("INTRACLUB_DB_PATH", "/from/env/data.db")
		if got := resolveDBPath("/from/flag/data.db"); got != "/from/flag/data.db" {
			t.Errorf("resolveDBPath = %q, want flag path", got)
		}
	})

	t.Run("falls back to env when flag empty", func(t *testing.T) {
		os.Setenv("INTRACLUB_DB_PATH", "/from/env/data.db")
		if got := resolveDBPath(""); got != "/from/env/data.db" {
			t.Errorf("resolveDBPath = %q, want env path", got)
		}
	})

	t.Run("empty when neither set", func(t *testing.T) {
		os.Unsetenv("INTRACLUB_DB_PATH")
		if got := resolveDBPath(""); got != "" {
			t.Errorf("resolveDBPath = %q, want empty", got)
		}
	})
}

func TestIsLoopbackAddress(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{"default bind", "127.0.0.1:8080", true},
		{"localhost host", "localhost:8080", true},
		{"bare loopback IP no port", "127.0.0.1", true},
		{"bare localhost no port", "localhost", true},
		{"any-interface port", ":8080", false},
		{"all interfaces", "0.0.0.0:8080", false},
		{"wildcard", "0.0.0.0", false},
		{"lan ip", "192.168.1.10:8080", false},
		{"non-loopback hostname", "example.com:8080", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLoopbackAddress(tt.addr); got != tt.want {
				t.Errorf("isLoopbackAddress(%q) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}

func TestResolveJwtLifetimeDefaultsToPackageDefault(t *testing.T) {
	// No flag, no env var -> package default (2h).
	api.JwtLifetime = time.Hour * 2
	t.Setenv("INTRACLUB_JWT_LIFETIME", "")

	d, err := resolveJwtLifetime("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != time.Hour*2 {
		t.Fatalf("expected default 2h, got %v", d)
	}
}

func TestResolveJwtLifetimeEnvVar(t *testing.T) {
	t.Setenv("INTRACLUB_JWT_LIFETIME", "90m")

	d, err := resolveJwtLifetime("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 90*time.Minute {
		t.Fatalf("expected 90m, got %v", d)
	}
}

func TestResolveJwtLifetimeFlagWins(t *testing.T) {
	t.Setenv("INTRACLUB_JWT_LIFETIME", "90m")

	d, err := resolveJwtLifetime("5s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 5*time.Second {
		t.Fatalf("expected 5s, got %v", d)
	}
}

func TestResolveJwtLifetimeInvalid(t *testing.T) {
	t.Setenv("INTRACLUB_JWT_LIFETIME", "not-a-duration")

	_, err := resolveJwtLifetime("")
	if err == nil {
		t.Fatal("expected an error for an invalid lifetime")
	}
}

func TestResolveJwtLifetimeRejectsNonPositive(t *testing.T) {
	for _, raw := range []string{"0s", "-5s", "-1h"} {
		t.Run(raw, func(t *testing.T) {
			t.Setenv("INTRACLUB_JWT_LIFETIME", raw)
			if _, err := resolveJwtLifetime(""); err == nil {
				t.Fatalf("expected an error for lifetime %q", raw)
			}
		})
	}
}
