package model

import (
	"context"
	"testing"

	"intraclub/database"
)

func newStoredTeam(t *testing.T, db database.DatabaseProvider, captain database.UserId) *Team {
	team := NewDefaultTeam(captain, "Test Team")

	v, err := database.CreateOne(context.Background(), db, team)
	if err != nil {
		t.Fatal(err)
	}

	// Create team assignment for captain
	assignment := &TeamAssignment{
		TeamId: v.ID,
		UserId: captain,
		Role:   TeamRoleCaptain,
	}
	_, err = database.CreateOne(context.Background(), db, assignment)
	if err != nil {
		t.Fatal(err)
	}

	return v
}
