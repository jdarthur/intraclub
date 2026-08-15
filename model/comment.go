package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"intraclub/database"
)

type CommentId database.RecordId

func (id CommentId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id CommentId) String() string {
	return id.RecordId().String()
}

func (id CommentId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id *CommentId) UnmarshalJSON(data []byte) error {
	rid := id.RecordId()
	if err := (*database.RecordId)(&rid).UnmarshalJSON(data); err != nil {
		return err
	}
	*id = CommentId(rid)
	return nil
}

type Comment struct {
	ID        CommentId       `json:"id"`                           // unique ID for this comment
	Blurb     BlurbId         `json:"blurb"`                        // ID of the Blurb that this comment is on
	ReplyTo   CommentId       `json:"reply_to"`                     // ID of the Comment that this is in reference to (if any)
	Owner     database.UserId `json:"user_id" bson:"user_id"`       // ID of the User that created this comment
	Content   string          `json:"content" bson:"content"`       // content of the comment itself
	EditedAt  time.Time       `json:"edited_at" bson:"edited_at"`   // time that this Comment was edited (if applicable)
	CreatedAt time.Time       `json:"created_at" bson:"created_at"` // when this comment was created
}

func (c *Comment) GetOwner() database.UserId {
	return c.Owner
}

func (c *Comment) CanOnlyDelete(ctx context.Context, db database.Provider, userId database.UserId) bool {
	return userId != c.Owner
}

func (c *Comment) GetTimeStamps() (created, updated time.Time) {
	return c.CreatedAt, c.EditedAt
}

func (c *Comment) SetCreateTimestamp(t time.Time) time.Time {
	oldValue := c.CreatedAt
	c.CreatedAt = t
	return oldValue
}

func (c *Comment) SetUpdateTimestamp(t time.Time) time.Time {
	oldValue := c.EditedAt
	c.EditedAt = t
	return oldValue
}

func NewComment() *Comment {
	return &Comment{}
}

func (c *Comment) Type() string {
	return "comment"
}

func (c *Comment) GetId() database.RecordId {
	return c.ID.RecordId()
}

func (c *Comment) SetId(id database.RecordId) {
	c.ID = CommentId(id)
}

func (c *Comment) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	blurb, err := database.GetExistingRecordById(ctx, db, &Blurb{}, c.Blurb.RecordId())
	if err != nil {
		return []database.UserId{}
	}
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, blurb.Season.RecordId())
	if err != nil {
		return []database.UserId{}
	}

	editors := []database.UserId{
		database.SysAdminUserId,
		c.Owner,
		blurb.Owner,
	}

	commissioners, err := season.GetCommissioners(ctx, db)
	if err == nil {
		for _, commissioner := range commissioners {
			editors = append(editors, commissioner)
		}
	}
	return editors
}

func (c *Comment) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (c *Comment) SetOwner(userId database.UserId) {
	c.Owner = userId
}

func (c *Comment) StaticallyValid() error {
	c.Content = strings.TrimSpace(c.Content)
	if c.Content == "" {
		return fmt.Errorf("comment is empty")
	}

	// The self-reference and created-timestamp checks only apply once the
	// comment has been persisted (an ID has been assigned). The generic CRUD
	// route wrapper pre-validates a fresh request body before
	// CreateOne/UpdateOne assigns the ID and timestamps, at which point both
	// ReplyTo and ID are zero; gating on a set ID keeps that pre-check from
	// spuriously rejecting a brand-new comment.
	if c.ID != CommentId(database.InvalidRecordId) {
		if c.ReplyTo == c.ID {
			return errors.New("comment references itself")
		}
		if c.CreatedAt.IsZero() {
			return fmt.Errorf("comment created timestamp is zero")
		}
	}

	return nil
}

func (c *Comment) DynamicallyValid(ctx context.Context, db database.Provider) error {
	err := database.ExistsById(ctx, db, &User{}, c.Owner.RecordId())
	if err != nil {
		return err
	}
	blurb, err := database.GetExistingRecordById(ctx, db, &Blurb{}, c.Blurb.RecordId())
	if err != nil {
		return err
	}

	season, err := database.GetExistingRecordById(ctx, db, &Season{}, blurb.Season.RecordId())
	if err != nil {
		return err
	}

	isParticipant, err := season.IsUserIdASeasonParticipant(ctx, db, c.Owner)
	if err != nil {
		return err
	}
	if !isParticipant {
		return fmt.Errorf("user %s is not a participant in season %s", c.Owner, season.ID)
	}

	if c.ReplyTo != CommentId(database.InvalidRecordId) {
		v, err := database.GetExistingRecordById(ctx, db, &Comment{}, c.ReplyTo.RecordId())
		if err != nil {
			return err
		}

		if v.Blurb != c.Blurb {
			return errors.New("referenced comment references a different blurb")
		}
	}

	// validate each reaction row (via the comment_reaction child table)
	reactions, err := c.getReactionRows(ctx, db)
	if err != nil {
		return err
	}
	for _, row := range reactions {
		r := &Reaction{UserId: row.UserId, Type: row.ReactionType}
		if err := r.DynamicallyValid(ctx, db); err != nil {
			return err
		}
	}
	return nil
}

