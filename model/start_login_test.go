package model

import (
	"fmt"
	"intraclub/common"
	"testing"
	"time"
)

func init() {
	err := common.DeleteKeyPair()
	if err != nil {
		panic(err)
	}
	err = common.GenerateJwtKeyPairIfNotExists()
	if err != nil {
		panic(err)
	}
}

func TestInvalidUserId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	m := &StartLoginTokenManager{}
	email := "fake@email.com"

	request := &RequestForLoginToken{
		Email: EmailAddress(email),
	}

	_, _, err := m.RequestToken(db, request)
	if err == nil {
		t.Fatalf("InvalidUserId should fail")
	}
	fmt.Println(err)
}

func TestValidUserId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	m := &StartLoginTokenManager{}
	user := newStoredUser(t, db)

	request := &RequestForLoginToken{
		Email: user.Email,
	}

	token, doesNotExist, err := m.RequestToken(db, request)
	if err != nil {
		t.Fatal(err)
	}
	if doesNotExist {
		t.Fatalf("Token should exist")
	}
	fmt.Printf("%s\n", token)
}

func TestGetLoginResponse(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	m := &StartLoginTokenManager{}
	user := newStoredUser(t, db)

	request := &RequestForLoginToken{
		Email: user.Email,
	}

	token, _, err := m.RequestToken(db, request)
	if err != nil {
		t.Fatal(err)
	}

	resp := m.LoginViaToken(token.Token)
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}
	fmt.Printf("%+v\n", resp)
}

func TestDoubleLogin(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	m := &StartLoginTokenManager{}
	user := newStoredUser(t, db)

	request := &RequestForLoginToken{
		Email: user.Email,
	}

	token, _, err := m.RequestToken(db, request)
	if err != nil {
		t.Fatal(err)
	}

	_ = m.LoginViaToken(token.Token)
	resp2 := m.LoginViaToken(token.Token)
	if resp2.Error == nil {
		t.Fatalf("LoginViaToken should fail")
	}
	fmt.Println(resp2.Error)
}

func TestTokenExpired(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	m := &StartLoginTokenManager{}
	user := newStoredUser(t, db)

	LoginTokenDefaultExpirationTime = time.Millisecond * 5

	request := &RequestForLoginToken{
		Email: user.Email,
	}

	token, _, err := m.RequestToken(db, request)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 10)

	resp2 := m.LoginViaToken(token.Token)
	if resp2.Error == nil {
		t.Fatalf("LoginViaToken should fail")
	}
	fmt.Println(resp2.Error)
}

func TestTokenIsFake(t *testing.T) {
	m := &StartLoginTokenManager{}
	resp2 := m.LoginViaToken("fake token")
	if resp2.Error == nil {
		t.Fatalf("LoginViaToken should fail")
	}
	fmt.Println(resp2.Error)
}
