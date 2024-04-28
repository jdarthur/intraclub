package model

import (
	"encoding/hex"
	"errors"
	"fmt"
)

type TeamColor struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

// ValidateStatic checks that the TeamColor has:
//   - a non-empty name
//   - a non-empty, 6-character hex code that is parseable as valid hex
func (t TeamColor) ValidateStatic() error {
	if t.Name == "" {
		return fmt.Errorf("name must not be empty")
	}

	if t.Hex == "" {
		return fmt.Errorf("hex code must not be empty")
	}

	if len(t.Hex) != 6 {
		return errors.New("hex code must be in format 'xxxxxx'")
	}

	_, err := hex.DecodeString(t.Hex)
	if err != nil {
		return fmt.Errorf("invalid hex code (error: %s)", err)
	}

	return nil
}

var Blue = TeamColor{Name: "Blue", Hex: "2b52ed"}
var Green = TeamColor{Name: "Green", Hex: "00913d"}
var Red = TeamColor{Name: "Red", Hex: "ff1f1f"}
var White = TeamColor{Name: "White", Hex: "ffffff"}
var Black = TeamColor{Name: "Black", Hex: "000000"}
var Pink = TeamColor{Name: "Pink", Hex: "ff73b0"}
var Purple = TeamColor{Name: "Purple", Hex: "7900bf"}
var Yellow = TeamColor{Name: "Yellow", Hex: "ffff45"}
var Orange = TeamColor{Name: "Orange", Hex: "ff961f"}

var DefaultColors = []TeamColor{
	Blue,
	Green,
	Red,
	White,
	Black,
	Pink,
	Purple,
	Yellow,
	Orange,
}
