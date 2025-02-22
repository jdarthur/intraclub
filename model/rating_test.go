package model

import (
	"intraclub/common"
	"testing"
)

func newStoredRating(t *testing.T, db common.DatabaseProvider) *Rating {
	r := NewRating()
	r.Name = "Rating 123"
	r.Description = "test description"
	v, err := common.CreateOne(db, r)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
