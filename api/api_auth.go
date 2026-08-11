package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"intraclub/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthTokenHeaderValue is the HTTP header name used to carry the JWT auth token.
const AuthTokenHeaderValue = "X-INTRACLUB-TOKEN"

// JwtCertFile is the filename for the persisted JWT public key certificate.
const JwtCertFile = "token.crt"

// JwtKeyFile is the filename for the persisted JWT private key.
const JwtKeyFile = "token.key"

// JwtLifetime is the duration for which issued tokens remain valid. It is a
// package variable (not a const) so it can be configured at startup via the
// --jwt-lifetime flag or the INTRACLUB_JWT_LIFETIME env var (see main.go).
// Tests can shorten it to exercise expiry without waiting for the real window.
var JwtLifetime = time.Hour * 2

// JwtPublicKey holds the loaded ECDSA public key used to verify JWT signatures.
var JwtPublicKey *ecdsa.PublicKey

// JwtPrivateKey holds the loaded ECDSA private key used to sign JWT tokens.
var JwtPrivateKey *ecdsa.PrivateKey

// AuthToken represents a validated JWT token containing the authenticated user's ID.
type AuthToken struct {
	UserId database.UserId
}

// GenerateToken creates a new JWT token for the given user ID, signed with the
// private key and set to expire after JwtLifetime.
func GenerateToken(userId database.RecordId) (string, error) {
	token := jwt.New(jwt.SigningMethodES512)
	token.Claims = jwt.RegisteredClaims{
		Subject:   userId.String(),
		ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(JwtLifetime)},
		NotBefore: &jwt.NumericDate{Time: time.Now()},
		IssuedAt:  &jwt.NumericDate{Time: time.Now()},
	}

	tokenStr, err := token.SignedString(JwtPrivateKey)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// ValidateToken parses and verifies a JWT string using the public key, returning
// the embedded user ID. Returns an error if the token is malformed, expired, or
// signed with an unexpected algorithm.
func ValidateToken(token string) (*AuthToken, error) {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JwtPublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	subject, err := parsed.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	userId, err := database.RecordIdFromString(subject)
	if err != nil {
		return nil, err
	}

	return &AuthToken{
		UserId: database.UserId(userId),
	}, nil
}

// UserType is the database record type used to verify that authenticated users still exist.
var UserType database.CrudRecord

// GetToken extracts the JWT from the request header, validates it, and optionally
// checks that the user still exists in the database. Returns nil if no token is
// present, allowing unauthenticated access.
func GetToken(c *gin.Context, db database.Provider) (*AuthToken, error) {
	token := c.Request.Header.Get(AuthTokenHeaderValue)
	if token == "" {
		return nil, nil
	}
	valid, err := ValidateToken(token)
	if err != nil {
		return nil, err
	}

	userId := valid.UserId
	if db != nil && UserType != nil {
		err := database.ExistsById(c.Request.Context(), db, UserType, userId.RecordId())
		if err != nil {
			return nil, err
		}
	}

	return valid, nil
}

// doesFileExist checks whether a file exists at the given path.
func doesFileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// deleteFileIfExists removes the file at the given path, ignoring errors if it
// does not exist.
func deleteFileIfExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Remove(path)
}

// GenerateJwtKeyPairIfNotExists ensures a JWT key pair is available. If the
// key files already exist, it loads them. Otherwise, it generates a new ECDSA
// P-521 key pair and persists it to disk.
func GenerateJwtKeyPairIfNotExists() error {
	exists, err := DoesKeyPairExist()
	if err != nil {
		return err
	}
	if exists {
		JwtPublicKey, JwtPrivateKey, err = LoadKeyPair()
		if err != nil {
			return err
		}
		return nil
	}

	JwtPublicKey, JwtPrivateKey, err = GenerateKeyPair()
	if err != nil {
		return err
	}

	return SerializeKeyPair(JwtPublicKey, JwtPrivateKey)
}

// DeleteKeyPair removes the persisted JWT key files from disk.
func DeleteKeyPair() error {
	if err := deleteFileIfExists(JwtCertFile); err != nil {
		return err
	}
	return deleteFileIfExists(JwtKeyFile)
}

// DoesKeyPairExist returns true if both the certificate and key files are
// present on disk.
func DoesKeyPairExist() (bool, error) {
	exists, err := doesFileExist(JwtCertFile)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	exists, err = doesFileExist(JwtKeyFile)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, errors.New("key does not exist, but cert does")
	}
	return true, nil
}

// LoadKeyPair reads and deserializes the PEM-encoded public and private keys
// from disk.
func LoadKeyPair() (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	publicKey, err := pemDecodeFromFile(JwtCertFile, false)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := pemDecodeFromFile(JwtKeyFile, true)
	if err != nil {
		return nil, nil, err
	}
	return publicKey.(*ecdsa.PublicKey), privateKey.(*ecdsa.PrivateKey), nil
}

// GenerateKeyPair creates a new ECDSA key pair using the P-521 curve.
func GenerateKeyPair() (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	publicKey := &privateKey.PublicKey
	return publicKey, privateKey, nil
}

// SerializeKeyPair writes the public and private keys to disk as PEM-encoded
// files.
func SerializeKeyPair(publicKey *ecdsa.PublicKey, privateKey *ecdsa.PrivateKey) error {
	encoded, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}

	err = pemEncodeToFile(encoded, JwtCertFile, "PUBLIC KEY")
	if err != nil {
		return err
	}

	encoded, err = x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}

	return pemEncodeToFile(encoded, JwtKeyFile, "EC PRIVATE KEY")
}

// pemEncodeToFile writes DER-encoded bytes to a file as a PEM block with the
// given type label.
func pemEncodeToFile(b []byte, filename string, blockType string) error {
	encoded := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: b})
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Write(encoded)
	if err != nil {
		return err
	}
	if n != len(encoded) {
		return io.ErrShortWrite
	}
	return nil
}

// pemDecodeFromFile reads a PEM-encoded file and returns the decoded key. If
// private is true, it parses as an EC private key; otherwise as a public key.
func pemDecodeFromFile(filename string, private bool) (any, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	if private {
		return x509.ParseECPrivateKey(block.Bytes)
	} else {
		return x509.ParsePKIXPublicKey(block.Bytes)
	}
}
