package models

import (
	"errors"
	"fmt"
)

type Match struct {
	Team1Matchup Matchup     `json:"team_1_matchup"`
	Team2Matchup Matchup     `json:"team_2_matchup"`
	SetResults   []SetResult `json:"set_results"`
}

type Scorecard struct {
	Team1   string       `json:"team_1"`
	Team2   string       `json:"team_2"`
	Results []LineResult `json:"results"`
}

func (s Scorecard) TotalPoints(teamId string) (int, error) {
	points := 0
	for _, result := range s.Results {
		p, err := result.Calculate(teamId)
		if err != nil {
			return -1, err
		}

		points += p
	}

	return points, nil
}

func (s Scorecard) ValidateStatic() error {
	team1Points := 0
	team2Points := 0
	for _, result := range s.Results {
		err := result.ValidateStatic()
		if err != nil {
			err := fmt.Sprintf("error in result: %s", err.Error())
			return errors.New(err)
		}

		p1, err := result.Calculate(s.Team1)
		if err != nil {
			return err
		}

		team1Points += p1

		p2, err := result.Calculate(s.Team2)
		if err != nil {
			return err
		}

		team2Points += p2
	}

	AllLines := map[string]bool{
		"1-1": false,
		"1-2": false,
		"1-3": false,
		"2-2": false,
		"2-3": false,
		"3-3": false,
	}

	// loop through all results in the list
	for _, result := range s.Results {

		invalidMatchupError := result.Team1.ValidateStatic()
		if invalidMatchupError != nil {
			return invalidMatchupError
		}

		lineString := fmt.Sprintf("%d-%d", result.Team1.Line1, result.Team1.Line2)
		if AllLines[lineString] == true {
			err := fmt.Sprintf("Team 1 had multiple entries for line %s", lineString)
			return errors.New(err)
		}

		AllLines[lineString] = true
	}

	AllLines = map[string]bool{
		"1-1": false,
		"1-2": false,
		"1-3": false,
		"2-2": false,
		"2-3": false,
		"3-3": false,
	}

	for _, result := range s.Results {

		invalidMatchupError := result.Team2.ValidateStatic()
		if invalidMatchupError != nil {
			return invalidMatchupError
		}

		lineString := fmt.Sprintf("%d-%d", result.Team2.Line1, result.Team2.Line2)
		if AllLines[lineString] == true {
			err := fmt.Sprintf("Team 2 had multiple entries for line %s", lineString)
			return errors.New(err)
		}

		AllLines[lineString] = true
	}

	for key, value := range AllLines {
		if value == false {
			err := fmt.Sprintf("Missing line %s", key)
			return errors.New(err)
		}
	}

	return nil
}
