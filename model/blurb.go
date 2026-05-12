package model

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"intraclub/database"
)

type BlurbId database.RecordId

func (id BlurbId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id BlurbId) String() string {
	return id.RecordId().String()
}

type Blurb struct {
	ID        BlurbId
	Title     string
	Content   string
	Photos    []PhotoId
	Owner     database.UserId
	Season    SeasonId
	Reactions ReactionList
}

func (b *Blurb) GetOwner() database.UserId {
	return b.Owner
}

func (b *Blurb) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{
		b.Owner,
	}
}

func (b *Blurb) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (b *Blurb) SetOwner(userId database.UserId) {
	b.Owner = userId
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

func (b *Blurb) DynamicallyValid(ctx context.Context, db database.Provider) error {

	err := database.ExistsById(ctx, db, &User{}, b.Owner.RecordId())
	if err != nil {
		return err
	}

	err = database.ExistsById(ctx, db, &Season{}, b.Season.RecordId())
	if err != nil {
		return err
	}

	for _, id := range b.Photos {
		v, exists, err := database.GetOneById(ctx, db, &Photo{}, id.RecordId())
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
	return b.Reactions.DynamicallyValid(ctx, db)
}

func (b *Blurb) Type() string {
	return "blurb"
}

func (b *Blurb) GetId() database.RecordId {
	return b.ID.RecordId()
}

func (b *Blurb) SetId(id database.RecordId) {
	b.ID = BlurbId(id)
}

func (b *Blurb) React(ctx context.Context, db database.Provider, u database.UserId, t reactionType) error {
	r := &Reaction{
		UserId: u,
		Type:   t,
	}

	err := b.Reactions.CanAddReaction(ctx, db, r)
	if err != nil {
		return err
	}

	err = b.CanUserCommentOrReact(ctx, db, u)
	if err != nil {
		return err
	}

	b.Reactions = append(b.Reactions, r)

	return database.UpdateOne(ctx, db, b)
}

func (b *Blurb) Unreact(ctx context.Context, db database.Provider, u database.UserId, t reactionType) error {
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
	return database.UpdateOne(ctx, db, b)
}

func (b *Blurb) CanUserCommentOrReact(ctx context.Context, db database.Provider, u database.UserId) error {
	// no error when we receive an empty user ID
	if u.RecordId() == database.InvalidRecordId {
		return nil
	}

	err := database.ExistsById(ctx, db, &User{}, u.RecordId())
	if err != nil {
		return err
	}

	season, exists, err := database.GetOneById(ctx, db, &Season{}, b.Season.RecordId())
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("season with ID %s does not exist", b.Season)
	}

	isInSeason, err := season.IsUserIdASeasonParticipant(ctx, db, u)
	if err != nil {
		return err
	}

	if !isInSeason {
		return fmt.Errorf("user '%s' is not a participant in season '%s'", u, b.Season)
	}

	return nil
}

func (b *Blurb) GetComments(ctx context.Context, db database.Provider) ([]*Comment, error) {
	v, err := database.GetAllWhere[*Comment](ctx, db, func(_ context.Context, c *Comment) bool {
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

func (b *Blurb) NewRecord() database.CrudRecord {
	return new(Blurb)
}
