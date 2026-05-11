package model

import (
	"context"
	"crypto/rand"
	"testing"

	"intraclub/database"
)

func newStoredPhoto(t *testing.T, db database.DatabaseProvider, owner database.UserId) *Photo {
	b := make([]byte, 64)
	n, err := rand.Read(b)
	if n != 64 {
		t.Fatal("failed to generate random data")
	}
	if err != nil {
		t.Fatal(err)
	}
	photo := NewPhoto()
	photo.Owner = owner
	photo.Contents = b
	photo.FileType = PhotoTypeJpeg
	v, err := database.CreateOne(context.Background(), db, photo)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
