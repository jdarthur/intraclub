package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

// WeeklyMatchupTeamMatchupId is the unique identifier for a WeeklyMatchupTeamMatchup record.
type WeeklyMatchupTeamMatchupId database.RecordId

func (id WeeklyMatchupTeamMatchupId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id WeeklyMatchupTeamMatchupId) String() string {
	return id.RecordId().String()
}

// WeeklyMatchupTeamMatchup is a join table record that links a WeeklyMatchup to its
// individual TeamMatchup entries. Each record represents one team's participation
// (as home, away, or bye) in a given weekly matchup.
// This normalizes the previously inline []*TeamMatchup slice into its own collection,
// enabling individual querying and indexing.
type WeeklyMatchupTeamMatchup struct {
	ID              WeeklyMatchupTeamMatchupId `json:"id"`
	WeeklyMatchupId WeeklyMatchupId            `json:"weekly_matchup_id"`
	HomeTeamId      TeamId                     `json:"home_team_id"`
	AwayTeamId      TeamId                     `json:"away_team_id"`
	Bye             bool                       `json:"bye"`
	Position        int                        `json:"position"`
}

// GetOwner returns InvalidUserId as WeeklyMatchupTeamMatchup has no specific owner.
func (w *WeeklyMatchupTeamMatchup) GetOwner() database.UserId {
	return database.InvalidUserId
}

// SetOwner is a no-op as WeeklyMatchupTeamMatchup has no specific owner.
func (w *WeeklyMatchupTeamMatchup) SetOwner(userId database.UserId) {}

// AccessibleTo returns everyone as WeeklyMatchupTeamMatchup records are public.
func (w *WeeklyMatchupTeamMatchup) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

// EditableBy returns sysadmin-only access for join records.
func (w *WeeklyMatchupTeamMatchup) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

// Type returns the record type identifier for WeeklyMatchupTeamMatchup.
func (w *WeeklyMatchupTeamMatchup) Type() string {
	return "weekly_matchup_team_matchup"
}

// GetId returns the unique identifier for this record.
func (w *WeeklyMatchupTeamMatchup) GetId() database.RecordId {
	return w.ID.RecordId()
}

// SetId sets the unique identifier for this record.
func (w *WeeklyMatchupTeamMatchup) SetId(id database.RecordId) {
	w.ID = WeeklyMatchupTeamMatchupId(id)
}

// StaticallyValid always returns nil as there are no static validation rules.
func (w *WeeklyMatchupTeamMatchup) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that the referenced WeeklyMatchup and Team records exist.
func (w *WeeklyMatchupTeamMatchup) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &WeeklyMatchup{}, w.WeeklyMatchupId.RecordId()); err != nil {
		return fmt.Errorf("weekly matchup error: %w", err)
	}
	if err := database.ExistsById(ctx, db, &Team{}, w.HomeTeamId.RecordId()); err != nil {
		return fmt.Errorf("home team error: %w", err)
	}
	if !w.Bye {
		if err := database.ExistsById(ctx, db, &Team{}, w.AwayTeamId.RecordId()); err != nil {
			return fmt.Errorf("away team error: %w", err)
		}
	}
	return nil
}

// UniquenessEquivalent enforces that each team appears at most once per weekly matchup.
// A team may not be a home team in one record and a home, away, or bye team in another
// record of the same weekly matchup.
func (w *WeeklyMatchupTeamMatchup) UniquenessEquivalent(other *WeeklyMatchupTeamMatchup) error {
	if w.WeeklyMatchupId != other.WeeklyMatchupId {
		return nil
	}

	// Compare this record's teams against both teams of the other record.
	if w.HomeTeamId == other.HomeTeamId || w.HomeTeamId == other.AwayTeamId {
		return fmt.Errorf("home team %s already has a matchup in weekly matchup %s", w.HomeTeamId, w.WeeklyMatchupId)
	}
	if !w.Bye {
		if w.AwayTeamId == other.HomeTeamId || w.AwayTeamId == other.AwayTeamId {
			return fmt.Errorf("away team %s already has a matchup in weekly matchup %s", w.AwayTeamId, w.WeeklyMatchupId)
		}
	}
	return nil
}

// NewRecord returns a new empty WeeklyMatchupTeamMatchup.
func (w *WeeklyMatchupTeamMatchup) NewRecord() database.CrudRecord {
	return new(WeeklyMatchupTeamMatchup)
}

// NewWeeklyMatchupTeamMatchup creates a new WeeklyMatchupTeamMatchup record.
func NewWeeklyMatchupTeamMatchup(weeklyMatchupId WeeklyMatchupId, homeTeamId TeamId, awayTeamId TeamId, bye bool, position int) *WeeklyMatchupTeamMatchup {
	return &WeeklyMatchupTeamMatchup{
		WeeklyMatchupId: weeklyMatchupId,
		HomeTeamId:      homeTeamId,
		AwayTeamId:      awayTeamId,
		Bye:             bye,
		Position:        position,
	}
}
