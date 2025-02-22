package model

import (
	"fmt"
	"testing"
)

func TestEmailStartsWithAtSign(t *testing.T) {
	email := EmailAddress("@google.com")
	err := email.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Println(err)
}

func TestEmailStartsWithDash(t *testing.T) {
	email := EmailAddress("-@google.com")
	err := email.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Println(err)
}

func TestEmailStartsWithUnderscore(t *testing.T) {
	email := EmailAddress("_@google.com")
	err := email.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Println(err)
}

func TestEmailStartsWithDot(t *testing.T) {
	email := EmailAddress(".@google.com")
	err := email.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Println(err)
}

func TestEmailContainsInvalidCharacter(t *testing.T) {
	email := EmailAddress("email!!!@google.com")
	err := email.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Println(err)
}

func TestEmailContainsTwoAtSigns(t *testing.T) {
	email := EmailAddress("email@@google.com")
	err := email.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Println(err)
}

func TestEmailContainsNoAtSigns(t *testing.T) {
	email := EmailAddress("google.com")
	err := email.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Println(err)
}

func TestEmailEndsWithAtSign(t *testing.T) {
	email := EmailAddress("google.com@")
	err := email.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Println(err)
}

func TestEmailHasUnderscoreAfterAtSign(t *testing.T) {
	email := EmailAddress("user@google_website.biz")
	err := email.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Println(err)
}

func TestValidEmail(t *testing.T) {
	email := EmailAddress("user@google.com")
	err := email.StaticallyValid()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(err)
}
