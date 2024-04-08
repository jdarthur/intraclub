package model

import (
	"errors"
	"fmt"
	"intraclub/common"
)

// Matchup represents a particular line for a particular team
type Matchup struct {
	ID             string `json:"id"`               // unique ID for this particular matchup
	TeamId         string `json:"team"`             // id for the team in question
	Line1          int    `json:"line_1"`           // 1, 2, or 3
	Player1        string `json:"player_1"`         // Player.ID of the Line1 player
	Player1Penalty bool   `json:"player_1_penalty"` // true if Player1 was playing down for this matchup
	Line2          int    `json:"line_2"`           // 1, 2, or 3
	Player2        string `json:"player_2"`         // Player.ID of the Line2 player
	Player2Penalty bool   `json:"player_2_penalty"` // true if Player2 was playing down for this matchup
}

func (m *Matchup) RecordType() string {
	return "matchup"
}

func (m *Matchup) OneRecord() common.CrudRecord {
	return new(Matchup)
}

type listOfMatchups []*Matchup

func (l listOfMatchups) Length() int {
	return len(l)
}

func (m *Matchup) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfMatchups, 0)
}

func (m *Matchup) SetId(id string) {
	//TODO implement me
	panic("implement me")
}

func (m *Matchup) GetId() string {
	//TODO implement me
	panic("implement me")
}

// ValidateStatic tests a Matchup and returns an error if either of the
// lines provided are outside of the range {1, 2, 3}
func (m *Matchup) ValidateStatic() error {

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

func (m *Matchup) ValidateDynamic(db common.DbProvider) error {
	player1, err := GetPlayer(m.Player1)
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

	player2, err := GetPlayer(m.Player2)
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

func GetPlayer(playerId string) (*Player, error) {
	player, exists, err := common.GetOne(common.GlobalDbProvider, &Player{ID: playerId})
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, common.RecordDoesNotExist(&Player{ID: playerId})
	}

	return player.(*Player), nil
}

func GetAllMatchupsWithPlayerPair(m *Matchup) ([]*Matchup, error) {
	matchups, err := common.GetAllWhere(common.GlobalDbProvider, m, map[string]interface{}{"player1": m.Player1, "player2": m.Player2})
	if err != nil {
		return nil, err
	}

	return matchups.(listOfMatchups), nil
}
