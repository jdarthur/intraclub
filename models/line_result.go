package models

import (
	"errors"
	"fmt"
)

type LineResult struct {
	Team1      Matchup     `json:"team_1_matchup"`
	Team2      Matchup     `json:"team_2_matchup"`
	SetResults []SetResult `json:"set_results"`
}

func (l LineResult) Winner() string {
	team1SetsWon := 0
	for _, result := range l.SetResults {
		if result.Team1.GamesWon > result.Team2.GamesWon {
			team1SetsWon += 1
		}
	}
	winner := l.Team1.Team
	if team1SetsWon < 2 {
		winner = l.Team2.Team
	}

	return winner
}

func (l LineResult) ValidateStatic() error {
	if l.Team1.Team == l.Team2.Team {
		return errors.New("team 1's ID == team 2's ID")
	}

	err := l.Team1.ValidateStatic()
	if err != nil {
		ret := fmt.Sprintf("team 1 matchup was invalid: %s", err.Error())
		return errors.New(ret)
	}

	err = l.Team2.ValidateStatic()
	if err != nil {
		ret := fmt.Sprintf("team 2 matchup was invalid: %s", err.Error())
		return errors.New(ret)
	}

	if l.Team1.Line1 != l.Team2.Line1 {
		ret := fmt.Sprintf("team 1's matchup line 1 (%d) did not match team 2's matchup line 1 (%d)", l.Team1.Line1, l.Team2.Line1)
		return errors.New(ret)
	}

	if l.Team1.Line2 != l.Team2.Line2 {
		ret := fmt.Sprintf("team 1's matchup line 2 (%d) did not match team 2's matchup line 2 (%d)", l.Team1.Line2, l.Team2.Line2)
		return errors.New(ret)
	}

	if len(l.SetResults) > 3 {
		return errors.New("the set results array in the line result had more than 3 entries")
	}

	for i, result := range l.SetResults {
		err = result.ValidateStatic()
		if err != nil {
			ret := fmt.Sprintf("set result %d was invalid: %s", i+1, err.Error())
			return errors.New(ret)
		}

		if result.Team1.ID != l.Team1.Team {
			ret := fmt.Sprintf("For set %d, team 1's ID (%s) did not match the overall line result (%s)", i+1, result.Team1, l.Team1)
			return errors.New(ret)
		}

		if result.Team2.ID != l.Team2.Team {
			ret := fmt.Sprintf("For set %d, team 2's ID (%s) did not match the overall line result (%s)", i+1, result.Team2, l.Team2)
			return errors.New(ret)
		}

		if i != 2 && result.Team1.TiebreakSet {
			return errors.New("team 1 had set 1 or 2 marked as a tiebreak set")
		}
		if i != 2 && result.Team2.TiebreakSet {
			return errors.New("team 2 had set 1 or 2 marked as a tiebreak set")
		}
		if result.Team1.TiebreakSet == true && result.Team2.TiebreakSet != true {
			return errors.New("team 1 had set marked as a tiebreak set, but team 2 did not")
		}
		if result.Team2.TiebreakSet == true && result.Team1.TiebreakSet != true {
			return errors.New("team 2 had set marked as a tiebreak set, but team 2 did not")
		}

	}

	return nil
}

func (l LineResult) ValidateDynamic() error {
	// validate that Team 1's ID is valid

	// validate that Team 2's ID is valid

	return nil
}

// Calculate a point total for a team from a LineResult
func (l LineResult) Calculate(teamId string) (int, error) {

	if teamId != l.Team1.Team && teamId != l.Team2.Team {
		return -1, errors.New("provided team ID is neither team 1's nor team 2's ID")
	}

	isTeam1 := teamId == l.Team1.Team

	totalPoints := 0

	if isTeam1 {
		if l.Team1.Player1Penalty {
			totalPoints -= 5
		}

		if l.Team1.Player2Penalty {
			totalPoints -= 5
		}
	} else {
		if l.Team2.Player1Penalty {
			totalPoints -= 5
		}

		if l.Team2.Player2Penalty {
			totalPoints -= 5
		}
	}

	for _, result := range l.SetResults {
		err := result.ValidateStatic()
		if err != nil {
			return -1, err
		}

		if isTeam1 {
			totalPoints += result.Team1.GamesWon
		} else {
			totalPoints += result.Team2.GamesWon
		}
	}

	if l.Winner() == teamId {
		totalPoints += 5
	}

	return totalPoints, nil
}
