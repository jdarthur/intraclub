package model

import "intraclub/common"

func GetCommissionedLeaguesByUserId(db common.DbProvider, userId string) ([]*League, error) {
	v, err := common.GetAll(db, &League{})
	if err != nil {
		return nil, err
	}

	leagues := v.(listOfLeagues)

	output := make([]*League, 0)
	for _, league := range leagues {
		if league.Commissioner == userId {

			active, err := league.IsActive(db)
			if err != nil {
				return nil, err
			}

			league.Active = active

			output = append(output, league)
		}
	}

	return output, nil
}
