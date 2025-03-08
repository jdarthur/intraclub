package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newStoredMatchPair(t *testing.T, db common.DatabaseProvider, s *ScoringStructure) (*Match, *Match) {
	match1 := NewMatch()
	match2 := NewMatch()

	match1.Structure = s.ID
	match2.Structure = s.ID

	match1.Editors = []UserId{s.Owner}
	match2.Editors = []UserId{s.Owner}

	created1, err := common.CreateOne(db, match1)
	if err != nil {
		t.Fatal(err)
	}
	match2.Opponent = created1.ID
	created2, err := common.CreateOne(db, match2)
	if err != nil {
		t.Fatal(err)
	}

	created1.Opponent = created2.ID
	err = common.UpdateOne(db, created1)
	if err != nil {
		t.Fatal(err)
	}
	err = created1.Initialize(db)
	if err != nil {
		t.Fatal(err)
	}
	err = created2.Initialize(db)
	if err != nil {
		t.Fatal(err)
	}

	return created1, created2
}

var sixZeroDustedFlow = []bool{
	true, true, true, true, true, true, // win set one
	true, true, true, true, true, true, // win set two
}

func TestMatchFlow(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	ss := newDefaultStoredScoringStructure(t, db)
	match1, match2 := newStoredMatchPair(t, db, ss)

	for _, won := range sixZeroDustedFlow {
		if won {
			err := match1.IncrementSecondary(db)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			err := match2.IncrementSecondary(db)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	if match1.Status != MatchWon {
		t.Fatal("expected match to be won")
	}
	if match2.Status != MatchLost {
		t.Fatal("expected match to be lost")
	}
	fmt.Println(match1.String(match2))
	fmt.Println(match2.String(match1))
}
