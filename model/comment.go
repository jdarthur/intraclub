package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	UserId UserId       `json:"user_id" bson:"user_id"`
	Type   reactionType `json:"reaction_type" bson:"reaction_type"`
}
