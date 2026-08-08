package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"intraclub/database"
	"intraclub/model"
)

// This file implements the #56 (Seasons & teams) and #57 (Schedules & weekly
// matchups) SQLite round-trip tests: field-by-field losslessness across
// Create -> GetOne/GetAll -> Update -> Delete on the SQLite provider.
//
// Most records are created through database.CreateOne (which also exercises
// static/dynamic validation). A few records depend on tables from later model
// tickets that don't exist yet, so they are persisted directly through the raw
// provider methods:
//   - TeamRating requires a Rating (table lands in #60)
//   - Week requires a Draft (table lands in #58)
//   - a second Season used to test a FK-field update would collide on the
//     UNIQUE(draft_id) constraint, so it is created raw with a fabricated id
//
// Using the raw provider methods still exercises the exact SQLite column
// mapping (insert/scan) that the round-trip test is about.

// testUserSeq keeps generated emails unique within a single provider instance.
var testUserSeq int

func createTestUser(t *testing.T, p database.Provider) *model.User {
	t.Helper()
	testUserSeq++
	u, err := database.CreateOne(context.Background(), p, &model.User{
		FirstName:   fmt.Sprintf("Test%d", testUserSeq),
		LastName:    "User",
		PhoneNumber: model.PhoneNumber(fmt.Sprintf("770555%04d", testUserSeq)),
		Email:       model.EmailAddress(fmt.Sprintf("roundtrip%04d@example.com", testUserSeq)),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

// createTestSeason creates a Season whose optional references (draft, facility,
// schedule, playoff structure) are all unset, so it passes dynamic validation
// without those tables existing yet.
func createTestSeason(t *testing.T, p database.Provider) *model.Season {
	t.Helper()
	s := model.NewSeason()
	s.Name = "Test Season"
	s.StartTime = model.NewStartTime(8, 30)
	v, err := database.CreateOne(context.Background(), p, s)
	if err != nil {
		t.Fatalf("create season: %v", err)
	}
	return v
}

// createTestSeasonRaw creates a Season directly through the provider with a
// fabricated, distinct DraftId. Used only to obtain a second season for FK-field
// update tests, where two CreateOne seasons would collide on UNIQUE(draft_id).
func createTestSeasonRaw(t *testing.T, p database.Provider) *model.Season {
	t.Helper()
	s := model.NewSeason()
	s.Name = "Raw Season"
	s.StartTime = model.NewStartTime(9, 0)
	s.DraftId = model.DraftId(database.NewRecordId())
	if _, err := p.Create(context.Background(), s); err != nil {
		t.Fatalf("raw create season: %v", err)
	}
	return s
}

// createTestTeam creates a Team with a captain user. Team.PostCreate adds a
// TeamAssignment for the captain as a side effect.
func createTestTeam(t *testing.T, p database.Provider) *model.Team {
	t.Helper()
	captain := createTestUser(t, p)
	team := model.NewDefaultTeam(captain.ID, "Team")
	team.Color = model.Blue
	v, err := database.CreateOne(context.Background(), p, team)
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	return v
}

// createTestWeekForSeason creates a Week directly through the provider (the
// draft table lands in #58) whose DraftId matches the season so that
// WeeklyMatchup dynamic validation passes.
func createTestWeekForSeason(t *testing.T, p database.Provider, season *model.Season) *model.Week {
	t.Helper()
	w := model.NewWeek()
	w.DraftId = season.DraftId
	w.Date = time.Date(2025, 1, 5, 8, 0, 0, 0, time.UTC)
	w.Note = "week"
	if _, err := p.Create(context.Background(), w); err != nil {
		t.Fatalf("raw create week: %v", err)
	}
	return w
}

func timePtrEqual(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}

// ---- #56: Seasons & teams ----

func TestSeasonRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	s := model.NewSeason()
	s.Name = "Intraclub 2025"
	s.StartTime = model.NewStartTime(8, 30)
	created, err := database.CreateOne(ctx, p, s)
	if err != nil {
		t.Fatalf("CreateOne(season): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(season) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Season{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(season): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(season): record not found")
	}
	if got.Name != created.Name || got.Owner != created.Owner ||
		got.Facility != created.Facility || got.DraftId != created.DraftId ||
		got.ScheduleID != created.ScheduleID || got.PlayoffStructure != created.PlayoffStructure ||
		!time.Time(got.StartTime).Equal(time.Time(created.StartTime)) {
		t.Fatalf("season round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.Season](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(season): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(season): got %d records, want 1 matching the created season", len(all))
	}

	created.Name = "Renamed Season"
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(season): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Season{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(season) after update: %v", err)
	}
	if got2.Name != "Renamed Season" {
		t.Fatalf("season update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Season{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(season): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Season{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(season) after delete: %v", err)
	}
	if exists {
		t.Fatal("season should have been deleted")
	}
}

func TestTeamRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	team := model.NewDefaultTeam(createTestUser(t, p).ID, "Team A")
	team.Color = model.Blue
	created, err := database.CreateOne(ctx, p, team)
	if err != nil {
		t.Fatalf("CreateOne(team): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(team) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Team{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(team): record not found")
	}
	if got.Name != created.Name || got.Color != created.Color ||
		!got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) ||
		!timePtrEqual(got.DeletedAt, created.DeletedAt) {
		t.Fatalf("team round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.Team](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(team): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(team): got %d records, want 1 matching the created team", len(all))
	}

	created.Name = "Team B"
	created.Color = model.Green
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(team): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Team{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team) after update: %v", err)
	}
	if got2.Name != "Team B" || got2.Color != model.Green {
		t.Fatalf("team update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Team{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(team): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Team{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team) after delete: %v", err)
	}
	if exists {
		t.Fatal("team should have been deleted")
	}
}

func TestTeamAssignmentRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	team := createTestTeam(t, p)
	member := createTestUser(t, p)
	a := &model.TeamAssignment{
		TeamId:    team.ID,
		UserId:    member.ID,
		Role:      model.TeamRoleMember,
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	created, err := database.CreateOne(ctx, p, a)
	if err != nil {
		t.Fatalf("CreateOne(assignment): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(assignment) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.TeamAssignment{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(assignment): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(assignment): record not found")
	}
	if got.TeamId != created.TeamId || got.UserId != created.UserId || got.Role != created.Role ||
		!got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) ||
		!timePtrEqual(got.DeletedAt, created.DeletedAt) {
		t.Fatalf("assignment round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	// The team's PostCreate added a captain assignment, so assert presence
	// rather than an exact count.
	all, err := database.GetAll[*model.TeamAssignment](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(assignment): %v", err)
	}
	found := false
	for _, x := range all {
		if x.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("GetAll(assignment): created assignment %s not present", created.ID)
	}

	created.Role = model.TeamRoleCoCaptain
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(assignment): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.TeamAssignment{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(assignment) after update: %v", err)
	}
	if got2.Role != model.TeamRoleCoCaptain {
		t.Fatalf("assignment update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.TeamAssignment{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(assignment): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.TeamAssignment{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(assignment) after delete: %v", err)
	}
	if exists {
		t.Fatal("assignment should have been deleted")
	}
}

func TestSeasonCommissionerRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	user := createTestUser(t, p)
	season := createTestSeason(t, p)
	sc := &model.SeasonCommissioner{
		SeasonId:  season.ID,
		UserId:    user.ID,
		CreatedAt: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 2, 2, 0, 0, 0, 0, time.UTC),
	}
	created, err := database.CreateOne(ctx, p, sc)
	if err != nil {
		t.Fatalf("CreateOne(commissioner): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.SeasonCommissioner{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(commissioner): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(commissioner): record not found")
	}
	if got.SeasonId != created.SeasonId || got.UserId != created.UserId ||
		!got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("commissioner round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.SeasonCommissioner](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(commissioner): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(commissioner): got %d records, want 1", len(all))
	}

	user2 := createTestUser(t, p)
	created.UserId = user2.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(commissioner): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.SeasonCommissioner{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(commissioner) after update: %v", err)
	}
	if got2.UserId != user2.ID {
		t.Fatalf("commissioner update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.SeasonCommissioner{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(commissioner): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.SeasonCommissioner{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(commissioner) after delete: %v", err)
	}
	if exists {
		t.Fatal("commissioner should have been deleted")
	}
}

func TestSeasonLateAdditionRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	user := createTestUser(t, p)
	season := createTestSeason(t, p)
	sla := &model.SeasonLateAddition{
		SeasonId:  season.ID,
		UserId:    user.ID,
		CreatedAt: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 3, 2, 0, 0, 0, 0, time.UTC),
	}
	created, err := database.CreateOne(ctx, p, sla)
	if err != nil {
		t.Fatalf("CreateOne(late_addition): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.SeasonLateAddition{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(late_addition): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(late_addition): record not found")
	}
	if got.SeasonId != created.SeasonId || got.UserId != created.UserId ||
		!got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("late_addition round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.SeasonLateAddition](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(late_addition): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(late_addition): got %d records, want 1", len(all))
	}

	user2 := createTestUser(t, p)
	created.UserId = user2.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(late_addition): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.SeasonLateAddition{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(late_addition) after update: %v", err)
	}
	if got2.UserId != user2.ID {
		t.Fatalf("late_addition update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.SeasonLateAddition{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(late_addition): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.SeasonLateAddition{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(late_addition) after delete: %v", err)
	}
	if exists {
		t.Fatal("late_addition should have been deleted")
	}
}

func TestSeasonTeamRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	season := createTestSeason(t, p)
	team := createTestTeam(t, p)
	st := &model.SeasonTeam{
		SeasonId:  season.ID,
		TeamId:    team.ID,
		CreatedAt: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 4, 2, 0, 0, 0, 0, time.UTC),
	}
	created, err := database.CreateOne(ctx, p, st)
	if err != nil {
		t.Fatalf("CreateOne(season_team): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.SeasonTeam{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(season_team): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(season_team): record not found")
	}
	if got.SeasonId != created.SeasonId || got.TeamId != created.TeamId ||
		!got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("season_team round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.SeasonTeam](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(season_team): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(season_team): got %d records, want 1", len(all))
	}

	season2 := createTestSeasonRaw(t, p)
	created.SeasonId = season2.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(season_team): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.SeasonTeam{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(season_team) after update: %v", err)
	}
	if got2.SeasonId != season2.ID {
		t.Fatalf("season_team update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.SeasonTeam{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(season_team): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.SeasonTeam{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(season_team) after delete: %v", err)
	}
	if exists {
		t.Fatal("season_team should have been deleted")
	}
}

func TestTeamRatingRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	team := createTestTeam(t, p)
	member := createTestUser(t, p)
	tr := &model.TeamRating{
		TeamId:   team.ID,
		UserId:   member.ID,
		RatingId: model.RatingId(database.NewRecordId()), // rating table lands in #60
	}
	// Created directly through the provider because TeamRating.DynamicallyValid
	// requires a Rating record, whose table is not migrated yet (#60).
	if _, err := p.Create(ctx, tr); err != nil {
		t.Fatalf("Create(team_rating): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.TeamRating{}, tr.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team_rating): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(team_rating): record not found")
	}
	if got.TeamId != tr.TeamId || got.UserId != tr.UserId || got.RatingId != tr.RatingId {
		t.Fatalf("team_rating round-trip mismatch:\n  got  %+v\n  want %+v", got, tr)
	}

	all, err := database.GetAll[*model.TeamRating](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(team_rating): %v", err)
	}
	if len(all) != 1 || all[0].ID != tr.ID {
		t.Fatalf("GetAll(team_rating): got %d records, want 1", len(all))
	}

	tr.RatingId = model.RatingId(database.NewRecordId())
	if err := p.Update(ctx, tr); err != nil {
		t.Fatalf("Update(team_rating): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.TeamRating{}, tr.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team_rating) after update: %v", err)
	}
	if got2.RatingId != tr.RatingId {
		t.Fatalf("team_rating update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.TeamRating{}, tr.GetId()); err != nil {
		t.Fatalf("DeleteOneById(team_rating): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.TeamRating{}, tr.GetId())
	if err != nil {
		t.Fatalf("GetOneById(team_rating) after delete: %v", err)
	}
	if exists {
		t.Fatal("team_rating should have been deleted")
	}
}

// ---- #57: Schedules & weekly matchups ----

func TestScheduleRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	season := createTestSeason(t, p)
	s := &model.Schedule{SeasonId: season.ID}
	created, err := database.CreateOne(ctx, p, s)
	if err != nil {
		t.Fatalf("CreateOne(schedule): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(schedule) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Schedule{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(schedule): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(schedule): record not found")
	}
	if got.SeasonId != created.SeasonId {
		t.Fatalf("schedule round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.Schedule](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(schedule): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(schedule): got %d records, want 1", len(all))
	}

	season2 := createTestSeasonRaw(t, p)
	created.SeasonId = season2.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(schedule): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Schedule{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(schedule) after update: %v", err)
	}
	if got2.SeasonId != season2.ID {
		t.Fatalf("schedule update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Schedule{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(schedule): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Schedule{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(schedule) after delete: %v", err)
	}
	if exists {
		t.Fatal("schedule should have been deleted")
	}
}

func TestWeekRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	w := model.NewWeek()
	w.DraftId = model.DraftId(database.NewRecordId()) // draft table lands in #58
	w.Date = time.Date(2025, 1, 5, 8, 0, 0, 0, time.UTC)
	w.Note = "Opening week"
	// Created directly through the provider because Week.DynamicallyValid
	// requires a Draft record, whose table is not migrated yet (#58).
	if _, err := p.Create(ctx, w); err != nil {
		t.Fatalf("Create(week): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Week{}, w.GetId())
	if err != nil {
		t.Fatalf("GetOneById(week): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(week): record not found")
	}
	if got.DraftId != w.DraftId || got.Note != w.Note || !got.Date.Equal(w.Date) {
		t.Fatalf("week round-trip mismatch:\n  got  %+v\n  want %+v", got, w)
	}

	all, err := database.GetAll[*model.Week](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(week): %v", err)
	}
	if len(all) != 1 || all[0].ID != w.ID {
		t.Fatalf("GetAll(week): got %d records, want 1", len(all))
	}

	w.Note = "Updated week"
	w.Date = time.Date(2025, 1, 12, 8, 0, 0, 0, time.UTC)
	if err := p.Update(ctx, w); err != nil {
		t.Fatalf("Update(week): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Week{}, w.GetId())
	if err != nil {
		t.Fatalf("GetOneById(week) after update: %v", err)
	}
	if got2.Note != "Updated week" || !got2.Date.Equal(w.Date) {
		t.Fatalf("week update not persisted, got %+v", got2)
	}

	if err := p.Delete(ctx, w); err != nil {
		t.Fatalf("Delete(week): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Week{}, w.GetId())
	if err != nil {
		t.Fatalf("GetOneById(week) after delete: %v", err)
	}
	if exists {
		t.Fatal("week should have been deleted")
	}
}

func TestWeeklyMatchupRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	season := createTestSeason(t, p)
	week := createTestWeekForSeason(t, p, season)
	wm := &model.WeeklyMatchup{WeekId: week.ID, SeasonId: season.ID}
	created, err := database.CreateOne(ctx, p, wm)
	if err != nil {
		t.Fatalf("CreateOne(weekly_matchup): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(weekly_matchup) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.WeeklyMatchup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(weekly_matchup): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(weekly_matchup): record not found")
	}
	if got.WeekId != created.WeekId || got.SeasonId != created.SeasonId {
		t.Fatalf("weekly_matchup round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.WeeklyMatchup](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(weekly_matchup): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(weekly_matchup): got %d records, want 1", len(all))
	}

	week2 := createTestWeekForSeason(t, p, season)
	created.WeekId = week2.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(weekly_matchup): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.WeeklyMatchup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(weekly_matchup) after update: %v", err)
	}
	if got2.WeekId != week2.ID {
		t.Fatalf("weekly_matchup update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.WeeklyMatchup{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(weekly_matchup): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.WeeklyMatchup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(weekly_matchup) after delete: %v", err)
	}
	if exists {
		t.Fatal("weekly_matchup should have been deleted")
	}
}

func TestScheduleMatchupRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	season := createTestSeason(t, p)
	week := createTestWeekForSeason(t, p, season)
	wm, err := database.CreateOne(ctx, p, &model.WeeklyMatchup{WeekId: week.ID, SeasonId: season.ID})
	if err != nil {
		t.Fatalf("CreateOne(weekly_matchup): %v", err)
	}
	sched, err := database.CreateOne(ctx, p, &model.Schedule{SeasonId: season.ID})
	if err != nil {
		t.Fatalf("CreateOne(schedule): %v", err)
	}

	sm := &model.ScheduleMatchup{ScheduleId: sched.ID, WeeklyMatchupId: wm.ID, Position: 0}
	created, err := database.CreateOne(ctx, p, sm)
	if err != nil {
		t.Fatalf("CreateOne(schedule_matchup): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(schedule_matchup) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.ScheduleMatchup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(schedule_matchup): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(schedule_matchup): record not found")
	}
	if got.ScheduleId != created.ScheduleId || got.WeeklyMatchupId != created.WeeklyMatchupId ||
		got.Position != created.Position {
		t.Fatalf("schedule_matchup round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.ScheduleMatchup](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(schedule_matchup): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(schedule_matchup): got %d records, want 1", len(all))
	}

	created.Position = 2
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(schedule_matchup): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.ScheduleMatchup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(schedule_matchup) after update: %v", err)
	}
	if got2.Position != 2 {
		t.Fatalf("schedule_matchup update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.ScheduleMatchup{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(schedule_matchup): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.ScheduleMatchup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(schedule_matchup) after delete: %v", err)
	}
	if exists {
		t.Fatal("schedule_matchup should have been deleted")
	}
}

func TestWeeklyMatchupTeamMatchupRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	season := createTestSeason(t, p)
	week := createTestWeekForSeason(t, p, season)
	wm, err := database.CreateOne(ctx, p, &model.WeeklyMatchup{WeekId: week.ID, SeasonId: season.ID})
	if err != nil {
		t.Fatalf("CreateOne(weekly_matchup): %v", err)
	}
	teamA := createTestTeam(t, p)
	teamB := createTestTeam(t, p)

	wmtm := &model.WeeklyMatchupTeamMatchup{
		WeeklyMatchupId: wm.ID,
		HomeTeamId:      teamA.ID,
		AwayTeamId:      teamB.ID,
		Bye:             false,
		Position:        0,
	}
	created, err := database.CreateOne(ctx, p, wmtm)
	if err != nil {
		t.Fatalf("CreateOne(weekly_matchup_team_matchup): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(weekly_matchup_team_matchup) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.WeeklyMatchupTeamMatchup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(weekly_matchup_team_matchup): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(weekly_matchup_team_matchup): record not found")
	}
	if got.WeeklyMatchupId != created.WeeklyMatchupId ||
		got.HomeTeamId != created.HomeTeamId || got.AwayTeamId != created.AwayTeamId ||
		got.Bye != created.Bye || got.Position != created.Position {
		t.Fatalf("weekly_matchup_team_matchup round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.WeeklyMatchupTeamMatchup](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(weekly_matchup_team_matchup): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(weekly_matchup_team_matchup): got %d records, want 1", len(all))
	}

	created.Position = 3
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(weekly_matchup_team_matchup): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.WeeklyMatchupTeamMatchup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(weekly_matchup_team_matchup) after update: %v", err)
	}
	if got2.Position != 3 {
		t.Fatalf("weekly_matchup_team_matchup update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.WeeklyMatchupTeamMatchup{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(weekly_matchup_team_matchup): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.WeeklyMatchupTeamMatchup{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(weekly_matchup_team_matchup) after delete: %v", err)
	}
	if exists {
		t.Fatal("weekly_matchup_team_matchup should have been deleted")
	}
}
