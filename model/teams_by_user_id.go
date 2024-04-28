package model

import "intraclub/common"

// GetTeamsForUserId gets all of the Team s that this User is a member of
// and calculates the Team.Active value for each team
func GetTeamsForUserId(db common.DbProvider, userId string) ([]*Team, error) {
	list, err := common.GetAll(db, &Team{})
	if err != nil {
		return nil, err
	}

	teams := list.(listOfTeams)

	output := make([]*Team, 0)

	for _, team := range teams {

		userIsOnTeam, err := UserIsOnTeam(db, team, userId)
		if err != nil {
			return nil, err
		}

		if userIsOnTeam {

			// if user is on this team, then we will
			// include it in the response.

			// check if the team is active
			active, err := team.IsActive(db)
			if err != nil {
				return nil, err
			}

			// mark this team as active or inactive based on
			// what we just calculated
			team.Active = active

			output = append(output, team)
		}
	}

	return teams, nil
}

func UserIsOnTeam(db common.DbProvider, team *Team, userId string) (bool, error) {
	players, err := team.GetPlayers(db)
	if err != nil {
		return false, err
	}

	for _, player := range players {
		if player.UserId == userId {
			return true, nil
		}
	}
	return false, nil
}
