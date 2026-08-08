package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

type WeeklyMatchupId database.RecordId

func (id WeeklyMatchupId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id WeeklyMatchupId) String() string {
	return id.RecordId().String()
}

// TeamMatchup is a lightweight value type representing a single matchup entry
// (home vs away, or a bye). It is no longer stored inline within WeeklyMatchup;
// instead each entry is normalized into a WeeklyMatchupTeamMatchup record.
// This type remains for API serialization and convenience constructors.
type TeamMatchup struct {
	HomeTeam TeamId
	AwayTeam TeamId
	Bye      bool
}

// Validate checks that the matchup entry is well-formed: bye entries have no away team,
// and all referenced teams exist and are assigned to the given season.
func (t *TeamMatchup) Validate(ctx context.Context, db database.Provider, season *Season) error {
	if t.Bye && t.AwayTeam != TeamId(database.InvalidRecordId) {
		return fmt.Errorf("away team ID must not be set during a bye")
	}

	err := database.ExistsById(ctx, db, &Team{}, t.HomeTeam.RecordId())
	if err != nil {
		return fmt.Errorf("home team error: %s", err)
	}

	if !season.IsTeamAssignedToSeason(ctx, db, t.HomeTeam) {
		return fmt.Errorf("home team %s is not assigned to season %s", t.HomeTeam, season.ID)
	}

	if !t.Bye {
		err = database.ExistsById(ctx, db, &Team{}, t.AwayTeam.RecordId())
		if err != nil {
			return fmt.Errorf("away team error: %s", err)
		}

		if !season.IsTeamAssignedToSeason(ctx, db, t.AwayTeam) {
			return fmt.Errorf("away team %s is not assigned to season %s", t.AwayTeam, season.ID)
		}
	}
	return nil
}

// teamMatchupFromRecord converts a WeeklyMatchupTeamMatchup row into a TeamMatchup value.
func teamMatchupFromRecord(wmtm *WeeklyMatchupTeamMatchup) *TeamMatchup {
	return &TeamMatchup{
		HomeTeam: wmtm.HomeTeamId,
		AwayTeam: wmtm.AwayTeamId,
		Bye:      wmtm.Bye,
	}
}

// WeeklyMatchup is an instance of one or more TeamMatchup entries for a given Week
// during a Season's Schedule. The individual matchups are stored as separate
// WeeklyMatchupTeamMatchup records rather than inline.
type WeeklyMatchup struct {
	ID       WeeklyMatchupId `json:"id"`
	WeekId   WeekId   // Week that this WeeklyMatchup corresponds to, i.e. a particular date
	SeasonId SeasonId // Season that this WeeklyMatchup corresponds to
}

func (w *WeeklyMatchup) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (w *WeeklyMatchup) UniquenessEquivalent(other *WeeklyMatchup) error {
	if w.SeasonId == other.SeasonId && w.WeekId == other.WeekId {
		return fmt.Errorf("duplicate record for season ID and week ID")
	}
	return nil
}

func NewWeeklyMatchup() *WeeklyMatchup {
	return &WeeklyMatchup{}
}

func (w *WeeklyMatchup) Type() string {
	return "weekly_matchup"
}

func (w *WeeklyMatchup) GetId() database.RecordId {
	return w.ID.RecordId()
}

func (w *WeeklyMatchup) SetId(id database.RecordId) {
	w.ID = WeeklyMatchupId(id)
}

func (w *WeeklyMatchup) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return EditableBySeason(ctx, db, w.SeasonId)
}

func (w *WeeklyMatchup) AccessibleTo(_ context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (w *WeeklyMatchup) SetOwner(userId database.UserId) {
	// don't need to do anything here as the ownership of the
	// WeeklyMatchup record type is automatically inferred &
	// enforced by the associated Season assigned to it
}

// StaticallyValid always returns nil — double-booking checks now require database access
// since matchups are stored in a separate WeeklyMatchupTeamMatchup table.
// Those checks are performed in DynamicallyValid via validateNoDoubleBooking.
func (w *WeeklyMatchup) StaticallyValid() error {
	return nil
}

func (w *WeeklyMatchup) DynamicallyValid(ctx context.Context, db database.Provider) error {
	week, err := database.GetExistingRecordById(ctx, db, &Week{}, w.WeekId.RecordId())
	if err != nil {
		return err
	}

	// validate that the season in question exists.
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, w.SeasonId.RecordId())
	if err != nil {
		return err
	}

	// validate that the weekly matchup
	if week.DraftId != season.DraftId {
		return fmt.Errorf("draft %s assigned to season %s does not match draft %s assigned to week", week.Date, w.SeasonId, week.DraftId)
	}

	matchups, err := w.GetMatchups(ctx, db)
	if err != nil {
		return err
	}

	// skip matchup-level checks when no matchups have been created yet
	// (matchups are separate records with a FK, so they are created after the WeeklyMatchup)
	if len(matchups) == 0 {
		return nil
	}

	// validate that each individual matchup is valid
	for _, matchup := range matchups {
		err = matchup.Validate(ctx, db, season)
		if err != nil {
			return err
		}
	}

	// check for double-booked teams
	if err := w.validateNoDoubleBooking(matchups); err != nil {
		return err
	}

	return w.ValidateThatEachTeamHasOneMatchup(ctx, db, season)
}

