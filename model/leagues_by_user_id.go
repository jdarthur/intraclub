package model

import "intraclub/common"

func GetLeaguesByUserId(db common.DbProvider, userId string) ([]*League, error) {
	v, err := common.GetAll(db, &League{})
	if err != nil {
		return nil, err
	}

	leagues := v.(listOfLeagues)

	output := make([]*League, 0)
	for _, league := range leagues {
		teams, err := league.GetTeamsForLeague(db)
		if err != nil {
			return nil, err
		}

		for _, team := range teams {

			userIsOnTeam, err := UserIsOnTeam(db, team, userId)
			if err != nil {
				return nil, err
			}

			if userIsOnTeam {

				// check if this league is currently active
				active, err := league.IsActive(db)
				if err != nil {
					return nil, err
				}

				league.Active = active

				leagues = append(leagues, league)
			}
		}
	}

	return output, nil
}
