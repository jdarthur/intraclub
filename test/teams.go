package test

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"intraclub/model"
)

// dummy IDs for team 1 and team 2

var BlueTeamId = primitive.NewObjectID()
var GreenTeamId = primitive.NewObjectID()

func blueTeam() *model.Team {

	t := &model.Team{
		ID:        BlueTeamId,
		LeagueId:  LeagueId,
		Color:     model.Blue,
		CaptainId: AndyLascik.ID.Hex(),
		CoCaptains: []string{
			JdArthur.ID.Hex(),
		},
	}

	err := t.AddPlayerByUserId(common.GlobalDbProvider, AndyLascik.ID, Andy.Line)
	if err != nil {
		panic(err)
	}

	err = t.AddPlayerByUserId(common.GlobalDbProvider, JdArthur.ID, JD.Line)
	if err != nil {
		panic(err)
	}

	err = t.AddPlayerByUserId(common.GlobalDbProvider, TomEasum.ID, Tom.Line)
	if err != nil {
		panic(err)
	}

	return t
}

func greenTeam() *model.Team {
	t := &model.Team{
		ID:         GreenTeamId,
		LeagueId:   LeagueId,
		Color:      model.Green,
		CaptainId:  TomerWagshal.ID.Hex(),
		CoCaptains: []string{PaulCohen.ID.Hex()},
	}

	err := t.AddPlayerByUserId(common.GlobalDbProvider, TomerWagshal.ID, Tomer.Line)
	if err != nil {
		panic(err)
	}

	err = t.AddPlayerByUserId(common.GlobalDbProvider, PaulCohen.ID, Paul.Line)
	if err != nil {
		panic(err)
	}

	err = t.AddPlayerByUserId(common.GlobalDbProvider, KevinCampbell.ID, Kevin.Line)
	if err != nil {
		panic(err)
	}

	return t
}
