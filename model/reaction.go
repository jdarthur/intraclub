package model

import (
	"fmt"
	"intraclub/common"
)

type reactionType int

func (t reactionType) StaticallyValid() error {
	if t >= InvalidReactionType {
		return fmt.Errorf("invalid reaction type: %d / %s", t, t)
	}
	return nil
}

const (
	ThumbsUp reactionType = iota
	ThumbsDown
	Laughing
	Fire
	Heart
	Crying
	InvalidReactionType
)

func (t reactionType) String() string {
	switch t {
	case ThumbsUp:
		return "Thumbs up"
	case ThumbsDown:
		return "Thumbs down"
	case Laughing:
		return "Laughing"
	case Fire:
		return "Fire"
	case Heart:
		return "Heart"
	case Crying:
		return "Crying"
	default:
		return "Unknown"
	}
}

func (t reactionType) Emoji() string {
	switch t {
	case ThumbsUp:
		return "👍"
	case ThumbsDown:
		return "👎"
	case Laughing:
		return "😂"
	case Fire:
		return "🔥"
	case Heart:
		return "❤️"
	case Crying:
		return "😭"
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
	ThumbsDown,
	Laughing,
	Fire,
	Heart,
	Crying,
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

func (r *Reaction) StaticallyValid() error {
	return r.Type.StaticallyValid()
}

func (r *Reaction) DynamicallyValid(db common.DatabaseProvider) error {
	return common.ExistsById(db, &User{}, r.UserId.RecordId())
}

func (r *Reaction) Equals(other *Reaction) bool {
	return r.UserId == other.UserId && r.Type == other.Type
}

type ReactionList []*Reaction

func (r ReactionList) StaticallyValid() error {
	for i, reaction := range r {
		err := reaction.StaticallyValid()
		if err != nil {
			return err
		}

		if i != len(r)-1 {
			for j, other := range r[i+1:] {
				if other.Equals(reaction) {
					return fmt.Errorf("duplicate reactions at index %d and %d", i, j+i+1)
				}
			}
		}
	}
	return nil
}

func (r ReactionList) DynamicallyValid(db common.DatabaseProvider) error {
	for _, reaction := range r {
		err := reaction.DynamicallyValid(db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r ReactionList) CanAddReaction(db common.DatabaseProvider, new *Reaction) error {
	err := common.Validate(db, new)
	if err != nil {
		return err
	}

	for _, existing := range r {
		if existing.Equals(new) {
			return fmt.Errorf("reaction already exists: %+v", existing)
		}
	}
	return nil
}
