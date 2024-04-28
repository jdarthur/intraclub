package test

import (
	"intraclub/common"
	"intraclub/model"
	"testing"
)

func init() {
	ResetDatabase()
	InitializeBlueAndGreenTeams()
}

func TestCaptainNotMemberOfTeam(t *testing.T) {
	team := &model.Team{
		LeagueId:  LeagueId,
		Color:     model.Blue,
		CaptainId: AndyLascik.ID.Hex(),
	}

	_, err := common.Create(common.GlobalDbProvider, team)
	if err == nil {
		t.Fatalf("Expected captain not a member error, got none")
	}

	ValidateErrorContains(t, err, "captain")
	ValidateErrorContains(t, err, "is not in player list")
}

func TestCoCaptainNotMemberOfTeam(t *testing.T) {

	team := &model.Team{
		LeagueId:   LeagueId,
		Color:      model.Blue,
		CaptainId:  AndyLascik.ID.Hex(),
		CoCaptains: []string{JdArthur.ID.Hex()},
	}

	err := team.AddPlayerByUserId(common.GlobalDbProvider, AndyLascik.ID, Andy.Line)
	if err != nil {
		t.Fatalf("Error adding player: %v", err)
	}

	_, err = common.Create(common.GlobalDbProvider, team)
	if err == nil {
		t.Fatalf("Expected co-captain not a member error, got none")
	}

	ValidateErrorContains(t, err, "co-captain")
	ValidateErrorContains(t, err, "is not in player list")
}

func TestInvalidPlayerId(t *testing.T) {

	team := &model.Team{
		LeagueId:  LeagueId,
		Color:     model.Blue,
		CaptainId: AndyLascik.ID.Hex(),
	}

	err := team.AddPlayerByUserId(common.GlobalDbProvider, AndyLascik.ID, Andy.Line)
	if err != nil {
		t.Fatalf("Error adding player: %v", err)
	}

	team.Players = append(team.Players, "invalid ID")

	_, err = common.Create(common.GlobalDbProvider, team)
	if err == nil {
		t.Fatalf("Expected error, got none")
	}

	ValidateErrorContains(t, err, "Invalid object ID")
}

func TestInvalidLeagueId(t *testing.T) {

}

func TestInvalidCaptainId(t *testing.T) {

}

func TestValidTeam(t *testing.T) {

	team := blueTeam()
	_, err := common.Create(common.GlobalDbProvider, team)
	if err != nil {
		t.Errorf("Error creating team: %v", err)
	}
}

func TestInvalidCoCaptainId(t *testing.T) {

}
