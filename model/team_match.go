package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

type TeamMatchId database.RecordId

func (id TeamMatchId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id TeamMatchId) String() string {
	return id.RecordId().String()
}

// TeamMatch represents a head-to-head match between a home and an away team
// for a given week, using a specific lineup to determine the pairings played.
//
// API/JSON shape decision: the former inline `individual_matches` field
// (`map[LineupPairingId]IndividualMatchId`) has been removed from `TeamMatch`
// and normalized into the `TeamMatchIndividualMatch` join table. TeamMatch
// records are not currently exposed via a REST CRUD route (see main.go), so
// removing the field is not a breaking wire change; in-process reads can
// reassemble the relationship rows into the old map shape via
// `TeamMatch.GetIndividualMatches`.
type TeamMatch struct {
	ID       TeamMatchId
	WeekId   WeekId
	HomeTeam TeamId
	AwayTeam TeamId
	Lineup   LineupId
}

func (t *TeamMatch) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (t *TeamMatch) SetOwner(userId database.UserId) {
	// no owner field; access is governed by the participating teams
}

func (t *TeamMatch) Type() string {
	return "team_match"
}

func (t *TeamMatch) GetId() database.RecordId {
	return t.ID.RecordId()
}

func (t *TeamMatch) SetId(id database.RecordId) {
	t.ID = TeamMatchId(id)
}

func (t *TeamMatch) StaticallyValid() error {
	return nil
}

func (t *TeamMatch) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Week{}, t.WeekId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &Team{}, t.HomeTeam.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &Team{}, t.AwayTeam.RecordId()); err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &Lineup{}, t.Lineup.RecordId())
}

func (t *TeamMatch) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (t *TeamMatch) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (t *TeamMatch) NewRecord() database.CrudRecord {
	return new(TeamMatch)
}

// TeamMatchIndividualMatch is a join table record that assigns an
// IndividualMatch to a LineupPairing for a particular TeamMatch. This replaces
// the former denormalized `TeamMatch.IndividualMatches` inline map (lineup
// pairing -> individual match), enabling the relationship to be queried/indexed
// individually and stored in its own collection/table.
//
// It is stored in its own collection (`team_match_individual_match`) with FKs
// to team_match, lineup_pairing, and individual_match, and a natural unique
// constraint on (TeamMatchId, LineupPairingId). When the #36 SQLite provider
// lands, this becomes a `team_match_individual_matches` table with a
// migration/backfill.
type TeamMatchIndividualMatch struct {
	ID                database.RecordId `json:"id"`
	TeamMatchId       TeamMatchId       `json:"team_match_id"`
	LineupPairingId   LineupPairingId   `json:"lineup_pairing_id"`
	IndividualMatchId IndividualMatchId `json:"individual_match_id"`
}

func (m *TeamMatchIndividualMatch) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (m *TeamMatchIndividualMatch) SetOwner(userId database.UserId) {}

func (m *TeamMatchIndividualMatch) Type() string {
	return "team_match_individual_match"
}

func (m *TeamMatchIndividualMatch) GetId() database.RecordId {
	return m.ID
}

func (m *TeamMatchIndividualMatch) SetId(id database.RecordId) {
	m.ID = id
}

func (m *TeamMatchIndividualMatch) StaticallyValid() error {
	return nil
}

func (m *TeamMatchIndividualMatch) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &TeamMatch{}, m.TeamMatchId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &LineupPairing{}, m.LineupPairingId.RecordId()); err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &IndividualMatch{}, m.IndividualMatchId.RecordId())
}

func (m *TeamMatchIndividualMatch) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (m *TeamMatchIndividualMatch) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (m *TeamMatchIndividualMatch) NewRecord() database.CrudRecord {
	return new(TeamMatchIndividualMatch)
}

// UniquenessEquivalent enforces the natural unique constraint on
// (TeamMatchId, LineupPairingId): a lineup pairing may only be assigned one
// individual match per team match.
func (m *TeamMatchIndividualMatch) UniquenessEquivalent(other *TeamMatchIndividualMatch) error {
	if m.TeamMatchId == other.TeamMatchId && m.LineupPairingId == other.LineupPairingId {
		return fmt.Errorf("lineup pairing %s already has an individual match assigned on team match %s", m.LineupPairingId, m.TeamMatchId)
	}
	return nil
}

// getIndividualMatchRows returns all TeamMatchIndividualMatch relationship
// rows assigned to this team match.
func (t *TeamMatch) getIndividualMatchRows(ctx context.Context, db database.Provider) ([]*TeamMatchIndividualMatch, error) {
	filter := func(_ context.Context, m *TeamMatchIndividualMatch) bool {
		return m.TeamMatchId == t.ID
	}
	return database.GetAllWhere[*TeamMatchIndividualMatch](ctx, db, filter)
}

// GetIndividualMatches reassembles the relationship rows into the former
// inline map shape (lineup pairing -> individual match) for in-process reads.
func (t *TeamMatch) GetIndividualMatches(ctx context.Context, db database.Provider) (map[LineupPairingId]IndividualMatchId, error) {
	rows, err := t.getIndividualMatchRows(ctx, db)
	if err != nil {
		return nil, err
	}
	result := make(map[LineupPairingId]IndividualMatchId, len(rows))
	for _, row := range rows {
		result[row.LineupPairingId] = row.IndividualMatchId
	}
	return result, nil
}

// AssignIndividualMatch creates the relationship row that assigns the given
// IndividualMatch to the given LineupPairing for this TeamMatch.
func (t *TeamMatch) AssignIndividualMatch(ctx context.Context, db database.Provider, lineupPairingId LineupPairingId, individualMatchId IndividualMatchId) (*TeamMatchIndividualMatch, error) {
	row := &TeamMatchIndividualMatch{
		TeamMatchId:       t.ID,
		LineupPairingId:   lineupPairingId,
		IndividualMatchId: individualMatchId,
	}
	return database.CreateOne(ctx, db, row)
}

// ValidateMatchesVsLineup verifies that every individual match assigned to
// this TeamMatch is tied to a lineup pairing that belongs to this team match's
// lineup, and that each referenced individual match exists.
func (t *TeamMatch) ValidateMatchesVsLineup(ctx context.Context, db database.Provider) error {
	rows, err := t.getIndividualMatchRows(ctx, db)
	if err != nil {
		return err
	}

	for _, row := range rows {
		pairing, err := database.GetExistingRecordById(ctx, db, &LineupPairing{}, row.LineupPairingId.RecordId())
		if err != nil {
			return err
		}
		if pairing.LineupId != t.Lineup {
			return fmt.Errorf("lineup pairing %s is not part of lineup %s for team match %s", row.LineupPairingId, t.Lineup, t.ID)
		}
		if _, err := database.GetExistingRecordById(ctx, db, &IndividualMatch{}, row.IndividualMatchId.RecordId()); err != nil {
			return err
		}
	}

	return nil
}
