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
