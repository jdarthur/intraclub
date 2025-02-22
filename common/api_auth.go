package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var AuthTokenHeaderValue = "X-INTRACLUB-TOKEN"
var JwtCertFile = "token.crt"
var JwtKeyFile = "token.key"
var JwtLifetime = time.Hour * 2

var JwtPublicKey *ecdsa.PublicKey
var JwtPrivateKey *ecdsa.PrivateKey

type AuthToken struct {
	UserId RecordId
}

func GenerateToken(userId RecordId) (string, error) {
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

func ValidateToken(token string) (*AuthToken, error) {

	parse, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JwtPublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	subject, err := parse.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	userId, err := RecordIdFromString(subject)
	if err != nil {
		return nil, err
	}

	return &AuthToken{
		UserId: userId,
	}, nil
}

func GetToken(c *gin.Context) (*AuthToken, error) {
	token := c.Request.Header.Get(AuthTokenHeaderValue)
	if token == "" {
		return nil, nil
	}
	valid, err := ValidateToken(token)
	if err != nil {
		return nil, err
	}
	return valid, nil
}

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

func LoadKeyPair() (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	publicKey, err := pemDecodeFromFile(JwtCertFile, false)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := pemDecodeFromFile(JwtKeyFile, false)
	if err != nil {
		return nil, nil, err
	}
	return publicKey.(*ecdsa.PublicKey), privateKey.(*ecdsa.PrivateKey), nil
}

func GenerateKeyPair() (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	publicKey := &privateKey.PublicKey
	return publicKey, privateKey, nil
}

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

	return pemEncodeToFile(encoded, JwtCertFile, "EC PRIVATE KEY")
}

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

func pemDecodeFromFile(filename string, private bool) (any, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
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
