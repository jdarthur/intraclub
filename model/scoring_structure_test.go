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

var testScoringSeq int

func newDefaultStoredScoringStructure(t *testing.T, db database.Provider) *ScoringStructure {

	s := newDefaultStoredSetScoringStructure(t, db)
	testScoringSeq++
	matchScoringStructure := TennisMatchScoringStructure
	matchScoringStructure.Owner = s.Owner
	matchScoringStructure.Name = fmt.Sprintf("test-match-scoring-%d", testScoringSeq)

	m, err := database.CreateOne(context.Background(), db, &matchScoringStructure)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.SetSecondaryScoringStructures(context.Background(), db, ScoringStructureList{
		s.ID,
		s.ID,
		s.ID,
	}); err != nil {
		t.Fatal(err)
	}
	return m
}

func newThirdSetTiebreakScoringStructure(t *testing.T, db database.Provider) *ScoringStructure {

	s := newDefaultStoredSetScoringStructure(t, db)
	s2 := newTenPointTiebreakSetScoringStructure(t, db)

	testScoringSeq++
	matchScoringStructure := TennisMatchScoringStructure
	matchScoringStructure.Owner = s.Owner
	matchScoringStructure.Name = fmt.Sprintf("test-tiebreak-match-scoring-%d", testScoringSeq)

	m, err := database.CreateOne(context.Background(), db, &matchScoringStructure)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.SetSecondaryScoringStructures(context.Background(), db, ScoringStructureList{
		s.ID,
		s.ID,
		s2.ID,
	}); err != nil {
		t.Fatal(err)
	}
	return m
}

func newDefaultStoredSetScoringStructure(t *testing.T, db database.Provider) *ScoringStructure {
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

func newTenPointTiebreakSetScoringStructure(t *testing.T, db database.Provider) *ScoringStructure {
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
	ctx := context.Background()
	ref := newDefaultStoredSetScoringStructure(t, db)

	s := TennisMatchScoringStructure
	s.Owner = newStoredUser(t, db).ID
	s.Name = "bad-count"
	created, err := database.CreateOne(ctx, db, &s)
	if err != nil {
		t.Fatal(err)
	}
	if err := created.SetSecondaryScoringStructures(ctx, db, ScoringStructureList{ref.ID}); err != nil {
		t.Fatal(err)
	}

	err = created.DynamicallyValid(ctx, db)
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestIndeterminateWinConditionWithSecondaryScoringStructures(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ctx := context.Background()
	ref := newDefaultStoredSetScoringStructure(t, db)

	s := ScoringStructure{}
	s.WinConditionCountingType = Set
	s.WinCondition = WinCondition{
		WinThreshold: 6,
		MustWinBy:    2,
	}
	s.Owner = newStoredUser(t, db).ID
	s.Name = "indeterminate"
	created, err := database.CreateOne(ctx, db, &s)
	if err != nil {
		t.Fatal(err)
	}
	if err := created.SetSecondaryScoringStructures(ctx, db, ScoringStructureList{ref.ID}); err != nil {
		t.Fatal(err)
	}

	err = created.DynamicallyValid(ctx, db)
	if err == nil {
		t.Fatal("expected error")
	}
	fmt.Println(err)
}

func TestInvalidSecondaryScoreReference(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ctx := context.Background()
	s := newDefaultStoredScoringStructure(t, db)

	bad := &ScoringStructureSecondary{
		ScoringStructureId:          s.ID,
		SecondaryScoringStructureId: ScoringStructureId(database.InvalidRecordId),
		SecondaryIndex:              99,
	}
	_, err := database.CreateOne(ctx, db, bad)
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

func TestScoringStructurePostDeleteCascadesSecondaries(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	sc := newDefaultStoredScoringStructure(t, db)

	count := func() int {
		rows, err := database.GetAllWhere[*ScoringStructureSecondary](context.Background(), db, func(_ context.Context, s *ScoringStructureSecondary) bool {
			return s.ScoringStructureId == sc.ID
		})
		if err != nil {
			t.Fatal(err)
		}
		return len(rows)
	}
	if count() == 0 {
		t.Fatal("expected scoring structure to have secondaries")
	}

	_, _, err := database.DeleteOneById(context.Background(), db, &ScoringStructure{}, sc.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}
	if count() != 0 {
		t.Fatal("expected 0 secondaries after delete")
	}
}
