package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

type Team struct {
	ID         string    `json:"team_id"`     // unique ID for this team
	Year       int       `json:"-"`           // year for this team, e.g. 2024
	Name       string    `json:"name"`        // custom team name
	Color      TeamColor `json:"color"`       // red, blue, green, white
	CaptainId  string    `json:"captain"`     // user ID of captain
	CoCaptains []string  `json:"co_captains"` // user ID(s) of any co-captains
	Players    []Player  `json:"players"`     // list of Player s on team
}

func (t *Team) RecordType() string {
	return "team"
}

func (t *Team) OneRecord() common.CrudRecord {
	return new(Team)
}

func (t *Team) ListOfRecords() interface{} {
	return make([]*Team, 0)
}

func (t *Team) SetId(id string) {
	t.ID = id
}

func (t *Team) GetId() string {
	return t.ID
}

func (t *Team) ValidateStatic() error {

	err := t.Color.ValidateStatic()
	if err != nil {
		return fmt.Errorf("invalid team color: %s", err)
	}

	for _, player := range t.Players {
		err = player.ValidateStatic()
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Team) ValidateDynamic(db common.DbProvider) error {

	err := common.CheckExistenceOrError(&User{ID: t.CaptainId})
	if err != nil {
		return err
	}

	for _, coCaption := range t.CoCaptains {
		err = common.CheckExistenceOrError(&User{ID: coCaption})
		if err != nil {
			return err
		}
	}

	for _, player := range t.Players {
		err = common.CheckExistenceOrError(&User{ID: player.UserId})
		if err != nil {
			return err
		}
	}

	return nil
}

func NewTeam(color TeamColor, captain string) Team {
	id := primitive.NewObjectID()

	return Team{
		ID:        id.String(),
		Color:     color,
		CaptainId: captain,
	}
}
