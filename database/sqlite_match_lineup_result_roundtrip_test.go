package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"intraclub/database"
	"intraclub/model"
)

// This file implements the #60 (Matches, lineups & results) SQLite round-trip
// tests: field-by-field losslessness across Create -> GetOne/GetAll -> Update
// -> Delete on the SQLite provider for the rating, match, match_editor,
// team_match, team_match_individual_match, lineup, lineup_pairing, and
// availability tables.
//
// The IndividualMatch.Editors slice was normalized into the match_editor child
// table (see model/individual_match.go), so editors are round-tripped on that
// child table rather than as a map/slice field on the match.
//
// A few records are persisted directly through the raw provider methods (which
// still exercise the exact SQLite column mapping) because their DynamicallyValid
// depends on a fully-populated chain that is heavy to set up here:
//   - Week requires a Draft (week is created raw with a fabricated DraftId)
//   - LineupPairing requires team membership and a real format with lines
//
// The remaining records are created through database.CreateOne, which also
// exercises static/dynamic validation since all their referenced tables exist.

// createTestWeek creates a Week directly through the provider with a fabricated
// DraftId (the week is only needed as a reference for lineup/team_match/etc).
func createTestWeek(t *testing.T, p database.Provider) *model.Week {
	t.Helper()
	w := model.NewWeek()
	w.DraftId = model.DraftId(database.NewRecordId())
	w.Date = time.Date(2025, 1, 5, 8, 0, 0, 0, time.UTC)
	w.Note = "week"
	if _, err := p.Create(context.Background(), w); err != nil {
		t.Fatalf("raw create week: %v", err)
	}
	return w
}

// createTestLineup creates a Lineup for the given team and week via
// database.CreateOne (Lineup.DynamicallyValid requires the team and week).
func createTestLineup(t *testing.T, p database.Provider, team *model.Team, week *model.Week) *model.Lineup {
	t.Helper()
	v, err := database.CreateOne(context.Background(), p, &model.Lineup{TeamId: team.ID, WeekId: week.ID})
	if err != nil {
		t.Fatalf("create lineup: %v", err)
	}
	return v
}

// testMatchScoringSeq keeps scoring-structure names unique within a single
// provider instance (ScoringStructure names are unique).
var testMatchScoringSeq int

// createTestMatch creates a single, unpaired IndividualMatch under a real
// scoring structure via database.CreateOne.
func createTestMatch(t *testing.T, p database.Provider) *model.IndividualMatch {
	t.Helper()
	testMatchScoringSeq++
	ss := createTestScoringStructure(t, p, fmt.Sprintf("Standard %d", testMatchScoringSeq), model.Game)
	m := model.NewMatch()
	m.Structure = ss.ID
	v, err := database.CreateOne(context.Background(), p, m)
	if err != nil {
		t.Fatalf("create match: %v", err)
	}
	return v
}

// createTestLineupPairingRaw creates a LineupPairing directly through the
// provider (LineupPairing.DynamicallyValid requires a real format with lines
// and team-membership validation, which is heavy to set up here).
func createTestLineupPairingRaw(t *testing.T, p database.Provider, lineup *model.Lineup, team *model.Team) *model.LineupPairing {
	t.Helper()
	player1 := createTestUser(t, p)
	player2 := createTestUser(t, p)
	lp := &model.LineupPairing{
		LineupId:        lineup.ID,
		TeamId:          team.ID,
		Player1:         player1.ID,
		Player2:         player2.ID,
		FormatLineIndex: 0,
	}
	if _, err := p.Create(context.Background(), lp); err != nil {
		t.Fatalf("raw create lineup_pairing: %v", err)
	}
	return lp
}

// ---- #60: Ratings ----

func TestRatingRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	user := createTestUser(t, p)
	r := model.NewRating()
	r.UserId = user.ID
	r.Name = "Roundtrip Rating 1"
	r.Description = "A rating used by the #60 round-trip test."
	created, err := database.CreateOne(ctx, p, r)
	if err != nil {
		t.Fatalf("CreateOne(rating): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(rating) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Rating{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(rating): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(rating): record not found")
	}
	if got.UserId != created.UserId || got.Name != created.Name || got.Description != created.Description {
		t.Fatalf("rating round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.Rating](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(rating): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(rating): got %d records, want 1", len(all))
	}

	created.Name = "Roundtrip Rating 1 Updated"
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(rating): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Rating{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(rating) after update: %v", err)
	}
	if got2.Name != "Roundtrip Rating 1 Updated" {
		t.Fatalf("rating update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Rating{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(rating): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Rating{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(rating) after delete: %v", err)
	}
	if exists {
		t.Fatal("rating should have been deleted")
	}
}

// ---- #60: Matches ----

func TestIndividualMatchRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	m := model.NewMatch()
	m.Structure = createTestScoringStructure(t, p, "Match Structure", model.Game).ID
	m.MainValue = 6
	m.SecondaryValue = 4
	m.WinOverride = false
	m.Status = model.MatchInProgress
	created, err := database.CreateOne(ctx, p, m)
	if err != nil {
		t.Fatalf("CreateOne(match): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(match) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.IndividualMatch{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(match): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(match): record not found")
	}
	if got.ID != created.ID || got.Opponent != created.Opponent ||
		got.Structure != created.Structure || got.MainValue != created.MainValue ||
		got.SecondaryValue != created.SecondaryValue || got.WinOverride != created.WinOverride ||
		got.Status != created.Status {
		t.Fatalf("match round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.IndividualMatch](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(match): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(match): got %d records, want 1", len(all))
	}

	created.MainValue = 7
	created.WinOverride = true
	created.Status = model.MatchWon
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(match): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.IndividualMatch{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(match) after update: %v", err)
	}
	if got2.MainValue != 7 || !got2.WinOverride || got2.Status != model.MatchWon {
		t.Fatalf("match update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.IndividualMatch{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(match): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.IndividualMatch{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(match) after delete: %v", err)
	}
	if exists {
		t.Fatal("match should have been deleted")
	}
}

func TestMatchEditorRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	match := createTestMatch(t, p)
	editor := createTestUser(t, p)
	me := &model.MatchEditor{MatchId: match.ID, EditorUserId: editor.ID}
	created, err := database.CreateOne(ctx, p, me)
	if err != nil {
		t.Fatalf("CreateOne(match_editor): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(match_editor) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.MatchEditor{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(match_editor): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(match_editor): record not found")
	}
	if got.MatchId != created.MatchId || got.EditorUserId != created.EditorUserId {
		t.Fatalf("match_editor round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.MatchEditor](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(match_editor): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(match_editor): got %d records, want 1", len(all))
	}

	other := createTestUser(t, p)
	created.EditorUserId = other.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(match_editor): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.MatchEditor{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(match_editor) after update: %v", err)
	}
	if got2.EditorUserId != other.ID {
		t.Fatalf("match_editor update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.MatchEditor{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(match_editor): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.MatchEditor{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(match_editor) after delete: %v", err)
	}
	if exists {
		t.Fatal("match_editor should have been deleted")
	}
}

// ---- #60: Team matches ----

func TestTeamMatchRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	week := createTestWeek(t, p)
	home := createTestTeam(t, p)
	away := createTestTeam(t, p)
	lineup := createTestLineup(t, p, home, week)

	tm := &model.TeamMatch{WeekId: week.ID, HomeTeam: home.ID, AwayTeam: away.ID, Lineup: lineup.ID}
	created, err := database.CreateOne(ctx, p, tm)
	if err != nil {
		t.Fatalf("CreateOne(team_match): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(team_match) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.TeamMatch{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team_match): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(team_match): record not found")
	}
	if got.WeekId != created.WeekId || got.HomeTeam != created.HomeTeam ||
		got.AwayTeam != created.AwayTeam || got.Lineup != created.Lineup {
		t.Fatalf("team_match round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.TeamMatch](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(team_match): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(team_match): got %d records, want 1", len(all))
	}

	// Move the match to a new lineup (same week/teams) and verify persistence.
	lineup2 := createTestLineup(t, p, away, week)
	created.Lineup = lineup2.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(team_match): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.TeamMatch{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team_match) after update: %v", err)
	}
	if got2.Lineup != lineup2.ID {
		t.Fatalf("team_match update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.TeamMatch{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(team_match): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.TeamMatch{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team_match) after delete: %v", err)
	}
	if exists {
		t.Fatal("team_match should have been deleted")
	}
}

func TestTeamMatchIndividualMatchRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	week := createTestWeek(t, p)
	home := createTestTeam(t, p)
	away := createTestTeam(t, p)
	lineup := createTestLineup(t, p, home, week)
	teamMatch, err := database.CreateOne(ctx, p, &model.TeamMatch{
		WeekId: week.ID, HomeTeam: home.ID, AwayTeam: away.ID, Lineup: lineup.ID,
	})
	if err != nil {
		t.Fatalf("CreateOne(team_match): %v", err)
	}

	pairing := createTestLineupPairingRaw(t, p, lineup, home)
	match := createTestMatch(t, p)

	row := &model.TeamMatchIndividualMatch{
		TeamMatchId:       teamMatch.ID,
		LineupPairingId:   pairing.ID,
		IndividualMatchId: match.ID,
	}
	created, err := database.CreateOne(ctx, p, row)
	if err != nil {
		t.Fatalf("CreateOne(team_match_individual_match): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(team_match_individual_match) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.TeamMatchIndividualMatch{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team_match_individual_match): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(team_match_individual_match): record not found")
	}
	if got.TeamMatchId != created.TeamMatchId || got.LineupPairingId != created.LineupPairingId ||
		got.IndividualMatchId != created.IndividualMatchId {
		t.Fatalf("team_match_individual_match round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.TeamMatchIndividualMatch](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(team_match_individual_match): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(team_match_individual_match): got %d records, want 1", len(all))
	}

	otherMatch := createTestMatch(t, p)
	created.IndividualMatchId = otherMatch.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(team_match_individual_match): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.TeamMatchIndividualMatch{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team_match_individual_match) after update: %v", err)
	}
	if got2.IndividualMatchId != otherMatch.ID {
		t.Fatalf("team_match_individual_match update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.TeamMatchIndividualMatch{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(team_match_individual_match): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.TeamMatchIndividualMatch{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team_match_individual_match) after delete: %v", err)
	}
	if exists {
		t.Fatal("team_match_individual_match should have been deleted")
	}
}

