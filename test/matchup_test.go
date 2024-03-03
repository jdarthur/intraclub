package test

import (
	"fmt"
	"intraclub/model"
	"testing"
)

func TestMatchupValid(t *testing.T) {
	m := model.Matchup{
		Line1: 1,
		Line2: 1,
	}

	err := m.ValidateStatic()
	if err != nil {
		t.Error(err)
	}
}

func TestZeroLine(t *testing.T) {
	m := model.Matchup{
		Line1: 0,
		Line2: 1,
	}

	err := m.ValidateStatic()
	if err == nil {
		t.Error("Putting zero for line 1 value did not produce an error")
	}

	fmt.Println(err)

	m = model.Matchup{
		Line1: 1,
		Line2: 0,
	}

	err = m.ValidateStatic()
	if err == nil {
		t.Error("Putting zero for line 1 value did not produce an error")
	}

	fmt.Println(err)
}

func TestGreaterThanThreeLine(t *testing.T) {
	m := model.Matchup{
		Line1: 4,
		Line2: 1,
	}

	err := m.ValidateStatic()
	if err == nil {
		t.Error("Putting 4 for line 1 value did not produce an error")
	}

	fmt.Println(err)

	m = model.Matchup{
		Line1: 1,
		Line2: 4,
	}

	err = m.ValidateStatic()
	if err == nil {
		t.Error("Putting 4 for line 2 value did not produce an error")
	}

	fmt.Println(err)
}
