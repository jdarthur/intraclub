package model

import (
	"context"
	"errors"
	"fmt"
	"math"

	"intraclub/database"
)

type PlayoffStructureId database.RecordId

func (id PlayoffStructureId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id PlayoffStructureId) String() string {
	return id.RecordId().String()
}

type PlayoffStructure struct {
	ID            PlayoffStructureId // unique ID for this record
	UserId        database.UserId
	Byes          int // number of teams which get a bye week
	NumberOfTeams int // number of teams which make the playoffs
}

func (p *PlayoffStructure) GetOwner() database.UserId {
	return p.UserId
}

func (p *PlayoffStructure) PreUpdate(ctx context.Context, db database.Provider, existingValues database.CrudRecord) error {
	s, err := p.GetAssignedSeasons(ctx, db)
	if err != nil {
		return err
	}
	if len(s) != 0 {
		return fmt.Errorf("playoff structure cannot be updated as it has %d assigned season(s)", len(s))
	}
	return nil
}

func (p *PlayoffStructure) PreDelete(ctx context.Context, db database.Provider) error {
	s, err := p.GetAssignedSeasons(ctx, db)
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

func (p *PlayoffStructure) SetOwner(userId database.UserId) {
	p.UserId = userId
}

func (p *PlayoffStructure) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{p.UserId}
}

func (p *PlayoffStructure) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (p *PlayoffStructure) Type() string {
	return "playoff_structure"
}

func (p *PlayoffStructure) GetId() database.RecordId {
	return p.ID.RecordId()
}

func (p *PlayoffStructure) SetId(id database.RecordId) {
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

func (p *PlayoffStructure) DynamicallyValid(ctx context.Context, db database.Provider) error {
	return database.ExistsById(ctx, db, &User{}, p.UserId.RecordId())
}

func (p *PlayoffStructure) GetAssignedSeasons(ctx context.Context, db database.Provider) ([]*Season, error) {
	return database.GetAllWhere[*Season](ctx, db, func(_ context.Context, c *Season) bool {
		return c.PlayoffStructure == p.ID
	})
}

func (p *PlayoffStructure) NewRecord() database.CrudRecord {
	return new(PlayoffStructure)
}
