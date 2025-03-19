package model

import (
	"errors"
	"fmt"
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

func (p *PlayoffStructure) PreUpdate(db common.DatabaseProvider, existingValues common.CrudRecord) error {
	s, err := p.GetAssignedSeasons(db)
	if err != nil {
		return err
	}
	if len(s) != 0 {
		return fmt.Errorf("playoff structure cannot be updated as it has %d assigned season(s)", len(s))
	}
	return nil
}

func (p *PlayoffStructure) PreDelete(db common.DatabaseProvider) error {
	s, err := p.GetAssignedSeasons(db)
	if err != nil {
		return err
	}
	if len(s) != 0 {
		return fmt.Errorf("playoff structure cannot be deleted as it has %d assigned season(s)", len(s))
	}
	return nil
}

func NewPlayoffStructure() *PlayoffStructure {
	return &PlayoffStructure{}
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
	if p.Byes < 0 {
		return errors.New("number of byes should be >= zero")
	}
	if p.NumberOfTeams < 2 {
		return errors.New("number of teams should be >= 2")
	}

	matchupsInFirstRound := p.NumberOfTeams - p.Byes
	if matchupsInFirstRound%2 != 0 {
		return errors.New("matchups in first round should be even")
	}
	if matchupsInFirstRound == 0 {
		return errors.New("matchups in first round should not be zero")
	}

	matchupsInSecondRound := (matchupsInFirstRound / 2) + p.Byes
	found := false
	for i := 1; i < 10; i++ {
		if matchupsInSecondRound == int(math.Pow(2, float64(i))) {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("invalid matchups in second round (must be a power of 2, got %d)", p.NumberOfTeams)
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

		v := math.Log2(float64(matchupsInSecondRound))
		return int(v) + 1
	}
	v := math.Log2(float64(p.NumberOfTeams))
	return int(v)
}

func (p *PlayoffStructure) DynamicallyValid(db common.DatabaseProvider) error {
	return common.ExistsById(db, &User{}, p.UserId.RecordId())
}

func (p *PlayoffStructure) GetAssignedSeasons(db common.DatabaseProvider) ([]*Season, error) {
	return common.GetAllWhere(db, &Season{}, func(c *Season) bool {
		return c.PlayoffStructure == p.ID
	})
}
