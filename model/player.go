package model

import (
	"errors"
)

type Player struct {
	UserId string
	Line   int
}

func (p Player) ValidateStatic() error {
	if p.Line <= 0 {
		return errors.New("player's line is less than or equal to zero")
	}

	if p.Line > 3 {
		return errors.New("player's line is greater than three")
	}

	//_, exists := controllers.UserExists(p.UserId)
	//if !exists {
	//return fmt.Errorf("user ID %s is not a valid user ID", p.UserId)
	//}
	return nil
}
