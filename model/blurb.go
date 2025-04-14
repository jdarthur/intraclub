package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"sort"
	"strings"
)

type BlurbId common.RecordId

func (id BlurbId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id BlurbId) String() string {
	return id.RecordId().String()
}

type Blurb struct {
	ID        BlurbId
	Title     string
	Content   string
	Photos    []PhotoId
	Owner     UserId
	Season    SeasonId
	Reactions ReactionList
}

func (b *Blurb) GetOwner() common.RecordId {
	return b.Owner.RecordId()
}

func (b *Blurb) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{
		b.Owner.RecordId(),
	}
}

func (b *Blurb) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (b *Blurb) SetOwner(recordId common.RecordId) {
	b.Owner = UserId(recordId)
}

func NewBlurb() *Blurb {
	return &Blurb{}
}

func (b *Blurb) StaticallyValid() error {
	b.Title = strings.TrimSpace(b.Title)
	b.Content = strings.TrimSpace(b.Content)

	if b.Title == "" {
		return errors.New("title is empty")
	}
	if b.Content == "" {
		return errors.New("content is empty")
	}
	return nil
}

func (b *Blurb) DynamicallyValid(db common.DatabaseProvider) error {

	err := common.ExistsById(db, &User{}, b.Owner.RecordId())
	if err != nil {
		return err
	}

	err = common.ExistsById(db, &Season{}, b.Season.RecordId())
	if err != nil {
		return err
	}

	for _, id := range b.Photos {
		v, exists, err := common.GetOneById(db, &Photo{}, id.RecordId())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("photo with ID '%s' does not exist", id)
		}
		if v.Owner != b.Owner {
			return fmt.Errorf("photo with ID '%s' is not owned by user '%s'", id, b.Owner)
		}
	}
	return b.Reactions.DynamicallyValid(db)
}

func (b *Blurb) Type() string {
	return "blurb"
}

func (b *Blurb) GetId() common.RecordId {
	return b.ID.RecordId()
}

func (b *Blurb) SetId(id common.RecordId) {
	b.ID = BlurbId(id)
}

func (b *Blurb) React(db common.DatabaseProvider, u UserId, t reactionType) error {
	r := &Reaction{
		UserId: u,
		Type:   t,
	}

	err := b.Reactions.CanAddReaction(db, r)
	if err != nil {
		return err
	}

	err = b.CanUserCommentOrReact(db, u)
	if err != nil {
		return err
	}

	b.Reactions = append(b.Reactions, r)

	return common.UpdateOne(db, b)
}

func (b *Blurb) Unreact(db common.DatabaseProvider, u UserId, t reactionType) error {
	r := &Reaction{
		UserId: u,
		Type:   t,
	}

	found := false
	newList := make(ReactionList, 0)
	for _, reaction := range b.Reactions {
		if reaction.Equals(r) {
			found = true
		} else {
			newList = append(newList, reaction)
		}
	}
	if !found {
		return fmt.Errorf("reaction with values %+v does not exist", r)
	}

	b.Reactions = newList
	return common.UpdateOne(db, b)
}

func (b *Blurb) CanUserCommentOrReact(db common.DatabaseProvider, u UserId) error {
	// no error when we receive an empty user ID
	if u.RecordId() == common.InvalidRecordId {
		return nil
	}

	err := common.ExistsById(db, &User{}, u.RecordId())
	if err != nil {
		return err
	}

	season, exists, err := common.GetOneById(db, &Season{}, b.Season.RecordId())
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("season with ID %s does not exist", b.Season)
	}

	isInSeason, err := season.IsUserIdASeasonParticipant(db, u)
	if err != nil {
		return err
	}

	if !isInSeason {
		return fmt.Errorf("user '%s' is not a participant in season '%s'", u, b.Season)
	}

	return nil
}

func (b *Blurb) GetComments(db common.DatabaseProvider) ([]*Comment, error) {
	v, err := common.GetAllWhere(db, &Comment{}, func(c *Comment) bool {
		return c.Blurb == b.ID
	})
	if err != nil {
		return nil, err
	}

	// sort the comments by create date
	sort.Slice(v, func(i, j int) bool {
		return v[i].CreatedAt.Before(v[j].CreatedAt)
	})
	return v, nil
}
