package model

import (
	"errors"
	"intraclub/common"
	"math"
)

type PlayoffStructureId common.RecordId

func (id PlayoffStructureId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id PlayoffStructureId) String() string {
	return id.RecordId().String()
}

type PlayoffStructure struct {
	ID            PlayoffStructureId // unique ID for this record
	UserId        UserId
	Byes          int // number of teams which get a bye week
	NumberOfTeams int // number of teams which make the playoffs
}

func (p *PlayoffStructure) SetOwner(recordId common.RecordId) {
	p.UserId = UserId(recordId)
}

func (p *PlayoffStructure) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{p.UserId.RecordId()}
}

func (p *PlayoffStructure) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (p *PlayoffStructure) Type() string {
	return "playoff_structure"
}

func (p *PlayoffStructure) GetId() common.RecordId {
	return p.ID.RecordId()
}

func (p *PlayoffStructure) SetId(id common.RecordId) {
	p.ID = PlayoffStructureId(id)
}

func (p *PlayoffStructure) StaticallyValid() error {
	if p.Byes <= 0 {
		return errors.New("number of byes should be greater than zero")
	}
	if p.NumberOfTeams <= 0 {
		return errors.New("number of teams should be greater than zero")
	}

	matchupsInFirstRound := p.NumberOfTeams - p.Byes
	if matchupsInFirstRound%2 != 0 {
		return errors.New("matchups in first round should be even")
	}

	matchupsInSecondRound := (matchupsInFirstRound / 2) + p.Byes
	found := false
	for i := 1; i < 10; i++ {
		if matchupsInSecondRound == int(math.Pow(2, float64(i))) {
			found = true
		}
	}
	if !found {
		return errors.New("invalid matchups in second round (must be a power of 2")
	}

	return nil
}

func (p *PlayoffStructure) NumberOfRounds() int {
	if p.Byes == 0 && p.NumberOfTeams == 2 {
		return 1
	}
	if p.Byes > 0 {
		matchupsInFirstRound := p.NumberOfTeams - p.Byes
		matchupsInSecondRound := (matchupsInFirstRound / 2) + p.Byes
		return int(math.Sqrt(float64(matchupsInSecondRound))) + 1
	}
	return int(math.Sqrt(float64(p.NumberOfTeams))) + 1
}

func (p *PlayoffStructure) DynamicallyValid(db common.DatabaseProvider) error {
	return nil
}
