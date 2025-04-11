package model

import (
	"fmt"
	"intraclub/common"
	"testing"
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
	WinConditionCountingType: Point,
	WinCondition: WinCondition{
		WinThreshold:        10,
		MustWinBy:           2,
		InstantWinThreshold: 0,
	},
}

func newDefaultStoredScoringStructure(t *testing.T, db common.DatabaseProvider) *ScoringStructure {

	s := newDefaultStoredSetScoringStructure(t, db)
	matchScoringStructure := &TennisMatchScoringStructure
	matchScoringStructure.Owner = s.Owner
	matchScoringStructure.SecondaryScoringStructures = []ScoringStructureId{
		s.ID,
		s.ID,
		s.ID,
	}

	m, err := common.CreateOne(db, matchScoringStructure)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func newThirdSetTiebreakScoringStructure(t *testing.T, db common.DatabaseProvider) *ScoringStructure {

	s := newDefaultStoredSetScoringStructure(t, db)
	s2 := newTenPointTiebreakSetScoringStructure(t, db)

	matchScoringStructure := &TennisMatchScoringStructure
	matchScoringStructure.Owner = s.Owner
	matchScoringStructure.SecondaryScoringStructures = []ScoringStructureId{
		s.ID,
		s.ID,
		s2.ID,
	}

	m, err := common.CreateOne(db, matchScoringStructure)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func newDefaultStoredSetScoringStructure(t *testing.T, db common.DatabaseProvider) *ScoringStructure {
	owner := newStoredUser(t, db)

	setScoringStructure := &TennisSetScoringStructure
	setScoringStructure.Owner = owner.ID
	s, err := common.CreateOne(db, setScoringStructure)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func newTenPointTiebreakSetScoringStructure(t *testing.T, db common.DatabaseProvider) *ScoringStructure {
	owner := newStoredUser(t, db)

	setScoringStructure := &TennisTiebreakThirdSet
	setScoringStructure.Owner = owner.ID
	s, err := common.CreateOne(db, setScoringStructure)
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
	db := common.NewUnitTestDBProvider()
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
	db := common.NewUnitTestDBProvider()
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
	db := common.NewUnitTestDBProvider()
	s := ScoringStructure{}
	s.WinConditionCountingType = Set
	s.WinCondition = WinCondition{
		WinThreshold: 6,
		MustWinBy:    2,
	}
	s.SecondaryScoringStructures = []ScoringStructureId{ScoringStructureId(common.InvalidRecordId)}

	err := s.DynamicallyValid(db)
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestInvalidOwnerId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	s := ScoringStructure{}
	s.WinConditionCountingType = Set
	s.WinCondition = WinCondition{
		WinThreshold: 6,
		MustWinBy:    2,
	}

	err := s.DynamicallyValid(db)
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}