// validateNoDoubleBooking ensures no team appears in more than one matchup within this weekly matchup.
func (w *WeeklyMatchup) validateNoDoubleBooking(matchups []*TeamMatchup) error {
	if len(matchups) < 2 {
		return nil
	}
	for i, match := range matchups[:len(matchups)-1] {
		for j, match2 := range matchups[i+1:] {
			if match.HomeTeam == match2.HomeTeam || match.HomeTeam == match2.AwayTeam {
				return fmt.Errorf("home team %s from matchup %d is also playing in matchup %d", match.HomeTeam, i, j+i+1)
			}
			if !match.Bye {
				if match.AwayTeam == match2.AwayTeam || match.AwayTeam == match2.HomeTeam {
					return fmt.Errorf("away team %s from matchup %d is also playing in matchup %d", match.AwayTeam, i, j+i+1)
				}
			}
		}
	}
	return nil
}

func (w *WeeklyMatchup) ValidateThatEachTeamHasOneMatchup(ctx context.Context, db database.Provider, season *Season) error {
	matchups, err := w.GetMatchups(ctx, db)
	if err != nil {
		return err
	}
	m := make(map[TeamId]bool)
	for _, matchup := range matchups {
		m[matchup.HomeTeam] = true
		if !matchup.Bye {
			m[matchup.AwayTeam] = true
		}
	}

	teams, err := season.GetTeams(ctx, db)
	if err != nil {
		return err
	}

	if len(m) != len(teams) {
		for _, team := range teams {
			_, ok := m[team.ID]
			if !ok {
				return fmt.Errorf("team %s does not have a matchup or bye", team.ID)
			}
		}
	}
	return nil
}

func (w *WeeklyMatchup) NewRecord() database.CrudRecord {
	return new(WeeklyMatchup)
}

// GetMatchups retrieves all TeamMatchup entries for this WeeklyMatchup by querying
// the WeeklyMatchupTeamMatchup relationship table, sorted by Position.
func (w *WeeklyMatchup) GetMatchups(ctx context.Context, db database.Provider) ([]*TeamMatchup, error) {
	records, err := database.GetAllWhere[*WeeklyMatchupTeamMatchup](ctx, db, func(_ context.Context, wmtm *WeeklyMatchupTeamMatchup) bool {
		return wmtm.WeeklyMatchupId == w.ID
	})
	if err != nil {
		return nil, err
	}

	// Sort by Position to preserve ordering
	for i := 0; i < len(records); i++ {
		for j := i + 1; j < len(records); j++ {
			if records[j].Position < records[i].Position {
				records[i], records[j] = records[j], records[i]
			}
		}
	}

	matchups := make([]*TeamMatchup, 0, len(records))
	for _, wmtm := range records {
		matchups = append(matchups, teamMatchupFromRecord(wmtm))
	}
	return matchups, nil
}

// SetMatchups replaces all matchup entries for this WeeklyMatchup.
// It deletes existing WeeklyMatchupTeamMatchup records and creates new ones.
func (w *WeeklyMatchup) SetMatchups(ctx context.Context, db database.Provider, matchups []*TeamMatchup) error {
	// Delete existing matchup records
	existing, err := database.GetAllWhere[*WeeklyMatchupTeamMatchup](ctx, db, func(_ context.Context, wmtm *WeeklyMatchupTeamMatchup) bool {
		return wmtm.WeeklyMatchupId == w.ID
	})
	if err != nil {
		return err
	}
	for _, wmtm := range existing {
		_, _, err = database.DeleteOneById(ctx, db, &WeeklyMatchupTeamMatchup{}, wmtm.ID.RecordId())
		if err != nil {
			return err
		}
	}

	// Create new matchup records
	for i, matchup := range matchups {
		wmtm := NewWeeklyMatchupTeamMatchup(w.ID, matchup.HomeTeam, matchup.AwayTeam, matchup.Bye, i)
		_, err = database.CreateOne(ctx, db, wmtm)
		if err != nil {
			return err
		}
	}
	return nil
}

// setMatchupsRaw creates matchup records directly without validation.
// Used for test scenarios with intentionally invalid data.
func (w *WeeklyMatchup) setMatchupsRaw(ctx context.Context, db database.Provider, matchups []*TeamMatchup) error {
	// Delete existing matchup records
	existing, err := database.GetAllWhere[*WeeklyMatchupTeamMatchup](ctx, db, func(_ context.Context, wmtm *WeeklyMatchupTeamMatchup) bool {
		return wmtm.WeeklyMatchupId == w.ID
	})
	if err != nil {
		return err
	}
	for _, wmtm := range existing {
		_, _, err = database.DeleteOneById(ctx, db, &WeeklyMatchupTeamMatchup{}, wmtm.ID.RecordId())
		if err != nil {
			return err
		}
	}

	// Create new matchup records directly, bypassing validation
	for i, matchup := range matchups {
		wmtm := NewWeeklyMatchupTeamMatchup(w.ID, matchup.HomeTeam, matchup.AwayTeam, matchup.Bye, i)
		wmtm.SetId(database.NewRecordId())
		_, err = db.Create(ctx, wmtm)
		if err != nil {
			return err
		}
	}
	return nil
}

// PreDelete cascades deletion to all associated WeeklyMatchupTeamMatchup records.
func (w *WeeklyMatchup) PreDelete(ctx context.Context, db database.Provider) error {
	existing, err := database.GetAllWhere[*WeeklyMatchupTeamMatchup](ctx, db, func(_ context.Context, wmtm *WeeklyMatchupTeamMatchup) bool {
		return wmtm.WeeklyMatchupId == w.ID
	})
	if err != nil {
		return err
	}
	for _, wmtm := range existing {
		_, _, err = database.DeleteOneById(ctx, db, &WeeklyMatchupTeamMatchup{}, wmtm.ID.RecordId())
		if err != nil {
			return err
		}
	}
	return nil
}
