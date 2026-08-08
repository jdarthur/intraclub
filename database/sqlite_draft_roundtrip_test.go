package database_test

import (
	"context"
	"testing"
	"time"

	"intraclub/database"
	"intraclub/model"
)

// This file implements the #58 (Drafts) SQLite round-trip tests:
// field-by-field losslessness across Create -> GetOne/GetAll -> Update -> Delete
// on the SQLite provider for the draft, draft_available_player, draft_captain,
// draft_format, draft_pick, draft_rating_cutoff, and pre_draft_grade tables.
//
// The Draft.RatingCutoffs inline map has already been normalized into the
// draft_rating_cutoff join table (see model/draft.go), so the round-trip is
// exercised on that child table rather than a map field.
//
// Records whose DynamicallyValid depends on tables that land in later model
// tickets are persisted directly through the raw provider methods (which still
// exercise the exact SQLite column mapping):
//   - Draft requires a Format (table lands in #59)
//   - DraftFormat requires a Format (#59)
//   - DraftRatingCutoff requires a Rating (table lands in #60)
//   - PreDraftGrade requires a Format (#59), a Rating (#60), and a non-empty
//     available-player list
//
// The remaining records (draft_available_player, draft_captain, draft_pick) are
// created through database.CreateOne, which also exercises static/dynamic
// validation since all their referenced tables now exist.

// createTestDraftRaw creates a Draft directly through the provider with a real
// owner user but a fabricated Format (format table lands in #59).
func createTestDraftRaw(t *testing.T, p database.Provider) *model.Draft {
	t.Helper()
	d := model.NewDraft()
	d.Name = "Test Draft"
	d.Owner = createTestUser(t, p).ID
	d.Format = model.FormatId(database.NewRecordId()) // format table lands in #59
	d.CompletedAt = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	d.DraftOrderPattern = model.DraftOrderPatternSnake{}
	if _, err := p.Create(context.Background(), d); err != nil {
		t.Fatalf("raw create draft: %v", err)
	}
	return d
}

// ---- #58: Drafts ----

func TestDraftRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	d := model.NewDraft()
	d.Name = "Intraclub 2025"
	d.Owner = createTestUser(t, p).ID
	d.Format = model.FormatId(database.NewRecordId()) // format table lands in #59
	d.CompletedAt = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	d.DraftOrderPattern = model.DraftOrderPatternLastPickDouble{}
	// Created directly through the provider because Draft.DynamicallyValid
	// requires a Format record, whose table is not migrated yet (#59).
	// p.Create mutates d in place, assigning its RecordId.
	if _, err := p.Create(ctx, d); err != nil {
		t.Fatalf("Create(draft): %v", err)
	}
	if d.GetId() == database.InvalidRecordId {
		t.Fatal("Create(draft) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Draft{}, d.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(draft): record not found")
	}
	if got.Name != d.Name || got.Owner != d.Owner ||
		got.Format != d.Format || !got.CompletedAt.Equal(d.CompletedAt) ||
		got.DraftOrderPattern.Name() != "Last pick double" {
		t.Fatalf("draft round-trip mismatch:\n  got  %+v\n  want %+v", got, d)
	}

	all, err := database.GetAll[*model.Draft](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(draft): %v", err)
	}
	if len(all) != 1 || all[0].ID != d.ID {
		t.Fatalf("GetAll(draft): got %d records, want 1 matching the created draft", len(all))
	}

	d.Name = "Renamed Draft"
	d.DraftOrderPattern = model.DraftOrderPatternStraightUp{}
	d.CompletedAt = time.Date(2025, 1, 2, 13, 0, 0, 0, time.UTC)
	if err := p.Update(ctx, d); err != nil {
		t.Fatalf("Update(draft): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Draft{}, d.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft) after update: %v", err)
	}
	if got2.Name != "Renamed Draft" || got2.DraftOrderPattern.Name() != "Straight-up" ||
		!got2.CompletedAt.Equal(d.CompletedAt) {
		t.Fatalf("draft update not persisted, got %+v", got2)
	}

	if err := p.Delete(ctx, d); err != nil {
		t.Fatalf("Delete(draft): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Draft{}, d.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft) after delete: %v", err)
	}
	if exists {
		t.Fatal("draft should have been deleted")
	}
}

func TestDraftAvailablePlayerRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	draft := createTestDraftRaw(t, p)
	player := createTestUser(t, p)
	dap := &model.DraftAvailablePlayer{
		DraftId:  draft.ID,
		PlayerId: player.ID,
	}
	created, err := database.CreateOne(ctx, p, dap)
	if err != nil {
		t.Fatalf("CreateOne(draft_available_player): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(draft_available_player) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.DraftAvailablePlayer{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_available_player): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(draft_available_player): record not found")
	}
	if got.DraftId != created.DraftId || got.PlayerId != created.PlayerId ||
		!got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("draft_available_player round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.DraftAvailablePlayer](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(draft_available_player): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(draft_available_player): got %d records, want 1", len(all))
	}

	player2 := createTestUser(t, p)
	created.PlayerId = player2.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(draft_available_player): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.DraftAvailablePlayer{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_available_player) after update: %v", err)
	}
	if got2.PlayerId != player2.ID {
		t.Fatalf("draft_available_player update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.DraftAvailablePlayer{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(draft_available_player): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.DraftAvailablePlayer{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_available_player) after delete: %v", err)
	}
	if exists {
		t.Fatal("draft_available_player should have been deleted")
	}
}

func TestDraftCaptainRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	draft := createTestDraftRaw(t, p)
	captain := createTestUser(t, p)
	team := model.NewDefaultTeam(captain.ID, "Captains")
	team.Color = model.Blue
	team, err := database.CreateOne(ctx, p, team)
	if err != nil {
		t.Fatalf("CreateOne(team): %v", err)
	}

	dc := &model.DraftCaptain{
		DraftId:    draft.ID,
		TeamId:     team.ID,
		CaptainId:  captain.ID,
		DraftOrder: 0,
	}
	created, err := database.CreateOne(ctx, p, dc)
	if err != nil {
		t.Fatalf("CreateOne(draft_captain): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(draft_captain) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.DraftCaptain{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_captain): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(draft_captain): record not found")
	}
	if got.DraftId != created.DraftId || got.TeamId != created.TeamId ||
		got.CaptainId != created.CaptainId || got.DraftOrder != created.DraftOrder ||
		!got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("draft_captain round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.DraftCaptain](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(draft_captain): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(draft_captain): got %d records, want 1", len(all))
	}

	created.DraftOrder = 2
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(draft_captain): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.DraftCaptain{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_captain) after update: %v", err)
	}
	if got2.DraftOrder != 2 {
		t.Fatalf("draft_captain update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.DraftCaptain{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(draft_captain): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.DraftCaptain{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_captain) after delete: %v", err)
	}
	if exists {
		t.Fatal("draft_captain should have been deleted")
	}
}

func TestDraftFormatRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	draft := createTestDraftRaw(t, p)
	df := &model.DraftFormat{
		DraftId:  draft.ID,
		FormatId: model.FormatId(database.NewRecordId()), // format table lands in #59
	}
	// Created directly through the provider because DraftFormat.DynamicallyValid
	// requires a Format record, whose table is not migrated yet (#59).
	if _, err := p.Create(ctx, df); err != nil {
		t.Fatalf("Create(draft_format): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.DraftFormat{}, df.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_format): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(draft_format): record not found")
	}
	if got.DraftId != df.DraftId || got.FormatId != df.FormatId {
		t.Fatalf("draft_format round-trip mismatch:\n  got  %+v\n  want %+v", got, df)
	}

	all, err := database.GetAll[*model.DraftFormat](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(draft_format): %v", err)
	}
	if len(all) != 1 || all[0].ID != df.ID {
		t.Fatalf("GetAll(draft_format): got %d records, want 1", len(all))
	}

	df.FormatId = model.FormatId(database.NewRecordId())
	if err := p.Update(ctx, df); err != nil {
		t.Fatalf("Update(draft_format): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.DraftFormat{}, df.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_format) after update: %v", err)
	}
	if got2.FormatId != df.FormatId {
		t.Fatalf("draft_format update not persisted, got %+v", got2)
	}

	if err := p.Delete(ctx, df); err != nil {
		t.Fatalf("Delete(draft_format): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.DraftFormat{}, df.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_format) after delete: %v", err)
	}
	if exists {
		t.Fatal("draft_format should have been deleted")
	}
}

func TestDraftPickRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	draft := createTestDraftRaw(t, p)
	team := createTestTeam(t, p)
	user := createTestUser(t, p)
	dp := &model.DraftPick{
		DraftId: draft.ID,
		TeamId:  team.ID,
		UserId:  user.ID,
		Round:   1,
		Pick:    1,
		Rating:  model.RatingId(database.NewRecordId()), // rating table lands in #60
	}
	created, err := database.CreateOne(ctx, p, dp)
	if err != nil {
		t.Fatalf("CreateOne(draft_pick): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(draft_pick) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.DraftPick{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_pick): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(draft_pick): record not found")
	}
	if got.DraftId != created.DraftId || got.TeamId != created.TeamId ||
		got.UserId != created.UserId || got.Round != created.Round ||
		got.Pick != created.Pick || got.Rating != created.Rating ||
		!got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("draft_pick round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.DraftPick](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(draft_pick): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(draft_pick): got %d records, want 1", len(all))
	}

	created.Pick = 4
	created.Round = 2
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(draft_pick): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.DraftPick{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_pick) after update: %v", err)
	}
	if got2.Pick != 4 || got2.Round != 2 {
		t.Fatalf("draft_pick update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.DraftPick{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(draft_pick): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.DraftPick{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_pick) after delete: %v", err)
	}
	if exists {
		t.Fatal("draft_pick should have been deleted")
	}
}

func TestDraftRatingCutoffRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	draft := createTestDraftRaw(t, p)
	drc := &model.DraftRatingCutoff{
		DraftId:     draft.ID,
		RatingId:    model.RatingId(database.NewRecordId()), // rating table lands in #60
		CutoffIndex: 8,
	}
	// Created directly through the provider because DraftRatingCutoff.DynamicallyValid
	// requires a Rating record, whose table is not migrated yet (#60).
	if _, err := p.Create(ctx, drc); err != nil {
		t.Fatalf("Create(draft_rating_cutoff): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.DraftRatingCutoff{}, drc.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_rating_cutoff): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(draft_rating_cutoff): record not found")
	}
	if got.DraftId != drc.DraftId || got.RatingId != drc.RatingId ||
		got.CutoffIndex != drc.CutoffIndex {
		t.Fatalf("draft_rating_cutoff round-trip mismatch:\n  got  %+v\n  want %+v", got, drc)
	}

	all, err := database.GetAll[*model.DraftRatingCutoff](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(draft_rating_cutoff): %v", err)
	}
	if len(all) != 1 || all[0].ID != drc.ID {
		t.Fatalf("GetAll(draft_rating_cutoff): got %d records, want 1", len(all))
	}

	drc.CutoffIndex = 16
	if err := p.Update(ctx, drc); err != nil {
		t.Fatalf("Update(draft_rating_cutoff): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.DraftRatingCutoff{}, drc.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_rating_cutoff) after update: %v", err)
	}
	if got2.CutoffIndex != 16 {
		t.Fatalf("draft_rating_cutoff update not persisted, got %+v", got2)
	}

	if err := p.Delete(ctx, drc); err != nil {
		t.Fatalf("Delete(draft_rating_cutoff): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.DraftRatingCutoff{}, drc.GetId())
	if err != nil {
		t.Fatalf("GetOneById(draft_rating_cutoff) after delete: %v", err)
	}
	if exists {
		t.Fatal("draft_rating_cutoff should have been deleted")
	}
}

func TestPreDraftGradeRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	draft := createTestDraftRaw(t, p)
	player := createTestUser(t, p)
	grader := createTestUser(t, p)
	grade := &model.PreDraftGrade{
		PlayerId: player.ID,
		DraftId:  draft.ID,
		GraderId: grader.ID,
		Modifier: model.StrongModifier,
		Rating:   model.RatingId(database.NewRecordId()), // rating table lands in #60
	}
	// Created directly through the provider because PreDraftGrade.DynamicallyValid
	// requires Format (#59) and Rating (#60) records and a non-empty draft list.
	if _, err := p.Create(ctx, grade); err != nil {
		t.Fatalf("Create(pre_draft_grade): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.PreDraftGrade{}, grade.GetId())
	if err != nil {
		t.Fatalf("GetOneById(pre_draft_grade): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(pre_draft_grade): record not found")
	}
	if got.PlayerId != grade.PlayerId || got.DraftId != grade.DraftId ||
		got.GraderId != grade.GraderId || got.Modifier != grade.Modifier ||
		got.Rating != grade.Rating {
		t.Fatalf("pre_draft_grade round-trip mismatch:\n  got  %+v\n  want %+v", got, grade)
	}

	all, err := database.GetAll[*model.PreDraftGrade](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(pre_draft_grade): %v", err)
	}
	if len(all) != 1 || all[0].ID != grade.ID {
		t.Fatalf("GetAll(pre_draft_grade): got %d records, want 1", len(all))
	}

	grade.Modifier = model.WeakModifier
	if err := p.Update(ctx, grade); err != nil {
		t.Fatalf("Update(pre_draft_grade): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.PreDraftGrade{}, grade.GetId())
	if err != nil {
		t.Fatalf("GetOneById(pre_draft_grade) after update: %v", err)
	}
	if got2.Modifier != model.WeakModifier {
		t.Fatalf("pre_draft_grade update not persisted, got %+v", got2)
	}

	if err := p.Delete(ctx, grade); err != nil {
		t.Fatalf("Delete(pre_draft_grade): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.PreDraftGrade{}, grade.GetId())
	if err != nil {
		t.Fatalf("GetOneById(pre_draft_grade) after delete: %v", err)
	}
	if exists {
		t.Fatal("pre_draft_grade should have been deleted")
	}
}
