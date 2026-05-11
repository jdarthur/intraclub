package model

import (
	"intraclub/database"
)

type Matchup struct {
	ID             database.RecordId
	HomeTeam       TeamId
	HomeTeamLineup database.RecordId
	AwayTeam       TeamId
	AwayTeamLineup database.RecordId
}
