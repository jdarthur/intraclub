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

type Comment struct {
	ID        CommentId    `json:"id"`                           // unique ID for this comment
	Blurb     BlurbId      `json:"references"`                   // ID of the Blurb that this comment is on
	ReplyTo   CommentId    `json:"references_comment"`           // ID of the Comment that this is in reference to (if any)
	Owner     UserId       `json:"user_id" bson:"user_id"`       // ID of the User that created this comment
	Content   string       `json:"content" bson:"content"`       // content of the comment itself
	EditedAt  time.Time    `json:"edited_at" bson:"edited_at"`   // time that this Comment was edited (if applicable)
	CreatedAt time.Time    `json:"created_at" bson:"created_at"` // when this comment was created
	Reactions ReactionList `json:"reactions" bson:"reactions"`   // list of user reactions to this comment, if any
}

func (c *Comment) GetOwner() database.RecordId {
	return c.Owner.RecordId()
}

func (c *Comment) CanOnlyDelete(ctx context.Context, db database.DatabaseProvider, userId database.RecordId) bool {
	return UserId(userId) != c.Owner
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

func (c *Comment) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	blurb, err := database.GetExistingRecordById(ctx, db, &Blurb{}, c.Blurb.RecordId())
	if err != nil {
		return []database.RecordId{}
	}
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, blurb.Season.RecordId())
	if err != nil {
		return []database.RecordId{}
	}

	editors := []database.RecordId{
		database.SysAdminRecordId,
		c.Owner.RecordId(),
		blurb.Owner.RecordId(),
	}

	commissioners, err := season.GetCommissioners(ctx, db)
	if err == nil {
		for _, commissioner := range commissioners {
			editors = append(editors, commissioner.RecordId())
		}
	}
	return editors
}

func (c *Comment) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return database.AccessibleToEveryone
}

func (c *Comment) SetOwner(recordId database.RecordId) {
	c.Owner = UserId(recordId)
}

func (c *Comment) StaticallyValid() error {
	c.Content = strings.TrimSpace(c.Content)
	if c.Content == "" {
		return fmt.Errorf("comment is empty")
	}

	if c.ReplyTo == c.ID {
		return errors.New("comment references itself")
	}

	if c.CreatedAt.IsZero() {
		return fmt.Errorf("comment created timestamp is zero")
	}

	return c.Reactions.StaticallyValid()
}

func (c *Comment) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
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

	return c.Reactions.DynamicallyValid(ctx, db)
}

func (c *Comment) BlankRecord() database.CrudRecord {
	return new(Comment)
}
