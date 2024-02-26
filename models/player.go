package models

import "errors"

type Player struct {
	PlayerId  string
	FirstName string
	LastName  string
	Line      int
}

func (p Player) ValidateStatic() error {
	if p.Line <= 0 {
		return errors.New("player's line is less than or equal to zero")
	}

	if p.Line > 3 {
		return errors.New("player's line is greater than three")
	}

	if p.FirstName == "" {
		return errors.New("player's first name is empty")
	}

	if p.LastName == "" {
		return errors.New("player's last name is empty")
	}

	return nil
}
