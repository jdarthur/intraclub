package main

import (
	"os"
	"testing"
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
