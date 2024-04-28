package model

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"strings"
	"time"
)

type Comment struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`                // unique ID for this comment
	References string             `json:"references" bson:"references"` // ID of the Blurb that this comment is on
	UserID     string             `json:"user_id" bson:"user_id"`       // ID of the User that created this comment
	Content    string             `json:"content" bson:"content"`       // content of the comment itself
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"` // when this comment was created
	Reactions  []Reaction         `json:"reactions" bson:"reactions"`   // list of user reactions to this comment, if any
}

func (c *Comment) RecordType() string {
	return "comment"
}

func (c *Comment) OneRecord() common.CrudRecord {
	return new(Comment)
}

type listOfComments []*Comment

func (l listOfComments) Length() int {
	return len(l)
}

func (l listOfComments) Get(index int) common.CrudRecord {
	return l[index]
}

func (c *Comment) ListOfRecords() common.ListOfCrudRecords {
	return listOfComments{}
}

func (c *Comment) SetId(id primitive.ObjectID) {
	c.ID = id
}

func (c *Comment) GetId() primitive.ObjectID {
	return c.ID
}

func (c *Comment) ValidateStatic() error {

	c.Content = strings.Trim(c.Content, " \t\r\n")

	if c.Content == "" {
		return errors.New("content must not be empty")
	}

	for _, reaction := range c.Reactions {
		err := reaction.ValidateStatic()
		if err != nil {
			return fmt.Errorf("error in reaction %+v: %s", reaction, err.Error())
		}
	}

	return nil
}

func (c *Comment) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {

	err := common.CheckExistenceOrErrorByStringId(db, &User{}, c.UserID)
	if err != nil {
		return err
	}

	err = common.CheckExistenceOrErrorByStringId(db, &Blurb{}, c.References)
	if err != nil {
		return err
	}

	for _, reaction := range c.Reactions {
		err := reaction.ValidateDynamic(db, isUpdate, nil)
		if err != nil {
			return fmt.Errorf("error in reaction %+v: %s", reaction, err.Error())
		}
	}

	return nil
}

func (c *Comment) AddReaction(db common.DbProvider, reaction Reaction) error {

	err := common.ValidateStaticAndDynamic(db, reaction)
	if err != nil {
		return err
	}

	c.Reactions = append(c.Reactions, reaction)
	return nil
}

type reactionType int

const (
	ThumbsUp reactionType = iota
	Laughing
	Fire
	Heart
)

func (t reactionType) String() string {
	switch t {
	case ThumbsUp:
		return "ThumbsUp"
	case Laughing:
		return "Laughing"
	case Fire:
		return "Fire"
	case Heart:
		return "Heart"
	default:
		return "Unknown"
	}
}

func GetAllReactionTypes() map[string]reactionType {
	m := make(map[string]reactionType)
	for _, r := range AllowedReactions {
		m[r.String()] = r
	}
	return m
}

var AllowedReactions = []reactionType{
	ThumbsUp,
	Laughing,
	Fire,
	Heart,
}

// Reaction is an object representing a user hitting an emoji reaction
// on a Blurb or Comment. The UserId must be a valid User.ID and the
// Type must be in the AllowedReactions list. These are not stored
// directly in the database but are instead stored on a []Reaction in the
// Blurb.Reactions and Comment.Reactions fields.
type Reaction struct {
	UserId string       `json:"user_id" bson:"user_id"`
	Type   reactionType `json:"reaction_type" bson:"reaction_type"`
}

func (r Reaction) ValidateStatic() error {
	for _, reaction := range AllowedReactions {
		if r.Type == reaction {
			return nil
		}
	}

	return fmt.Errorf("invalid reaction type: %d", r.Type)
}

func (r Reaction) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {
	return common.CheckExistenceOrErrorByStringId(db, &User{}, r.UserId)
}
