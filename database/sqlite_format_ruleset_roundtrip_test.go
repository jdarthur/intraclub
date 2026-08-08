package database_test

import (
	"context"
	"testing"
	"time"

	"intraclub/database"
	"intraclub/model"
)

// This file implements the #59 (Formats, rulesets & facilities) SQLite
// round-trip tests: field-by-field losslessness across
// Create -> GetOne/GetAll -> Update -> Delete on the SQLite provider for the
// format, format_rating, format_line, ruleset, rule_section, ruleset_section,
// scoring_structure, scoring_structure_secondary, playoff_structure, and
// facility tables.
//
// A few records are persisted directly through the raw provider methods (which
// still exercise the exact SQLite column mapping) because their DynamicallyValid
// depends on tables that land in later model tickets:
//   - Format requires Rating records (rating table lands in #60)
//   - FormatRating requires a Rating (#60)
//   - FormatLine requires Rating records (#60)
//
// Ruleset uses raw provider update because its PreUpdate hook rejects direct
// modification (use Ruleset.Amend instead); the raw update still exercises the
// exact SQLite column mapping.

// ---- #59: Formats ----

func TestFormatRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	f := model.NewFormat()
	f.UserId = createTestUser(t, p).ID
	f.Name = "Men's Intraclub 1/2/3"
	// Created directly through the provider because Format.DynamicallyValid
	// requires Rating records, whose table is not migrated yet (#60).
	if _, err := p.Create(ctx, f); err != nil {
		t.Fatalf("Create(format): %v", err)
	}
	if f.GetId() == database.InvalidRecordId {
		t.Fatal("Create(format) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Format{}, f.GetId())
	if err != nil {
		t.Fatalf("GetOneById(format): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(format): record not found")
	}
	if got.Name != f.Name || got.UserId != f.UserId || got.ID != f.ID {
		t.Fatalf("format round-trip mismatch:\n  got  %+v\n  want %+v", got, f)
	}

	all, err := database.GetAll[*model.Format](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(format): %v", err)
	}
	if len(all) != 1 || all[0].ID != f.ID {
		t.Fatalf("GetAll(format): got %d records, want 1 matching the created format", len(all))
	}

	f.Name = "Renamed Format"
	if err := p.Update(ctx, f); err != nil {
		t.Fatalf("Update(format): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Format{}, f.GetId())
	if err != nil {
		t.Fatalf("GetOneById(format) after update: %v", err)
	}
	if got2.Name != "Renamed Format" {
		t.Fatalf("format update not persisted, got %+v", got2)
	}

	if err := p.Delete(ctx, f); err != nil {
		t.Fatalf("Delete(format): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Format{}, f.GetId())
	if err != nil {
		t.Fatalf("GetOneById(format) after delete: %v", err)
	}
	if exists {
		t.Fatal("format should have been deleted")
	}
}

func TestFormatRatingRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	format := createTestFormatRaw(t, p)
	fr := &model.FormatRating{
		FormatId:    format.ID,
		RatingId:    model.RatingId(database.NewRecordId()), // rating table lands in #60
		RatingIndex: 1,
	}
	// Created directly through the provider because FormatRating.DynamicallyValid
	// requires a Rating record, whose table is not migrated yet (#60).
	if _, err := p.Create(ctx, fr); err != nil {
		t.Fatalf("Create(format_rating): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.FormatRating{}, fr.GetId())
	if err != nil {
		t.Fatalf("GetOneById(format_rating): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(format_rating): record not found")
	}
	if got.FormatId != fr.FormatId || got.RatingId != fr.RatingId || got.RatingIndex != fr.RatingIndex {
		t.Fatalf("format_rating round-trip mismatch:\n  got  %+v\n  want %+v", got, fr)
	}

	all, err := database.GetAll[*model.FormatRating](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(format_rating): %v", err)
	}
	if len(all) != 1 || all[0].ID != fr.ID {
		t.Fatalf("GetAll(format_rating): got %d records, want 1", len(all))
	}

	fr.RatingIndex = 3
	if err := p.Update(ctx, fr); err != nil {
		t.Fatalf("Update(format_rating): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.FormatRating{}, fr.GetId())
	if err != nil {
		t.Fatalf("GetOneById(format_rating) after update: %v", err)
	}
	if got2.RatingIndex != 3 {
		t.Fatalf("format_rating update not persisted, got %+v", got2)
	}

	if err := p.Delete(ctx, fr); err != nil {
		t.Fatalf("Delete(format_rating): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.FormatRating{}, fr.GetId())
	if err != nil {
		t.Fatalf("GetOneById(format_rating) after delete: %v", err)
	}
	if exists {
		t.Fatal("format_rating should have been deleted")
	}
}

func TestFormatLineRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	format := createTestFormatRaw(t, p)
	fl := &model.FormatLine{
		FormatId:      format.ID,
		FormatIndex:   0,
		Player1Rating: model.RatingId(database.NewRecordId()), // rating table lands in #60
		Player2Rating: model.RatingId(database.NewRecordId()),
	}
	// Created directly through the provider because FormatLine.DynamicallyValid
	// requires Rating records, whose table is not migrated yet (#60).
	if _, err := p.Create(ctx, fl); err != nil {
		t.Fatalf("Create(format_line): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.FormatLine{}, fl.GetId())
	if err != nil {
		t.Fatalf("GetOneById(format_line): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(format_line): record not found")
	}
	if got.FormatId != fl.FormatId || got.FormatIndex != fl.FormatIndex ||
		got.Player1Rating != fl.Player1Rating || got.Player2Rating != fl.Player2Rating {
		t.Fatalf("format_line round-trip mismatch:\n  got  %+v\n  want %+v", got, fl)
	}

	all, err := database.GetAll[*model.FormatLine](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(format_line): %v", err)
	}
	if len(all) != 1 || all[0].ID != fl.ID {
		t.Fatalf("GetAll(format_line): got %d records, want 1", len(all))
	}

	fl.FormatIndex = 4
	if err := p.Update(ctx, fl); err != nil {
		t.Fatalf("Update(format_line): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.FormatLine{}, fl.GetId())
	if err != nil {
		t.Fatalf("GetOneById(format_line) after update: %v", err)
	}
	if got2.FormatIndex != 4 {
		t.Fatalf("format_line update not persisted, got %+v", got2)
	}

	if err := p.Delete(ctx, fl); err != nil {
		t.Fatalf("Delete(format_line): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.FormatLine{}, fl.GetId())
	if err != nil {
		t.Fatalf("GetOneById(format_line) after delete: %v", err)
	}
	if exists {
		t.Fatal("format_line should have been deleted")
	}
}

// createTestFormatRaw creates a Format directly through the provider with a
// real owner user (Format.DynamicallyValid requires Rating records, which land
// in #60).
func createTestFormatRaw(t *testing.T, p database.Provider) *model.Format {
	t.Helper()
	f := model.NewFormat()
	f.UserId = createTestUser(t, p).ID
	f.Name = "Raw Format"
	if _, err := p.Create(context.Background(), f); err != nil {
		t.Fatalf("raw create format: %v", err)
	}
	return f
}

// ---- #59: Rulesets ----

func TestRulesetRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	r := model.NewRuleset()
	r.Name = "Intraclub Rules v1"
	r.Revision = 1
	r.Date = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	r.Owner = createTestUser(t, p).ID
	created, err := database.CreateOne(ctx, p, r)
	if err != nil {
		t.Fatalf("CreateOne(ruleset): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(ruleset) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Ruleset{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(ruleset): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(ruleset): record not found")
	}
	if got.Name != created.Name || got.Revision != created.Revision ||
		got.SupersededBy != created.SupersededBy || got.Owner != created.Owner ||
		!got.Date.Equal(created.Date) {
		t.Fatalf("ruleset round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.Ruleset](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(ruleset): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(ruleset): got %d records, want 1", len(all))
	}

	// Ruleset.PreUpdate rejects direct modification (use Amend instead); the
	// raw provider update still exercises the exact SQLite column mapping.
	created.Revision = 2
	created.Name = "Intraclub Rules v2"
	if err := p.Update(ctx, created); err != nil {
		t.Fatalf("Update(ruleset): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Ruleset{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(ruleset) after update: %v", err)
	}
	if got2.Revision != 2 || got2.Name != "Intraclub Rules v2" {
		t.Fatalf("ruleset update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Ruleset{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(ruleset): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Ruleset{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(ruleset) after delete: %v", err)
	}
	if exists {
		t.Fatal("ruleset should have been deleted")
	}
}

func TestRuleSectionRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	ruleset := createTestRuleset(t, p)
	sec := &model.RuleSection{
		Parent:   ruleset.ID,
		Title:    "Scoring",
		Markdown: "A match is won by winning two of three sets.",
		Owner:    createTestUser(t, p).ID,
	}
	created, err := database.CreateOne(ctx, p, sec)
	if err != nil {
		t.Fatalf("CreateOne(rule_section): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(rule_section) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.RuleSection{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(rule_section): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(rule_section): record not found")
	}
	if got.ID != created.ID || got.Parent != created.Parent ||
		got.Title != created.Title || got.Markdown != created.Markdown ||
		got.Owner != created.Owner {
		t.Fatalf("rule_section round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.RuleSection](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(rule_section): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(rule_section): got %d records, want 1", len(all))
	}

	created.Title = "Updated Scoring"
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(rule_section): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.RuleSection{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(rule_section) after update: %v", err)
	}
	if got2.Title != "Updated Scoring" {
		t.Fatalf("rule_section update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.RuleSection{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(rule_section): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.RuleSection{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(rule_section) after delete: %v", err)
	}
	if exists {
		t.Fatal("rule_section should have been deleted")
	}
}

func TestRulesetSectionRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	ruleset := createTestRuleset(t, p)
	section := createTestRuleSection(t, p, ruleset)
	rs := &model.RulesetSection{
		RulesetId:    ruleset.ID,
		SectionId:    section.ID,
		SectionIndex: 0,
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	created, err := database.CreateOne(ctx, p, rs)
	if err != nil {
		t.Fatalf("CreateOne(ruleset_section): %v", err)
	}

	got, exists, err := database.GetOneById(ctx, p, &model.RulesetSection{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(ruleset_section): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(ruleset_section): record not found")
	}
	if got.RulesetId != created.RulesetId || got.SectionId != created.SectionId ||
		got.SectionIndex != created.SectionIndex ||
		!got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("ruleset_section round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.RulesetSection](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(ruleset_section): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(ruleset_section): got %d records, want 1", len(all))
	}

	created.SectionIndex = 2
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(ruleset_section): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.RulesetSection{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(ruleset_section) after update: %v", err)
	}
	if got2.SectionIndex != 2 {
		t.Fatalf("ruleset_section update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.RulesetSection{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(ruleset_section): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.RulesetSection{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(ruleset_section) after delete: %v", err)
	}
	if exists {
		t.Fatal("ruleset_section should have been deleted")
	}
}

// createTestRuleset creates a Ruleset with a real owner user.
func createTestRuleset(t *testing.T, p database.Provider) *model.Ruleset {
	t.Helper()
	r := model.NewRuleset()
	r.Name = "Test Ruleset"
	r.Owner = createTestUser(t, p).ID
	v, err := database.CreateOne(context.Background(), p, r)
	if err != nil {
		t.Fatalf("create ruleset: %v", err)
	}
	return v
}

// createTestRuleSection creates a RuleSection belonging to the given ruleset.
func createTestRuleSection(t *testing.T, p database.Provider, ruleset *model.Ruleset) *model.RuleSection {
	t.Helper()
	sec := &model.RuleSection{
		Parent:   ruleset.ID,
		Title:    "Section",
		Markdown: "Contents",
		Owner:    createTestUser(t, p).ID,
	}
	v, err := database.CreateOne(context.Background(), p, sec)
	if err != nil {
		t.Fatalf("create rule_section: %v", err)
	}
	return v
}

// ---- #59: Scoring structures ----

func TestScoringStructureRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	s := model.NewScoringStructure()
	s.Name = "Tennis standard match"
	s.Owner = createTestUser(t, p).ID
	s.WinConditionCountingType = model.Game
	s.WinCondition = model.WinCondition{
		WinThreshold:        6,
		MustWinBy:           2,
		InstantWinThreshold: 7,
	}
	created, err := database.CreateOne(ctx, p, s)
	if err != nil {
		t.Fatalf("CreateOne(scoring_structure): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(scoring_structure) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.ScoringStructure{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(scoring_structure): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(scoring_structure): record not found")
	}
	if got.Name != created.Name || got.Owner != created.Owner ||
		got.WinConditionCountingType != created.WinConditionCountingType ||
		got.WinCondition != created.WinCondition {
		t.Fatalf("scoring_structure round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.ScoringStructure](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(scoring_structure): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(scoring_structure): got %d records, want 1", len(all))
	}

	created.Name = "Renamed scoring"
	created.WinCondition.InstantWinThreshold = 9
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(scoring_structure): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.ScoringStructure{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(scoring_structure) after update: %v", err)
	}
	if got2.Name != "Renamed scoring" || got2.WinCondition.InstantWinThreshold != 9 {
		t.Fatalf("scoring_structure update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.ScoringStructure{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(scoring_structure): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.ScoringStructure{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(scoring_structure) after delete: %v", err)
	}
	if exists {
		t.Fatal("scoring_structure should have been deleted")
	}
}

func TestScoringStructureSecondaryRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	parent := createTestScoringStructure(t, p, "Parent", model.Set)
	secondary := createTestScoringStructure(t, p, "Secondary", model.Game)
	ss := &model.ScoringStructureSecondary{
		ScoringStructureId:          parent.ID,
		SecondaryScoringStructureId: secondary.ID,
		SecondaryIndex:              0,
	}
	created, err := database.CreateOne(ctx, p, ss)
	if err != nil {
		t.Fatalf("CreateOne(scoring_structure_secondary): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(scoring_structure_secondary) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.ScoringStructureSecondary{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(scoring_structure_secondary): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(scoring_structure_secondary): record not found")
	}
	if got.ScoringStructureId != created.ScoringStructureId ||
		got.SecondaryScoringStructureId != created.SecondaryScoringStructureId ||
		got.SecondaryIndex != created.SecondaryIndex {
		t.Fatalf("scoring_structure_secondary round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.ScoringStructureSecondary](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(scoring_structure_secondary): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(scoring_structure_secondary): got %d records, want 1", len(all))
	}

	created.SecondaryIndex = 2
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(scoring_structure_secondary): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.ScoringStructureSecondary{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(scoring_structure_secondary) after update: %v", err)
	}
	if got2.SecondaryIndex != 2 {
		t.Fatalf("scoring_structure_secondary update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.ScoringStructureSecondary{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(scoring_structure_secondary): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.ScoringStructureSecondary{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(scoring_structure_secondary) after delete: %v", err)
	}
	if exists {
		t.Fatal("scoring_structure_secondary should have been deleted")
	}
}

// createTestScoringStructure creates a simple (non-composite) scoring
// structure with the given counting type and a real owner user.
func createTestScoringStructure(t *testing.T, p database.Provider, name string, countingType model.ScoreCountingType) *model.ScoringStructure {
	t.Helper()
	s := model.NewScoringStructure()
	s.Name = name
	s.Owner = createTestUser(t, p).ID
	s.WinConditionCountingType = countingType
	s.WinCondition = model.WinCondition{WinThreshold: 6, MustWinBy: 1}
	v, err := database.CreateOne(context.Background(), p, s)
	if err != nil {
		t.Fatalf("create scoring_structure: %v", err)
	}
	return v
}

// ---- #59: Playoff structures ----

func TestPlayoffStructureRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	ps := model.NewPlayoffStructure()
	ps.UserId = createTestUser(t, p).ID
	ps.Byes = 0
	ps.NumberOfTeams = 4
	created, err := database.CreateOne(ctx, p, ps)
	if err != nil {
		t.Fatalf("CreateOne(playoff_structure): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(playoff_structure) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.PlayoffStructure{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(playoff_structure): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(playoff_structure): record not found")
	}
	if got.UserId != created.UserId || got.Byes != created.Byes ||
		got.NumberOfTeams != created.NumberOfTeams {
		t.Fatalf("playoff_structure round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.PlayoffStructure](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(playoff_structure): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(playoff_structure): got %d records, want 1", len(all))
	}

	created.NumberOfTeams = 8
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(playoff_structure): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.PlayoffStructure{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(playoff_structure) after update: %v", err)
	}
	if got2.NumberOfTeams != 8 {
		t.Fatalf("playoff_structure update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.PlayoffStructure{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(playoff_structure): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.PlayoffStructure{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(playoff_structure) after delete: %v", err)
	}
	if exists {
		t.Fatal("playoff_structure should have been deleted")
	}
}

// ---- #59: Facilities ----

func TestFacilityRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	fac := model.NewFacility()
	fac.UserId = createTestUser(t, p).ID
	fac.Name = "Martin's Landing River Club"
	fac.Address = "123 River Rd"
	fac.NumberOfCourts = 4
	created, err := database.CreateOne(ctx, p, fac)
	if err != nil {
		t.Fatalf("CreateOne(facility): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(facility) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Facility{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(facility): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(facility): record not found")
	}
	if got.UserId != created.UserId || got.Name != created.Name ||
		got.Address != created.Address || got.NumberOfCourts != created.NumberOfCourts ||
		got.LayoutPhoto != created.LayoutPhoto {
		t.Fatalf("facility round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.Facility](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(facility): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(facility): got %d records, want 1", len(all))
	}

	created.Name = "Renamed Facility"
	created.NumberOfCourts = 6
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(facility): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Facility{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(facility) after update: %v", err)
	}
	if got2.Name != "Renamed Facility" || got2.NumberOfCourts != 6 {
		t.Fatalf("facility update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Facility{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(facility): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Facility{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(facility) after delete: %v", err)
	}
	if exists {
		t.Fatal("facility should have been deleted")
	}
}
