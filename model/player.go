package model

import (
	"errors"
	"intraclub/common"
)

type Player struct {
	ID     string
	UserId string
	Line   int
}

func (p *Player) RecordType() string {
	return "player"
}

func (p *Player) OneRecord() common.CrudRecord {
	return new(Player)
}

type listOfPlayers []*Player

func (l listOfPlayers) Length() int {
	return len(l)
}

func (p *Player) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfPlayers, 0)
}

func (p *Player) SetId(id string) {
	p.ID = id
}

func (p *Player) GetId() string {
	return p.ID
}

func (p *Player) ValidateDynamic(db common.DbProvider) error {
	err := common.CheckExistenceOrError(db, &User{ID: p.UserId})
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