// ---- #60: Lineups ----

func TestLineupRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	team := createTestTeam(t, p)
	week := createTestWeek(t, p)
	created, err := database.CreateOne(ctx, p, &model.Lineup{TeamId: team.ID, WeekId: week.ID})
	if err != nil {
		t.Fatalf("CreateOne(lineup): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(lineup) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Lineup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(lineup): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(lineup): record not found")
	}
	if got.TeamId != created.TeamId || got.WeekId != created.WeekId {
		t.Fatalf("lineup round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.Lineup](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(lineup): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(lineup): got %d records, want 1", len(all))
	}

	week2 := createTestWeek(t, p)
	created.WeekId = week2.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(lineup): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Lineup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(lineup) after update: %v", err)
	}
	if got2.WeekId != week2.ID {
		t.Fatalf("lineup update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Lineup{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(lineup): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Lineup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(lineup) after delete: %v", err)
	}
	if exists {
		t.Fatal("lineup should have been deleted")
	}
}

func TestLineupPairingRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	team := createTestTeam(t, p)
	week := createTestWeek(t, p)
	lineup := createTestLineup(t, p, team, week)
	lp := createTestLineupPairingRaw(t, p, lineup, team)

	got, exists, err := database.GetOneById(ctx, p, &model.LineupPairing{}, lp.GetId())
	if err != nil {
		t.Fatalf("GetOneById(lineup_pairing): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(lineup_pairing): record not found")
	}
	if got.LineupId != lp.LineupId || got.TeamId != lp.TeamId ||
		got.Player1 != lp.Player1 || got.Player2 != lp.Player2 ||
		got.FormatLineIndex != lp.FormatLineIndex {
		t.Fatalf("lineup_pairing round-trip mismatch:\n  got  %+v\n  want %+v", got, lp)
	}

	all, err := database.GetAll[*model.LineupPairing](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(lineup_pairing): %v", err)
	}
	if len(all) != 1 || all[0].ID != lp.ID {
		t.Fatalf("GetAll(lineup_pairing): got %d records, want 1", len(all))
	}

	lp.FormatLineIndex = 2
	// Updated through the raw provider because LineupPairing.DynamicallyValid
	// requires a real format with lines and team membership (see
	// createTestLineupPairingRaw); the raw update still exercises the exact
	// SQLite column mapping.
	if err := p.Update(ctx, lp); err != nil {
		t.Fatalf("Update(lineup_pairing): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.LineupPairing{}, lp.GetId())
	if err != nil {
		t.Fatalf("GetOneById(lineup_pairing) after update: %v", err)
	}
	if got2.FormatLineIndex != 2 {
		t.Fatalf("lineup_pairing update not persisted, got %+v", got2)
	}

	if err := p.Delete(ctx, lp); err != nil {
		t.Fatalf("Delete(lineup_pairing): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.LineupPairing{}, lp.GetId())
	if err != nil {
		t.Fatalf("GetOneById(lineup_pairing) after delete: %v", err)
	}
	if exists {
		t.Fatal("lineup_pairing should have been deleted")
	}
}

// ---- #60: Availability ----

func TestAvailabilityRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	user := createTestUser(t, p)
	week := createTestWeek(t, p)
	a := model.NewAvailability()
	a.UserId = user.ID
	a.WeekId = week.ID
	a.Available = model.AvailabilityAvailable
	created, err := database.CreateOne(ctx, p, a)
	if err != nil {
		t.Fatalf("CreateOne(availability): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(availability) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Availability{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(availability): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(availability): record not found")
	}
	if got.UserId != created.UserId || got.WeekId != created.WeekId ||
		got.Available != created.Available {
		t.Fatalf("availability round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.Availability](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(availability): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(availability): got %d records, want 1", len(all))
	}

	created.Available = model.AvailabilityMaybe
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(availability): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Availability{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(availability) after update: %v", err)
	}
	if got2.Available != model.AvailabilityMaybe {
		t.Fatalf("availability update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Availability{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(availability): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Availability{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(availability) after delete: %v", err)
	}
	if exists {
		t.Fatal("availability should have been deleted")
	}
}
