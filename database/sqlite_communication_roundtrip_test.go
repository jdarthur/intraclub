package database_test

import (
	"bytes"
	"context"
	"testing"

	"intraclub/database"
	"intraclub/model"
)

// This file implements the #61 (Communications & proposals) SQLite round-trip
// tests: field-by-field losslessness across Create -> GetOne/GetAll -> Update
// -> Delete on the SQLite provider for the photo, blurb, blurb_photo, comment,
// blurb_reaction, comment_reaction, commissioner_proposal, and
// commissioner_proposal_vote tables.
//
// The former inline Blurb.Photos / Blurb.Reactions and Comment.Reactions
// slices were normalized into the blurb_photo / blurb_reaction /
// comment_reaction child tables (migrations 0042-0049), so each of those
// relationship records gets its own round-trip test alongside its parent.
//
// Comment.DynamicallyValid requires its owner to be a Season participant, so
// the comment owner is registered as a Season commissioner (a participant)
// before the comment is created.

// createTestPhoto creates a Photo with random contents owned by the given user.
func createTestPhoto(t *testing.T, p database.Provider, owner *model.User) *model.Photo {
	t.Helper()
	photo := model.NewPhoto()
	photo.Owner = owner.ID
	photo.AltText = "photo alt text"
	photo.Contents = []byte{0x89, 0x50, 0x4e, 0x47, 0x00, 0x01, 0x02, 0xff}
	photo.FileType = model.PhotoTypePng
	v, err := database.CreateOne(context.Background(), p, photo)
	if err != nil {
		t.Fatalf("create photo: %v", err)
	}
	return v
}

// createTestBlurb creates a Blurb owned by the given user in the given season.
func createTestBlurb(t *testing.T, p database.Provider, owner *model.User, season *model.Season) *model.Blurb {
	t.Helper()
	b := model.NewBlurb()
	b.Owner = owner.ID
	b.Season = season.ID
	b.Title = "Welcome"
	b.Content = "Welcome to the season"
	v, err := database.CreateOne(context.Background(), p, b)
	if err != nil {
		t.Fatalf("create blurb: %v", err)
	}
	return v
}

// createTestSeasonCommissioner registers the given user as a commissioner
// (and therefore a participant) of the given season.
func createTestSeasonCommissioner(t *testing.T, p database.Provider, season *model.Season, user *model.User) {
	t.Helper()
	if _, err := database.CreateOne(context.Background(), p, &model.SeasonCommissioner{
		SeasonId: season.ID,
		UserId:   user.ID,
	}); err != nil {
		t.Fatalf("create season_commissioner: %v", err)
	}
}

// createTestComment creates a Comment from the given user on the given blurb.
func createTestComment(t *testing.T, p database.Provider, blurb *model.Blurb, owner *model.User) *model.Comment {
	t.Helper()
	c := model.NewComment()
	c.Blurb = blurb.ID
	c.Owner = owner.ID
	c.Content = "Great blurb!"
	v, err := database.CreateOne(context.Background(), p, c)
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	return v
}

// createTestProposal creates a CommissionerProposal for the given season.
func createTestProposal(t *testing.T, p database.Provider, season *model.Season) *model.CommissionerProposal {
	t.Helper()
	prop := model.NewCommissionerProposal()
	prop.Description = "Add a player to a team"
	prop.SeasonId = season.ID
	prop.MustBeUnanimous = true
	v, err := database.CreateOne(context.Background(), p, prop)
	if err != nil {
		t.Fatalf("create commissioner_proposal: %v", err)
	}
	return v
}

// ---- #61: Photos ----

func TestPhotoRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)
	owner := createTestUser(t, p)

	photo := model.NewPhoto()
	photo.Owner = owner.ID
	photo.AltText = "court layout"
	photo.Contents = []byte{0x89, 0x50, 0x4e, 0x47, 0x00, 0x01, 0x02, 0xff}
	photo.FileType = model.PhotoTypeJpeg
	created, err := database.CreateOne(ctx, p, photo)
	if err != nil {
		t.Fatalf("CreateOne(photo): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(photo) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.Photo{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(photo): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(photo): record not found")
	}
	if got.Owner != created.Owner || got.AltText != created.AltText ||
		got.FileType != created.FileType || !bytes.Equal(got.Contents, created.Contents) {
		t.Fatalf("photo round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.Photo](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(photo): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(photo): got %d records, want 1", len(all))
	}

	created.AltText = "renamed layout"
	created.Contents = []byte{0xde, 0xad, 0xbe, 0xef}
	created.FileType = model.PhotoTypeGif
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(photo): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Photo{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(photo) after update: %v", err)
	}
	if got2.AltText != "renamed layout" || got2.FileType != model.PhotoTypeGif ||
		!bytes.Equal(got2.Contents, []byte{0xde, 0xad, 0xbe, 0xef}) {
		t.Fatalf("photo update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Photo{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(photo): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Photo{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(photo) after delete: %v", err)
	}
	if exists {
		t.Fatal("photo should have been deleted")
	}
}

// ---- #61: Blurbs ----

func TestBlurbRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)
	owner := createTestUser(t, p)
	season := createTestSeason(t, p)

	blurb := createTestBlurb(t, p, owner, season)

	got, exists, err := database.GetOneById(ctx, p, &model.Blurb{}, blurb.GetId())
	if err != nil {
		t.Fatalf("GetOneById(blurb): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(blurb): record not found")
	}
	if got.Owner != blurb.Owner || got.Season != blurb.Season ||
		got.Title != blurb.Title || got.Content != blurb.Content {
		t.Fatalf("blurb round-trip mismatch:\n  got  %+v\n  want %+v", got, blurb)
	}

	all, err := database.GetAll[*model.Blurb](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(blurb): %v", err)
	}
	if len(all) != 1 || all[0].ID != blurb.ID {
		t.Fatalf("GetAll(blurb): got %d records, want 1", len(all))
	}

	blurb.Title = "Renamed"
	blurb.Content = "Updated content"
	if err := database.UpdateOne(ctx, p, blurb); err != nil {
		t.Fatalf("UpdateOne(blurb): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Blurb{}, blurb.GetId())
	if err != nil {
		t.Fatalf("GetOneById(blurb) after update: %v", err)
	}
	if got2.Title != "Renamed" || got2.Content != "Updated content" {
		t.Fatalf("blurb update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Blurb{}, blurb.GetId()); err != nil {
		t.Fatalf("DeleteOneById(blurb): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Blurb{}, blurb.GetId())
	if err != nil {
		t.Fatalf("GetOneById(blurb) after delete: %v", err)
	}
	if exists {
		t.Fatal("blurb should have been deleted")
	}
}

func TestBlurbPhotoRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)
	owner := createTestUser(t, p)
	season := createTestSeason(t, p)
	blurb := createTestBlurb(t, p, owner, season)
	photo := createTestPhoto(t, p, owner)
	photo2 := createTestPhoto(t, p, owner)

	row := &model.BlurbPhoto{BlurbId: blurb.ID, PhotoId: photo.ID}
	created, err := database.CreateOne(ctx, p, row)
	if err != nil {
		t.Fatalf("CreateOne(blurb_photo): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(blurb_photo) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.BlurbPhoto{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(blurb_photo): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(blurb_photo): record not found")
	}
	if got.BlurbId != created.BlurbId || got.PhotoId != created.PhotoId {
		t.Fatalf("blurb_photo round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.BlurbPhoto](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(blurb_photo): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(blurb_photo): got %d records, want 1", len(all))
	}

	created.PhotoId = photo2.ID
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(blurb_photo): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.BlurbPhoto{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(blurb_photo) after update: %v", err)
	}
	if got2.PhotoId != photo2.ID {
		t.Fatalf("blurb_photo update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.BlurbPhoto{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(blurb_photo): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.BlurbPhoto{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(blurb_photo) after delete: %v", err)
	}
	if exists {
		t.Fatal("blurb_photo should have been deleted")
	}
}

func TestBlurbReactionRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)
	owner := createTestUser(t, p)
	season := createTestSeason(t, p)
	blurb := createTestBlurb(t, p, owner, season)
	reactor := createTestUser(t, p)

	row := &model.BlurbReaction{BlurbId: blurb.ID, UserId: reactor.ID, ReactionType: model.ThumbsUp}
	created, err := database.CreateOne(ctx, p, row)
	if err != nil {
		t.Fatalf("CreateOne(blurb_reaction): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(blurb_reaction) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.BlurbReaction{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(blurb_reaction): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(blurb_reaction): record not found")
	}
	if got.BlurbId != created.BlurbId || got.UserId != created.UserId ||
		got.ReactionType != created.ReactionType {
		t.Fatalf("blurb_reaction round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.BlurbReaction](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(blurb_reaction): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(blurb_reaction): got %d records, want 1", len(all))
	}

	created.ReactionType = model.Fire
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(blurb_reaction): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.BlurbReaction{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(blurb_reaction) after update: %v", err)
	}
	if got2.ReactionType != model.Fire {
		t.Fatalf("blurb_reaction update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.BlurbReaction{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(blurb_reaction): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.BlurbReaction{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(blurb_reaction) after delete: %v", err)
	}
	if exists {
		t.Fatal("blurb_reaction should have been deleted")
	}
}

