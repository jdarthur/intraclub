package test

import (
	"intraclub/middleware"
	"testing"
)

func init() {
	ResetDatabase()
	InitializeBlueAndGreenTeams()
}

func TestAsTeamMember(t *testing.T) {
	team := blueTeam()

	err := middleware.TeamMemberOperation(team, JdArthur.ID.Hex())
	if err != nil {
		t.Error(err)
	}
}

func TestNotAsTeamMember(t *testing.T) {
	team := blueTeam()

	err := middleware.TeamMemberOperation(team, EthanMoland.ID.Hex())
	if err == nil {
		t.Fatalf("Expected non-blue team member to fail 'as team member' check")
	}

	ValidateErrorContains(t, err, "not a member")

}
