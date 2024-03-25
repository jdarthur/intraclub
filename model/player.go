package model

import (
	"errors"
	"intraclub/common"
)

type Player struct {
	UserId string
	Line   int
}

func (p *Player) ValidateDynamic(db common.DbProvider) error {
	err := common.CheckExistenceOrError(&User{ID: p.UserId})
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
