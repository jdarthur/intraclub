package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newStoredPlayoffStructure(t *testing.T, db common.DatabaseProvider) *PlayoffStructure {
	user := newStoredUser(t, db)
	s := NewPlayoffStructure()
	s.UserId = user.ID
	s.Byes = 1
	s.NumberOfTeams = 3
	v, err := common.CreateOne(db, s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func copyPlayoffStructure(p *PlayoffStructure) *PlayoffStructure {
	return &PlayoffStructure{
		ID:            p.ID,
		UserId:        p.UserId,
		Byes:          p.Byes,
		NumberOfTeams: p.NumberOfTeams,
	}
}

func TestUserIdIsValid(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	s := NewPlayoffStructure()
	err := s.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected error on invalid user id")
	}
	fmt.Println(err)

}

func TestNumberOfTeamsIsLessThanTwo(t *testing.T) {
	s := NewPlayoffStructure()
	s.NumberOfTeams = 1
	err := s.StaticallyValid()
	if err == nil {
		t.Errorf("Expected error on 1 team playoff")
	}
	fmt.Println(err)
}

func TestOddNumberOfTeamsWithoutBye(t *testing.T) {
	s := NewPlayoffStructure()
	s.NumberOfTeams = 3
	s.Byes = 0
	err := s.StaticallyValid()
	if err == nil {
		t.Errorf("Expected error on 3 team no-bye playoff")
	}
	fmt.Println(err)
}

func TestInvalidSecondRound(t *testing.T) {
	s := NewPlayoffStructure()
	s.Byes = 1
	s.NumberOfTeams = 5
	err := s.StaticallyValid()
	if err == nil {
		t.Errorf("Expected error on 5 team, 1 bye playoff")
	}
	fmt.Println(err)
}

func TestTwoTeamTwoBye(t *testing.T) {
	s := NewPlayoffStructure()
	s.Byes = 2
	s.NumberOfTeams = 2
	err := s.StaticallyValid()
	if err == nil {
		t.Errorf("Expected error on 2 team, two bye playoff")
	}
	fmt.Println(err)
}

func TestNflFormat(t *testing.T) {
	s := NewPlayoffStructure()
	s.Byes = 1
	s.NumberOfTeams = 7
	err := s.StaticallyValid()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(err)
}

func TestTwoWildCards(t *testing.T) {
	s := NewPlayoffStructure()
	s.Byes = 6
	s.NumberOfTeams = 10
	err := s.StaticallyValid()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(err)
}

func TestPlayoffStructureCannotBeDeletedAfterUsage(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	s := newStoredPlayoffStructure(t, db)
	season := newDefaultSeason(t, db)
	season.PlayoffStructure = s.ID
	err := common.UpdateOne(db, season)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = common.DeleteOneById(db, &PlayoffStructure{}, s.ID.RecordId())
	if err == nil {
		t.Fatal(err)
	}
	fmt.Println(err)
}

func TestPlayoffStructureCannotBeEditedAfterUsage(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	s := newStoredPlayoffStructure(t, db)
	season := newDefaultSeason(t, db)
	season.PlayoffStructure = s.ID
	err := common.UpdateOne(db, season)
	if err != nil {
		t.Fatal(err)
	}

	copied := copyPlayoffStructure(s)

	err = common.UpdateOne(db, copied)
	if err == nil {
		t.Fatal("expected error when updating in-use playoff structure")
	}
	fmt.Println(err)
}

func TestNumberOfRoundsTwoRound(t *testing.T) {
	s := NewPlayoffStructure()
	s.Byes = 1
	s.NumberOfTeams = 3
	rounds := s.NumberOfRounds()
	if rounds != 2 {
		t.Errorf("Expected 2 rounds, got %d", rounds)
	}
}

func TestNumberOfRoundsNfl(t *testing.T) {
	s := NewPlayoffStructure()
	s.Byes = 2
	s.NumberOfTeams = 14
	rounds := s.NumberOfRounds()
	if rounds != 4 {
		t.Errorf("Expected 4 rounds, got %d", rounds)
	}
}

func TestNumberOfRoundsMarchMadness(t *testing.T) {
	s := NewPlayoffStructure()
	s.Byes = 0
	s.NumberOfTeams = 64
	rounds := s.NumberOfRounds()
	if rounds != 6 {
		t.Errorf("Expected 6 rounds, got %d", rounds)
	}
}

func TestNoByeNonPowerOfTwo(t *testing.T) {
	s := NewPlayoffStructure()
	s.Byes = 0
	s.NumberOfTeams = 48
	err := s.StaticallyValid()
	if err == nil {
		t.Errorf("Expected error on non power of two no-bye playoff")
	}
	fmt.Println(err)
}
