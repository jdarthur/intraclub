package test

import (
	"fmt"
	"intraclub/model"
	"testing"
)

// TestNeitherTeamTo6Games should throw an error when you
// have a set result where neither team won at least 6 games
func TestNeitherTeamTo6Games(t *testing.T) {
	setResult := baseSetResult(5, 5)

	err := setResult.ValidateStatic()
	if err == nil {
		t.Error("Expected error on set result where neither team won 6 games")
	}

	fmt.Println(err)
}

// Test7To4SetResult should throw an error when a team win's 7 games without
// the other team winning at least 5 games.
func Test7To4SetResult(t *testing.T) {
	setResult := baseSetResult(7, 4)

	err := setResult.ValidateStatic()
	if err == nil {
		t.Error("Expected error on 7-4 set result")
	}

	fmt.Println(err)
}

// Test7To4SetResult should throw an error when a team win's 7 games without
// the other team winning at least 5 games.
func Test6To5SetResult(t *testing.T) {
	setResult := baseSetResult(6, 5)

	err := setResult.ValidateStatic()
	if err == nil {
		t.Error("Expected error on 6-5 set result")
	}

	fmt.Println(err)
}

func Test6To6SetResult(t *testing.T) {
	setResult := baseSetResult(6, 6)

	err := setResult.ValidateStatic()
	if err == nil {
		t.Error("Expected error on 6-6 set result")
	}

	fmt.Println(err)
}

func Test7To7SetResult(t *testing.T) {
	setResult := baseSetResult(7, 7)

	err := setResult.ValidateStatic()
	if err == nil {
		t.Error("Expected error on 6-6 set result")
	}

	fmt.Println(err)
}

func TestTiebreak(t *testing.T) {
	setResult := baseSetResult(7, 6)
	err := setResult.ValidateStatic()
	if err != nil {
		t.Error(err)
	}

	if setResult.Tiebreak() == false {
		t.Error("Expected 7-6 set to be marked as a tiebreak")
	}

	if setResult.TiebreakWinner() != setResult.Team1.ID {
		t.Error("Expected tiebreak winner to be team 1's ID")
	}
}

func baseSetResult(team1Games, team2Games int) model.SetResult {
	return model.SetResult{
		Team1: model.TeamSetResult{
			ID:       Team1Id,
			GamesWon: team1Games,
		},
		Team2: model.TeamSetResult{
			ID:       Team2Id,
			GamesWon: team2Games,
		},
	}
}