// ---- #61: Comments ----

func TestCommentRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)
	owner := createTestUser(t, p)
	season := createTestSeason(t, p)
	// the comment owner must be a Season participant, so register them as a
	// commissioner
	createTestSeasonCommissioner(t, p, season, owner)
	blurb := createTestBlurb(t, p, owner, season)

	comment := createTestComment(t, p, blurb, owner)

	got, exists, err := database.GetOneById(ctx, p, &model.Comment{}, comment.GetId())
	if err != nil {
		t.Fatalf("GetOneById(comment): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(comment): record not found")
	}
	if got.Blurb != comment.Blurb || got.ReplyTo != comment.ReplyTo ||
		got.Owner != comment.Owner || got.Content != comment.Content ||
		!got.CreatedAt.Equal(comment.CreatedAt) {
		t.Fatalf("comment round-trip mismatch:\n  got  %+v\n  want %+v", got, comment)
	}

	all, err := database.GetAll[*model.Comment](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(comment): %v", err)
	}
	if len(all) != 1 || all[0].ID != comment.ID {
		t.Fatalf("GetAll(comment): got %d records, want 1", len(all))
	}

	comment.Content = "Updated comment"
	if err := database.UpdateOne(ctx, p, comment); err != nil {
		t.Fatalf("UpdateOne(comment): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Comment{}, comment.GetId())
	if err != nil {
		t.Fatalf("GetOneById(comment) after update: %v", err)
	}
	if got2.Content != "Updated comment" {
		t.Fatalf("comment update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.Comment{}, comment.GetId()); err != nil {
		t.Fatalf("DeleteOneById(comment): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Comment{}, comment.GetId())
	if err != nil {
		t.Fatalf("GetOneById(comment) after delete: %v", err)
	}
	if exists {
		t.Fatal("comment should have been deleted")
	}
}

func TestCommentReactionRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)
	owner := createTestUser(t, p)
	season := createTestSeason(t, p)
	createTestSeasonCommissioner(t, p, season, owner)
	blurb := createTestBlurb(t, p, owner, season)
	comment := createTestComment(t, p, blurb, owner)
	reactor := createTestUser(t, p)

	row := &model.CommentReaction{CommentId: comment.ID, UserId: reactor.ID, ReactionType: model.Heart}
	created, err := database.CreateOne(ctx, p, row)
	if err != nil {
		t.Fatalf("CreateOne(comment_reaction): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(comment_reaction) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.CommentReaction{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(comment_reaction): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(comment_reaction): record not found")
	}
	if got.CommentId != created.CommentId || got.UserId != created.UserId ||
		got.ReactionType != created.ReactionType {
		t.Fatalf("comment_reaction round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.CommentReaction](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(comment_reaction): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(comment_reaction): got %d records, want 1", len(all))
	}

	created.ReactionType = model.Laughing
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(comment_reaction): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.CommentReaction{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(comment_reaction) after update: %v", err)
	}
	if got2.ReactionType != model.Laughing {
		t.Fatalf("comment_reaction update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.CommentReaction{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(comment_reaction): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.CommentReaction{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(comment_reaction) after delete: %v", err)
	}
	if exists {
		t.Fatal("comment_reaction should have been deleted")
	}
}

// ---- #61: Commissioner proposals ----

func TestCommissionerProposalRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)
	season := createTestSeason(t, p)

	proposal := createTestProposal(t, p, season)

	got, exists, err := database.GetOneById(ctx, p, &model.CommissionerProposal{}, proposal.GetId())
	if err != nil {
		t.Fatalf("GetOneById(commissioner_proposal): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(commissioner_proposal): record not found")
	}
	if got.Description != proposal.Description || got.SeasonId != proposal.SeasonId ||
		got.MustBeUnanimous != proposal.MustBeUnanimous {
		t.Fatalf("commissioner_proposal round-trip mismatch:\n  got  %+v\n  want %+v", got, proposal)
	}

	all, err := database.GetAll[*model.CommissionerProposal](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(commissioner_proposal): %v", err)
	}
	if len(all) != 1 || all[0].ID != proposal.ID {
		t.Fatalf("GetAll(commissioner_proposal): got %d records, want 1", len(all))
	}

	proposal.Description = "Change the playoff structure"
	if err := database.UpdateOne(ctx, p, proposal); err != nil {
		t.Fatalf("UpdateOne(commissioner_proposal): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.CommissionerProposal{}, proposal.GetId())
	if err != nil {
		t.Fatalf("GetOneById(commissioner_proposal) after update: %v", err)
	}
	if got2.Description != "Change the playoff structure" {
		t.Fatalf("commissioner_proposal update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.CommissionerProposal{}, proposal.GetId()); err != nil {
		t.Fatalf("DeleteOneById(commissioner_proposal): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.CommissionerProposal{}, proposal.GetId())
	if err != nil {
		t.Fatalf("GetOneById(commissioner_proposal) after delete: %v", err)
	}
	if exists {
		t.Fatal("commissioner_proposal should have been deleted")
	}
}

func TestCommissionerProposalVoteRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)
	season := createTestSeason(t, p)
	proposal := createTestProposal(t, p, season)
	voter := createTestUser(t, p)

	row := &model.CommissionerProposalVote{ProposalId: proposal.ID, UserId: voter.ID, Vote: true}
	created, err := database.CreateOne(ctx, p, row)
	if err != nil {
		t.Fatalf("CreateOne(commissioner_proposal_vote): %v", err)
	}
	if created.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(commissioner_proposal_vote) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.CommissionerProposalVote{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(commissioner_proposal_vote): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(commissioner_proposal_vote): record not found")
	}
	if got.ProposalId != created.ProposalId || got.UserId != created.UserId || got.Vote != created.Vote {
		t.Fatalf("commissioner_proposal_vote round-trip mismatch:\n  got  %+v\n  want %+v", got, created)
	}

	all, err := database.GetAll[*model.CommissionerProposalVote](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(commissioner_proposal_vote): %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("GetAll(commissioner_proposal_vote): got %d records, want 1", len(all))
	}

	created.Vote = false
	if err := database.UpdateOne(ctx, p, created); err != nil {
		t.Fatalf("UpdateOne(commissioner_proposal_vote): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.CommissionerProposalVote{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(commissioner_proposal_vote) after update: %v", err)
	}
	if got2.Vote {
		t.Fatalf("commissioner_proposal_vote update not persisted, got %+v", got2)
	}

	if _, _, err := database.DeleteOneById(ctx, p, &model.CommissionerProposalVote{}, created.GetId()); err != nil {
		t.Fatalf("DeleteOneById(commissioner_proposal_vote): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.CommissionerProposalVote{}, created.GetId())
	if err != nil {
		t.Fatalf("GetOneById(commissioner_proposal_vote) after delete: %v", err)
	}
	if exists {
		t.Fatal("commissioner_proposal_vote should have been deleted")
	}
}
