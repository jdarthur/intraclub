package common

import (
	"fmt"
	"os"
	"testing"
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

	userId := NewRecordId()
	token, err := GenerateToken(userId)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	at, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if at.UserId != userId {
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

	userId := NewRecordId()
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

	userId1 := NewRecordId()
	userId2 := NewRecordId()

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

	if at1.UserId != userId1 {
		t.Fatalf("token1.UserId != userId1")
	}
	if at2.UserId != userId2 {
		t.Fatalf("token2.UserId != userId2")
	}
	if token1 == token2 {
		t.Fatal("tokens for different users should not be equal")
	}
}
