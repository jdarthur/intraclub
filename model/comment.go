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
	Reactions  ReactionList       `json:"reactions" bson:"reactions"`   // list of user reactions to this comment, if any
}
