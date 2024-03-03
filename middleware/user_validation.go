package middleware

import (
	"fmt"
	"intraclub/model"
)

// CommissionerOperation validates that the provided ID is a commissioner of the
// given model.League. This is done to protect operations such as creating teams /
// league years, signing scorecards, and adding weekly recaps and photos
func CommissionerOperation(league model.League, commissioner string) error {
	if commissioner != league.Commissioner {
		return fmt.Errorf("user %s is not the commisioner of league %s", commissioner, league.ID)
	}

	return nil
}

// CaptainOperation validates that the provided user ID is the captain of the team. This
// is done to protect captain-specific operations such as adding and removing co-captains
func CaptainOperation(team model.Team, captainId string) error {
	if captainId != team.CaptainId {
		return fmt.Errorf("user %s is not the captain of league %s", captainId, team.ID)
	}

	return nil
}

// CaptainOrCoCaptainOperation validates that the provided user ID is the captain or a co-captain
// of the provided team. This is done to protect operations such as setting weekly lineups.
func CaptainOrCoCaptainOperation(team model.Team, userId string) error {
	if userId == team.CaptainId {
		return nil
	}

	for _, coCaptain := range team.CoCaptains {
		if userId == coCaptain {
			return nil
		}
	}

	return fmt.Errorf("user %s is not a captain or co-captain of league %s", userId, team.ID)
}

// TeamMemberOperation validates that a particular user ID is a member of the given team. This is used
// to protect team-specific API endpoints, e.g. viewing weekly availability
func TeamMemberOperation(team model.Team, userId string) error {
	for _, teamMember := range team.Players {
		if userId == teamMember.UserId {
			return nil
		}
	}

	return fmt.Errorf("user %s is not a captain or co-captain of league %s", userId, team.ID)
}
