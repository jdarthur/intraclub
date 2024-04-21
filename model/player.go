package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

type Player struct {
	ID     primitive.ObjectID `json:"player_id" bson:"_id"`
	UserId string             `json:"user_id" bson:"user_id"`
	Line   int                `json:"line" bson:"line"`
}

func (p *Player) RecordType() string {
	return "player"
}

func (p *Player) OneRecord() common.CrudRecord {
	return new(Player)
}

type listOfPlayers []*Player

func (l listOfPlayers) Get(index int) common.CrudRecord {
	return l[index]
}

func (l listOfPlayers) Length() int {
	return len(l)
}

func (p *Player) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfPlayers, 0)
}

func (p *Player) SetId(id primitive.ObjectID) {
	p.ID = id
}

func (p *Player) GetId() primitive.ObjectID {
	return p.ID
}

func (p *Player) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {
	err := common.CheckExistenceOrErrorByStringId(db, &User{}, p.UserId)
	if err != nil {
		return err
	}

	return nil
}

func (p *Player) ValidateStatic() error {
	if p.Line <= 0 {
		return errors.New("player's line is less than or equal to zero")
	}

	if p.Line > 3 {
		return errors.New("player's line is greater than three")
	}

	return nil
}
