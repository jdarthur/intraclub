package model

import (
	"intraclub/database"
)

type TeamMatchId database.RecordId

func (id TeamMatchId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id TeamMatchId) String() string {
	return id.RecordId().String()
}

type TeamMatch struct {
	ID                TeamMatchId
	WeekId            WeekId
	HomeTeam          TeamId
	AwayTeam          TeamId
	Lineup            LineupId
	IndividualMatches map[LineupPairingId]IndividualMatchId
}

// func (t *TeamMatch) ValidateMatchesVsLineup(db common.DatabaseProvider) error {
//
//	lineup, err := common.GetExistingRecordById(ctx, db, &Lineup{}, t.Lineup.RecordId())
//	if err != nil {
//		return err
//	}
//
//	for lineupPairingId, individualMatchId := range t.IndividualMatches {
//		lineupPairing, err := common.GetExistingRecordById(ctx, db, &LineupPairing{}, lineupPairingId.RecordId())
//		if err != nil {
//			return err
//		}
//
//		individualMatch, err := common.GetExistingRecordById(ctx, db, &IndividualMatch{}, individualMatchId.RecordId())
//		if err != nil {
//			return err
//		}
//
//		if individualMatch
//
//	}
//
// }
