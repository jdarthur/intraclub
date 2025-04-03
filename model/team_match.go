package model

import "intraclub/common"

type TeamMatchId common.RecordId

func (id TeamMatchId) RecordId() common.RecordId {
	return common.RecordId(id)
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

//func (t *TeamMatch) ValidateMatchesVsLineup(db common.DatabaseProvider) error {
//
//	lineup, err := common.GetExistingRecordById(db, &Lineup{}, t.Lineup.RecordId())
//	if err != nil {
//		return err
//	}
//
//	for lineupPairingId, individualMatchId := range t.IndividualMatches {
//		lineupPairing, err := common.GetExistingRecordById(db, &LineupPairing{}, lineupPairingId.RecordId())
//		if err != nil {
//			return err
//		}
//
//		individualMatch, err := common.GetExistingRecordById(db, &IndividualMatch{}, individualMatchId.RecordId())
//		if err != nil {
//			return err
//		}
//
//		if individualMatch
//
//	}
//
//}
