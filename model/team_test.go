package model

import (
	"context"
	"intraclub/common"
	"testing"
)

func newStoredTeam(t *testing.T, db common.DatabaseProvider, captain UserId) *Team {
	team := NewDefaultTeam(captain, "Test Team")

	v, err := common.CreateOne(context.Background(), db, team)
	if err != nil {
		t.Fatal(err)
	}

	// Create team assignment for captain
	assignment := &TeamAssignment{
		TeamId: v.ID,
		UserId: captain,
		Role:   TeamRoleCaptain,
	}
	_, err = common.CreateOne(context.Background(), db, assignment)
	if err != nil {
		t.Fatal(err)
	}

	return v
}
