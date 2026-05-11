package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

// TennisMatchScoringStructure is a default
var TennisMatchScoringStructure = ScoringStructure{
	WinConditionCountingType: Set,
	WinCondition: WinCondition{
		WinThreshold:        2,
		MustWinBy:           1,
		InstantWinThreshold: 0,
	},
}

var TennisSetScoringStructure = ScoringStructure{
	WinConditionCountingType: Game,
	WinCondition: WinCondition{
		WinThreshold:        6,
		MustWinBy:           2,
		InstantWinThreshold: 7,
	},
}

var TennisTiebreakThirdSet = ScoringStructure{
	WinConditionCountingType: Game,
	WinCondition: WinCondition{
		WinThreshold:        10,
		MustWinBy:           2,
		InstantWinThreshold: 0,
	},
}

func newDefaultStoredScoringStructure(t *testing.T, db database.DatabaseProvider) *ScoringStructure {

	s := newDefaultStoredSetScoringStructure(t, db)
	matchScoringStructure := &TennisMatchScoringStructure
	matchScoringStructure.Owner = s.Owner
	matchScoringStructure.Name = "test-match-scoring"
	matchScoringStructure.SecondaryScoringStructures = []ScoringStructureId{
		s.ID,
		s.ID,
		s.ID,
	}

	m, err := database.CreateOne(context.Background(), db, matchScoringStructure)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func newThirdSetTiebreakScoringStructure(t *testing.T, db database.DatabaseProvider) *ScoringStructure {

	s := newDefaultStoredSetScoringStructure(t, db)
	s2 := newTenPointTiebreakSetScoringStructure(t, db)

	matchScoringStructure := &TennisMatchScoringStructure
	matchScoringStructure.Owner = s.Owner
	matchScoringStructure.Name = "test-tiebreak-match-scoring"
	matchScoringStructure.SecondaryScoringStructures = []ScoringStructureId{
		s.ID,
		s.ID,
		s2.ID,
	}

	m, err := database.CreateOne(context.Background(), db, matchScoringStructure)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func newDefaultStoredSetScoringStructure(t *testing.T, db database.DatabaseProvider) *ScoringStructure {
	owner := newStoredUser(t, db)

	setScoringStructure := &TennisSetScoringStructure
	setScoringStructure.Owner = owner.ID
	setScoringStructure.Name = "test-set-scoring"
	s, err := database.CreateOne(context.Background(), db, setScoringStructure)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func newTenPointTiebreakSetScoringStructure(t *testing.T, db database.DatabaseProvider) *ScoringStructure {
	owner := newStoredUser(t, db)

	setScoringStructure := &TennisTiebreakThirdSet
	setScoringStructure.Owner = owner.ID
	setScoringStructure.Name = "test-tiebreak-set-scoring"
	s, err := database.CreateOne(context.Background(), db, setScoringStructure)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestInstantWinThresholdLessThanMainWinThreshold(t *testing.T) {
	s := ScoringStructure{}
	s.WinConditionCountingType = Game
	s.WinCondition = WinCondition{
		WinThreshold:        5,
		MustWinBy:           1,
		InstantWinThreshold: 3,
	}
	err := s.StaticallyValid()
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestInstantWinThresholdEqualToMainWinThresholdInWinByTwo(t *testing.T) {
	s := ScoringStructure{}
	s.WinConditionCountingType = Game
	s.WinCondition = WinCondition{
		WinThreshold:        5,
		MustWinBy:           2,
		InstantWinThreshold: 5,
	}
	err := s.StaticallyValid()
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestMainWinThresholdIsZero(t *testing.T) {
	s := ScoringStructure{}
	s.WinConditionCountingType = Game
	s.WinCondition = WinCondition{
		WinThreshold: 0,
	}
	err := s.StaticallyValid()
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestMainWinThresholdIsLessThanWinByConstraint(t *testing.T) {
	s := ScoringStructure{}
	s.WinConditionCountingType = Game
	s.WinCondition = WinCondition{
		WinThreshold: 1,
		MustWinBy:    2,
	}
	err := s.StaticallyValid()
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestWinByConstraintIsZero(t *testing.T) {
	s := ScoringStructure{}
	s.WinConditionCountingType = Game
	s.WinCondition = WinCondition{
		WinThreshold: 1,
		MustWinBy:    0,
	}
	err := s.StaticallyValid()
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestIncorrectAmountOfSecondaryScoringStructures(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ref := newDefaultStoredSetScoringStructure(t, db)

	s := ScoringStructure{}
	s.WinConditionCountingType = Set
	s.WinCondition = WinCondition{
		WinThreshold: 2,
		MustWinBy:    1,
	}
	s.SecondaryScoringStructures = []ScoringStructureId{ref.ID}

	err := s.StaticallyValid()
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestIndeterminateWinConditionWithSecondaryScoringStructures(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ref := newDefaultStoredSetScoringStructure(t, db)

	s := ScoringStructure{}
	s.WinConditionCountingType = Set
	s.WinCondition = WinCondition{
		WinThreshold: 6,
		MustWinBy:    2,
	}
	s.SecondaryScoringStructures = []ScoringStructureId{ref.ID}

	err := s.StaticallyValid()
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestInvalidSecondaryScoreReference(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	s := ScoringStructure{}
	s.WinConditionCountingType = Set
	s.WinCondition = WinCondition{
		WinThreshold: 6,
		MustWinBy:    2,
	}
	s.SecondaryScoringStructures = []ScoringStructureId{ScoringStructureId(database.InvalidRecordId)}

	err := s.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestInvalidOwnerId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	s := ScoringStructure{}
	s.WinConditionCountingType = Set
	s.WinCondition = WinCondition{
		WinThreshold: 6,
		MustWinBy:    2,
	}

	err := s.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}
