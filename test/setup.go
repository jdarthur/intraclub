package test

import (
	"intraclub/common"
	"intraclub/model"
)

var LeagueId = ""

func InitializeBlueAndGreenTeams() {

	ResetDatabase()

	league, err := common.Create(common.GlobalDbProvider, newLeague())
	if err != nil {
		panic(err)
	}

	LeagueId = league.(*model.League).ID.Hex()

	for _, user := range UnitTestUsers {
		CreatePlayer(user)
	}

	// create blue team and save ID
	team, err := common.Create(common.GlobalDbProvider, blueTeam())
	if err != nil {
		panic(err)
	}
	BlueTeamId = team.(*model.Team).ID

	// create green team and save ID
	team, err = common.Create(common.GlobalDbProvider, greenTeam())
	if err != nil {
		panic(err)
	}
	GreenTeamId = team.(*model.Team).ID
}

func CreatePlayer(user *model.User) {

	u, err := common.Create(common.GlobalDbProvider, user)
	if err != nil {
		panic(err)
	}

	user.ID = u.GetId()

	player := &model.Player{
		UserId: user.ID.Hex(),
		Line:   1,
	}

	p, err := common.Create(common.GlobalDbProvider, player)
	if err != nil {
		panic(err)
	}

	player.ID = p.GetId()
}

func ResetDatabase() {
	common.GlobalDbProvider = NewUnitTestDbProvider()
}
