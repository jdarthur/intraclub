package controllers

import (
	"fmt"
	"intraclub/common"
	"intraclub/model"
)

type CaptainTeamAssignment map[model.TeamColor]string

func InitializeIntraclubYear(db common.DbProvider, year int, league model.League, assignments CaptainTeamAssignment) ([]model.Team, error) {
	err := validateColorAssignment(db, league, assignments)
	if err != nil {
		return nil, err
	}

	output := make([]model.Team, 0)

	for color, captain := range assignments {
		team := model.NewTeam(color, captain)
		team.Year = year

		output = append(output, team)
	}

	return output, nil
}

func validateColorAssignment(db common.DbProvider, league model.League, assignments CaptainTeamAssignment) error {
	for _, color := range league.Colors {
		captainId, ok := assignments[color]
		if !ok {
			return fmt.Errorf("color %s was not present in color/captain assignment", color)
		}

		err := common.CheckExistenceOrErrorByStringId(db, &model.User{}, captainId)
		if err != nil {
			return fmt.Errorf("captain with id %s does not exist", captainId)
		}
	}

	return nil
}
