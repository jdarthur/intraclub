package api

import (
	"fmt"
	"os"
	"testing"
	"time"

	"intraclub/database"
)

func deleteIfExists(t *testing.T, filename string) {
	exists, err := doesFileExist(filename)
	if err != nil {
		t.Fatalf("Error in doesFileExist of %s: %s", filename, err)
	}
	if exists {
		err = os.Remove(filename)
		if err != nil {
			t.Fatalf("Error in os.Remove of %s: %s", filename, err)
		}
	}
}

func deleteKeyPair(t *testing.T) {
	deleteIfExists(t, JwtCertFile)
	deleteIfExists(t, JwtKeyFile)
}

func TestJwtKeyPairCreation(t *testing.T) {
	_, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Error in GenerateKeyPair: %s", err)
	}
}

func TestCreateToken(t *testing.T) {
	deleteKeyPair(t)
	err := GenerateJwtKeyPairIfNotExists()
	if err != nil {
		t.Fatalf("GenerateJwtKeyPairIfNotExists failed: %v", err)
	}

	userId := database.NewRecordId()
	token, err := GenerateToken(userId)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	at, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if at.UserId != database.UserId(userId) {
		t.Fatalf("token2.Owner != userId")
	}

	fmt.Printf("%+v\n", at)
}

func TestValidateTokenTampered(t *testing.T) {
	deleteKeyPair(t)
	err := GenerateJwtKeyPairIfNotExists()
	if err != nil {
		t.Fatalf("GenerateJwtKeyPairIfNotExists failed: %v", err)
	}

	userId := database.NewRecordId()
	token, err := GenerateToken(userId)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Tamper with the token
	tampered := token[:len(token)-10] + "0000000000"
	_, err = ValidateToken(tampered)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestValidateTokenMalformed(t *testing.T) {
	_, err := ValidateToken("not-a-valid-token")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestValidateTokenEmpty(t *testing.T) {
	_, err := ValidateToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestGenerateTokenDifferentUsers(t *testing.T) {
	deleteKeyPair(t)
	err := GenerateJwtKeyPairIfNotExists()
	if err != nil {
		t.Fatalf("GenerateJwtKeyPairIfNotExists failed: %v", err)
	}

	userId1 := database.NewRecordId()
	userId2 := database.NewRecordId()

	token1, err := GenerateToken(userId1)
	if err != nil {
		t.Fatalf("GenerateToken for user1 failed: %v", err)
	}
	token2, err := GenerateToken(userId2)
	if err != nil {
		t.Fatalf("GenerateToken for user2 failed: %v", err)
	}

	at1, err := ValidateToken(token1)
	if err != nil {
		t.Fatalf("ValidateToken for token1 failed: %v", err)
	}
	at2, err := ValidateToken(token2)
	if err != nil {
		t.Fatalf("ValidateToken for token2 failed: %v", err)
	}

	if at1.UserId != database.UserId(userId1) {
		t.Fatalf("token1.UserId != userId1")
	}
	if at2.UserId != database.UserId(userId2) {
		t.Fatalf("token2.UserId != userId2")
	}
	if token1 == token2 {
		t.Fatal("tokens for different users should not be equal")
	}
}

// TestGenerateTokenExpired validates that a token minted with an already-elapsed
// JwtLifetime is rejected. It also proves JwtLifetime is configurable (not a
// hard-coded const), which the --jwt-lifetime flag / INTRACLUB_JWT_LIFETIME env
// var rely on and which lets tests exercise the expiry path without waiting 2h.
func TestGenerateTokenExpired(t *testing.T) {
	deleteKeyPair(t)
	err := GenerateJwtKeyPairIfNotExists()
	if err != nil {
		t.Fatalf("GenerateJwtKeyPairIfNotExists failed: %v", err)
	}

	userId := database.NewRecordId()

	// A negative lifetime mints an already-expired token (ExpiresAt in the past).
	orig := JwtLifetime
	JwtLifetime = -time.Minute
	token, err := GenerateToken(userId)
	JwtLifetime = orig
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ValidateToken(token)
	if err == nil {
		t.Fatal("expected an error for an expired token")
	}
}
