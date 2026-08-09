package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newStoredMatchPair(t *testing.T, db database.Provider, s *ScoringStructure) (*IndividualMatch, *IndividualMatch) {
	match1 := NewMatch()
	match2 := NewMatch()

	match1.Structure = s.ID
	match2.Structure = s.ID

	created1, err := database.CreateOne(context.Background(), db, match1)
	if err != nil {
		t.Fatal(err)
	}
	match2.Opponent = created1.ID
	created2, err := database.CreateOne(context.Background(), db, match2)
	if err != nil {
		t.Fatal(err)
	}

	created1.Opponent = created2.ID
	err = database.UpdateOne(context.Background(), db, created1)
	if err != nil {
		t.Fatal(err)
	}

	// Grant edit rights via the normalized match_editor child table (the
	// former inline Editors slice).
	if _, err := created1.AssignEditor(context.Background(), db, s.Owner); err != nil {
		t.Fatal(err)
	}
	if _, err := created2.AssignEditor(context.Background(), db, s.Owner); err != nil {
		t.Fatal(err)
	}

	err = created1.Initialize(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	err = created2.Initialize(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	return created1, created2
}

var sixZeroDustedFlow = []bool{
	true, true, true, true, true, true, // win set one
	true, true, true, true, true, true, // win set two
}

func runMatchFlow(t *testing.T, db database.Provider, match1, match2 *IndividualMatch, flow []bool) {
	for _, won := range flow {
		if won {
			err := match1.IncrementSecondary(context.Background(), db)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			err := match2.IncrementSecondary(context.Background(), db)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestMatchFlow(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ss := newDefaultStoredScoringStructure(t, db)
	match1, match2 := newStoredMatchPair(t, db, ss)

	runMatchFlow(t, db, match1, match2, sixZeroDustedFlow)

	if match1.Status != MatchWon {
		t.Fatal("expected match to be won")
	}
	if match2.Status != MatchLost {
		t.Fatal("expected match to be lost")
	}
	fmt.Println(match1.String(match2))
	fmt.Println(match2.String(match1))
}

var closeThreeSets = []bool{
	true, true, false, true, false, false, true, false, true, true, // win set one, 6-4
	true, true, false, false, true, false, false, true, false, true, false, false, // lost set two, 5-7
	false, true, true, false, false, true, false, true, true, true, // won set three, 6-4
}

func TestMatchFlow2(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ss := newDefaultStoredScoringStructure(t, db)
	match1, match2 := newStoredMatchPair(t, db, ss)

	runMatchFlow(t, db, match1, match2, closeThreeSets)

	if match1.Status != MatchWon {
		t.Fatal("expected match to be won")
	}
	if match2.Status != MatchLost {
		t.Fatal("expected match to be lost")
	}
	fmt.Println(match1.String(match2))
	fmt.Println(match2.String(match1))
}

var thirdSetTiebreak = []bool{
	true, true, false, true, false, false, true, false, true, true, // win set one, 6-4
	true, true, false, false, true, false, false, true, false, true, false, false, // lost set two, 5-7
	false, true, true, false, false, true, false, true, true, true, false, false, true, false, true, false, true, true, // won set three, 10-8
}

func TestMatchFlowThirdSetTiebreak(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ss := newThirdSetTiebreakScoringStructure(t, db)
	match1, match2 := newStoredMatchPair(t, db, ss)

	runMatchFlow(t, db, match1, match2, thirdSetTiebreak)

	if match1.Status != MatchWon {
		t.Fatal("expected match to be won")
	}
	if match2.Status != MatchLost {
		t.Fatal("expected match to be lost")
	}
	fmt.Println(match1.String(match2))
	fmt.Println(match2.String(match1))
}

func TestIndividualPointTotals(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ss := newDefaultStoredScoringStructure(t, db)
	match1, match2 := newStoredMatchPair(t, db, ss)

	runMatchFlow(t, db, match1, match2, closeThreeSets)

	if match1.GetSecondaryPointTotal() != 17 {
		t.Fatal("expected secondary point to be 17")
	}
	if match2.GetSecondaryPointTotal() != 15 {
		t.Fatal("expected secondary point to be 13")
	}
}
