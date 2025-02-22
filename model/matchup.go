package model

import "intraclub/common"

type Matchup struct {
	ID             common.RecordId
	HomeTeam       TeamId
	HomeTeamLineup common.RecordId
	AwayTeam       TeamId
	AwayTeamLineup common.RecordId
}
