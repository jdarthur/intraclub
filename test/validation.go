package test

import (
	"fmt"
	"strings"
	"testing"
)

func ValidateErrorContains(t *testing.T, err error, expectedSubstring string) {
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Error \"%s\" did not contain expected substring \"%s\"", err.Error(), expectedSubstring)
		return
	}

	fmt.Printf("Correct error: \"%s\"\n", err.Error())
}
