package model

import (
	"errors"
	"fmt"
)

// Matchup represents a particular line
type Matchup struct {
	TeamId         string `json:"team"`
	Line1          int    `json:"line_1"`   // 1, 2, or 3
	Player1        string `json:"player_1"` // UUID of player 1
	Player1Penalty bool   `json:"player_1_penalty"`
	Line2          int    `json:"line_2"`   // 1, 2, or 3
	Player2        string `json:"player_2"` // UUID of player 2
	Player2Penalty bool   `json:"player_2_penalty"`
}

// ValidateStatic tests a Matchup and returns an error if either of the
// lines provided are outside of the range {1, 2, 3}
func (m Matchup) ValidateStatic() error {

	if m.Line1 <= 0 {
		return errors.New("line 1 value was <= 0")
	}

	if m.Line1 > 3 {
		return errors.New("line 1 value was > 3")
	}

	if m.Line2 <= 0 {
		return errors.New("line 2 value was <= 0")
	}

	if m.Line2 > 3 {
		return errors.New("line 2 value was > 3")
	}

	return nil
}

func (m Matchup) ValidateDynamic() error {
	player1, err := getPlayer(m.Player1)
	if err != nil {
		return err
	}

	if m.Player1Penalty == true {
		if player1.Line >= m.Line1 {
			ret := fmt.Sprintf("player 1 was marked as a line penalty, but their line (%d) is >= the line 1 value (%d)", player1.Line, m.Line1)
			return errors.New(ret)
		}
	} else {
		if player1.Line < m.Line1 {
			ret := fmt.Sprintf("player 1's line is %d, but they are marked as playing at line %d without penalty)", player1.Line, m.Line1)
			return errors.New(ret)
		}
	}

	player2, err := getPlayer(m.Player2)
	if err != nil {
		return err
	}

	if m.Player2Penalty == true {
		if player2.Line >= m.Line2 {
			ret := fmt.Sprintf("player 2 was marked as a line penalty, but their line (%d) is >= the line 2 value (%d)", player2.Line, m.Line2)
			return errors.New(ret)
		}
	} else {
		if player2.Line < m.Line2 {
			ret := fmt.Sprintf("player 2's line is %d, but they are marked as playing at line %d without penalty)", player2.Line, m.Line2)
			return errors.New(ret)
		}
	}

	return nil
}

func getPlayer(playerId string) (Player, error) {
	return Player{}, nil
}
