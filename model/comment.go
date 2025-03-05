package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"strings"
	"time"
)

type CommentId common.RecordId

func (id CommentId) RecordId() common.RecordId {
	return common.RecordId(id)
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

func (c *Comment) CanOnlyDelete(db common.DatabaseProvider, userId common.RecordId) bool {
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

func (c *Comment) GetId() common.RecordId {
	return c.ID.RecordId()
}

func (c *Comment) SetId(id common.RecordId) {
	c.ID = CommentId(id)
}

func (c *Comment) EditableBy(db common.DatabaseProvider) []common.RecordId {
	blurb, err := common.GetExistingRecordById(db, &Blurb{}, c.Blurb.RecordId())
	if err != nil {
		return []common.RecordId{}
	}
	season, err := common.GetExistingRecordById(db, &Season{}, blurb.Season.RecordId())
	if err != nil {
		return []common.RecordId{}
	}

	editors := []common.RecordId{
		common.SysAdminRecordId,
		c.Owner.RecordId(),
		blurb.Owner.RecordId(),
	}
	for _, commissioner := range season.Commissioners {
		editors = append(editors, commissioner.RecordId())
	}
	return editors
}

func (c *Comment) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (c *Comment) SetOwner(recordId common.RecordId) {
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

func (c *Comment) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &User{}, c.Owner.RecordId())
	if err != nil {
		return err
	}
	blurb, err := common.GetExistingRecordById(db, &Blurb{}, c.Blurb.RecordId())
	if err != nil {
		return err
	}

	season, err := common.GetExistingRecordById(db, &Season{}, blurb.Season.RecordId())
	if err != nil {
		return err
	}

	isParticipant, err := season.IsUserIdASeasonParticipant(db, c.Owner)
	if err != nil {
		return err
	}
	if !isParticipant {
		return fmt.Errorf("user %s is not a participant in season %s", c.Owner, season.ID)
	}

	if c.ReplyTo != CommentId(common.InvalidRecordId) {
		v, err := common.GetExistingRecordById(db, &Comment{}, c.ReplyTo.RecordId())
		if err != nil {
			return err
		}

		if v.Blurb != c.Blurb {
			return errors.New("referenced comment references a different blurb")
		}
	}

	return c.Reactions.DynamicallyValid(db)
}
