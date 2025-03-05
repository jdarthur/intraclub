package model

import (
	"crypto/rand"
	"intraclub/common"
	"testing"
)

func newStoredPhoto(t *testing.T, db common.DatabaseProvider, owner UserId) *Photo {
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
	v, err := common.CreateOne(db, photo)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
