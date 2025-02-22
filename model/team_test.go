package model

import (
	"intraclub/common"
	"testing"
)

func newStoredTeam(t *testing.T, db common.DatabaseProvider, captain UserId) *Team {
	team := NewTeam()
	team.Captain = captain
	team.Members = append(team.Members, captain)

	v, err := common.CreateOne(db, team)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
