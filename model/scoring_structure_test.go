package model

import (
	"intraclub/common"
	"testing"
)

func newDefaultStoredScoringStructure(t *testing.T, db common.DatabaseProvider) *ScoringStructure {

	owner := newStoredUser(t, db)
	structure := &TennisScoringStructure
	structure.Owner = owner.ID

	v, err := common.CreateOne(db, structure)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
