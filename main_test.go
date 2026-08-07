package main

import "testing"

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
