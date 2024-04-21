package model

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

// Matchup represents a particular line for a particular team
type Matchup struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`                            // unique ID for this particular matchup
	TeamId         string             `json:"team_id" bson:"team_id"`                   // id for the team in question
	Line1          int                `json:"line_1" bson:"line_1"`                     // 1, 2, or 3
	Player1        string             `json:"player_1" bson:"player_1"`                 // Player.ID of the Line1 player
	Player1Penalty bool               `json:"player_1_penalty" bson:"player_1_penalty"` // true if Player1 was playing down for this matchup
	Line2          int                `json:"line_2" bson:"line_2"`                     // 1, 2, or 3
	Player2        string             `json:"player_2" bson:"player_2"`                 // Player.ID of the Line2 player
	Player2Penalty bool               `json:"player_2_penalty" bson:"player_2_penalty"` // true if Player2 was playing down for this matchup
}

func (m *Matchup) RecordType() string {
	return "matchup"
}

func (m *Matchup) OneRecord() common.CrudRecord {
	return new(Matchup)
}

type listOfMatchups []*Matchup

func (l listOfMatchups) Get(index int) common.CrudRecord {
	return l[index]
}

func (l listOfMatchups) Length() int {
	return len(l)
}

func (m *Matchup) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfMatchups, 0)
}

func (m *Matchup) SetId(id primitive.ObjectID) {
	m.ID = id
}

func (m *Matchup) GetId() primitive.ObjectID {
	return m.ID
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

func (m *Matchup) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {
	player1, err := GetPlayer(db, m.Player1)
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

	player2, err := GetPlayer(db, m.Player2)
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

func GetPlayer(db common.DbProvider, playerId string) (*Player, error) {

	id, err := primitive.ObjectIDFromHex(playerId)
	if err != nil {
		return nil, err
	}

	player, exists, err := common.GetOne(db, &Player{ID: id})
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, common.RecordDoesNotExist(&Player{ID: id})
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
