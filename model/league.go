package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

type League struct {
	ID           string      `json:"league_id"`
	Colors       []TeamColor `json:"colors"`
	Commissioner string      `json:"commissioner"`
}

func (l *League) GetUserId() string {
	return l.Commissioner
}

func (l *League) RecordType() string {
	return "league"
}

func (l *League) OneRecord() common.CrudRecord {
	return new(League)
}

type listOfLeagues []*League

func (l listOfLeagues) Length() int {
	return len(l)
}

func (l *League) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfLeagues, 0)
}

func (l *League) SetId(id string) {
	l.ID = id
}

func (l *League) GetId() string {
	return l.ID
}

func (l *League) ValidateStatic() error {

	for i, color := range l.Colors {
		err := color.ValidateStatic()
		if err != nil {
			return fmt.Errorf("invalid team color at index %d: %s", i, err.Error())
		}
	}

	return nil
}

func (l *League) ValidateDynamic(db common.DbProvider) error {
	_, exists, err := common.GetOne(common.GlobalDbProvider, &User{ID: l.Commissioner})
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("user with ID %s does not exist", l.Commissioner)
	}

	return nil
}

func NewLeague(colors []TeamColor, commissioner string) League {
	leagueId := primitive.NewObjectID()
	return League{
		ID:           leagueId.String(),
		Colors:       colors,
		Commissioner: commissioner,
	}
}