// PostDelete cascades deletion to this comment's comment_reaction child rows.
// Without this, deleting a comment would orphan those rows (see #97).
func (c *Comment) PostDelete(ctx context.Context, db database.Provider) error {
	reactions, err := database.GetAllWhere[*CommentReaction](ctx, db, func(_ context.Context, r *CommentReaction) bool {
		return r.CommentId == c.ID
	})
	if err != nil {
		return err
	}
	for _, r := range reactions {
		if _, _, err := database.DeleteOneById(ctx, db, r, r.ID); err != nil {
			return err
		}
	}
	return nil
}

func (c *Comment) NewRecord() database.CrudRecord {
	return new(Comment)
}

// CommentReaction is a child-table record that assigns a single Reaction from
// a User to a Comment. This is the normalized replacement for the former inline
// `Comment.Reactions` slice, enabling the relationship to be queried/indexed
// individually and stored in its own table rather than as an embedded list.
type CommentReaction struct {
	ID           database.RecordId `json:"id"`
	CommentId    CommentId         `json:"comment_id"`
	UserId       database.UserId   `json:"user_id"`
	ReactionType reactionType      `json:"reaction_type"`
}

func (r *CommentReaction) GetOwner() database.UserId {
	return r.UserId
}

func (r *CommentReaction) SetOwner(userId database.UserId) {
	r.UserId = userId
}

func (r *CommentReaction) Type() string {
	return "comment_reaction"
}

func (r *CommentReaction) GetId() database.RecordId {
	return r.ID
}

func (r *CommentReaction) SetId(id database.RecordId) {
	r.ID = id
}

func (r *CommentReaction) StaticallyValid() error {
	return r.ReactionType.StaticallyValid()
}

func (r *CommentReaction) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Comment{}, r.CommentId.RecordId()); err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &User{}, r.UserId.RecordId())
}

func (r *CommentReaction) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (r *CommentReaction) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{r.UserId}
}

func (r *CommentReaction) NewRecord() database.CrudRecord {
	return new(CommentReaction)
}

// getReactionRows returns all CommentReaction relationship rows assigned to
// this comment.
func (c *Comment) getReactionRows(ctx context.Context, db database.Provider) ([]*CommentReaction, error) {
	filter := func(_ context.Context, r *CommentReaction) bool {
		return r.CommentId == c.ID
	}
	return database.GetAllWhere[*CommentReaction](ctx, db, filter)
}

// GetReactions reassembles the comment_reaction child rows into the former
// inline Comment.Reactions slice shape.
func (c *Comment) GetReactions(ctx context.Context, db database.Provider) (ReactionList, error) {
	rows, err := c.getReactionRows(ctx, db)
	if err != nil {
		return nil, err
	}
	reactions := make(ReactionList, 0, len(rows))
	for _, row := range rows {
		reactions = append(reactions, &Reaction{UserId: row.UserId, Type: row.ReactionType})
	}
	return reactions, nil
}

// React adds a Reaction to this comment, storing it in the comment_reaction
// child table. A duplicate (comment, user, type) reaction is rejected.
func (c *Comment) React(ctx context.Context, db database.Provider, u database.UserId, t reactionType) error {
	r := &Reaction{UserId: u, Type: t}

	if err := database.Validate(ctx, db, r); err != nil {
		return err
	}

	rows, err := c.getReactionRows(ctx, db)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.UserId == u && row.ReactionType == t {
			return fmt.Errorf("reaction already exists: %+v", r)
		}
	}

	row := &CommentReaction{
		CommentId:    c.ID,
		UserId:       u,
		ReactionType: t,
	}
	_, err = database.CreateOne(ctx, db, row)
	return err
}

// Unreact removes the Reaction matching (user, type) from this comment.
func (c *Comment) Unreact(ctx context.Context, db database.Provider, u database.UserId, t reactionType) error {
	r := &Reaction{UserId: u, Type: t}

	rows, err := c.getReactionRows(ctx, db)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.UserId == u && row.ReactionType == t {
			_, _, err := database.DeleteOneById(ctx, db, row, row.GetId())
			return err
		}
	}
	return fmt.Errorf("reaction with values %+v does not exist", r)
}
