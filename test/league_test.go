package test

import (
	"intraclub/common"
	"intraclub/model"
	"testing"
)

func init() {
	ResetDatabase()
}

func TestCommissionerNonUpdatable(t *testing.T) {

	commish1 := newUser(t)
	league := NewLeagueWithCommish(commish1)

	created, err := common.Create(common.GlobalDbProvider, league)
	if err != nil {
		t.Error(err)
	}

	league = created.(*model.League)

	commish2 := newUser(t)
	update := copyLeague(league)
	update.Commissioner = commish2.ID.Hex()

	err = common.Update(common.GlobalDbProvider, update)
	if err == nil {
		t.Errorf("expected updated commissioner ID to throw error")
	}

	ValidateErrorContains(t, err, "is not updatable")
}

func TestInvalidColorIncorrectLength(t *testing.T) {

	commish := newUser(t)
	league := &model.League{
		Colors: []model.TeamColor{{
			Name: "test color",
			Hex:  "1234",
		}},
		Commissioner: commish.ID.Hex(),
	}

	_, err := common.Create(common.GlobalDbProvider, league)
	if err == nil {
		t.Errorf("expected invalid color hex to throw error")
	}

	ValidateErrorContains(t, err, "invalid team color")
	ValidateErrorContains(t, err, "hex code")
}

func TestInvalidColorNonHex(t *testing.T) {

	commish := newUser(t)
	league := &model.League{
		Colors: []model.TeamColor{{
			Name: "test color",
			Hex:  "xyzabc",
		}},
		Commissioner: commish.ID.Hex(),
	}

	_, err := common.Create(common.GlobalDbProvider, league)
	if err == nil {
		t.Errorf("expected invalid color hex to throw error")
	}

	ValidateErrorContains(t, err, "invalid team color")
	ValidateErrorContains(t, err, "hex code")
}

func TestDuplicateColorId(t *testing.T) {

	commish := newUser(t)
	league := &model.League{
		Colors:       []model.TeamColor{model.Blue, model.Green, model.Red, model.White, model.Blue},
		Commissioner: commish.ID.Hex(),
	}

	_, err := common.Create(common.GlobalDbProvider, league)
	if err == nil {
		t.Errorf("expected duplicate color to throw error")
	}

	ValidateErrorContains(t, err, "duplicate color name")
}

func TestDuplicateColorHex(t *testing.T) {

	commish := newUser(t)
	league := &model.League{
		Colors: []model.TeamColor{
			{
				Name: "color1",
				Hex:  "ffffff",
			},
			{
				Name: "color2",
				Hex:  "ffffff",
			},
		},
		Commissioner: commish.ID.Hex(),
	}

	_, err := common.Create(common.GlobalDbProvider, league)
	if err == nil {
		t.Errorf("expected duplicate color to throw error")
	}

	ValidateErrorContains(t, err, "duplicate color hex code")
}

func copyLeague(league *model.League) *model.League {
	return &model.League{
		ID:           league.ID,
		Colors:       league.Colors,
		Commissioner: league.Commissioner,
		Facility:     league.Facility,
	}
}
