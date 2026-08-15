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

func (id BlurbId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id *BlurbId) UnmarshalJSON(data []byte) error {
	rid := id.RecordId()
	if err := (*database.RecordId)(&rid).UnmarshalJSON(data); err != nil {
		return err
	}
	*id = BlurbId(rid)
	return nil
}

type Blurb struct {
	ID      BlurbId         `json:"id"`
	Title   string          `json:"title"`
	Content string          `json:"content"`
	Owner   database.UserId `json:"owner"`
	Season  SeasonId        `json:"season"`
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

	season, err := database.GetExistingRecordById(ctx, db, &Season{}, b.Season.RecordId())
	if err != nil {
		return err
	}

	// The blurb's owner must be a participant of the season (mirrors the
	// participant check enforced for comments and reactions).
	isParticipant, err := season.IsUserIdASeasonParticipant(ctx, db, b.Owner)
	if err != nil {
		return err
	}
	if !isParticipant {
		return fmt.Errorf("user %s is not a participant in season %s", b.Owner, b.Season)
	}

	// each attached photo (via the blurb_photo child table) must exist and be
	// owned by the blurb's owner
	photos, err := b.getPhotoRows(ctx, db)
	if err != nil {
		return err
	}
	for _, row := range photos {
		v, exists, err := database.GetOneById(ctx, db, &Photo{}, row.PhotoId.RecordId())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("photo with ID '%s' does not exist", row.PhotoId)
		}
		if v.Owner != b.Owner {
			return fmt.Errorf("photo with ID '%s' is not owned by user '%s'", row.PhotoId, b.Owner)
		}
	}

	// validate each reaction row (via the blurb_reaction child table)
	reactions, err := b.getReactionRows(ctx, db)
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

	// validate the reaction itself (static + dynamic)
	err := database.Validate(ctx, db, r)
	if err != nil {
		return err
	}

	err = b.CanUserCommentOrReact(ctx, db, u)
	if err != nil {
		return err
	}

	// one reaction per (blurb, user, type); reject duplicates
	rows, err := b.getReactionRows(ctx, db)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.UserId == u && row.ReactionType == t {
			return fmt.Errorf("reaction already exists: %+v", r)
		}
	}

	row := &BlurbReaction{
		BlurbId:      b.ID,
		UserId:       u,
		ReactionType: t,
	}
	_, err = database.CreateOne(ctx, db, row)
	return err
}

func (b *Blurb) Unreact(ctx context.Context, db database.Provider, u database.UserId, t reactionType) error {
	r := &Reaction{
		UserId: u,
		Type:   t,
	}

	rows, err := b.getReactionRows(ctx, db)
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

// PostDelete cascades deletion to this blurb's blurb_photo and blurb_reaction
// child rows. Without this, deleting a blurb would orphan those rows (see #97).
func (b *Blurb) PostDelete(ctx context.Context, db database.Provider) error {
	photos, err := database.GetAllWhere[*BlurbPhoto](ctx, db, func(_ context.Context, p *BlurbPhoto) bool {
		return p.BlurbId == b.ID
	})
	if err != nil {
		return err
	}
	for _, p := range photos {
		if _, _, err := database.DeleteOneById(ctx, db, p, p.ID); err != nil {
			return err
		}
	}

	reactions, err := database.GetAllWhere[*BlurbReaction](ctx, db, func(_ context.Context, r *BlurbReaction) bool {
		return r.BlurbId == b.ID
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

func (b *Blurb) NewRecord() database.CrudRecord {
	return new(Blurb)
}

// BlurbPhoto is a child-table record that attaches a Photo to a Blurb. This is
// the normalized replacement for the former inline `Blurb.Photos` slice,
// enabling the relationship to be queried/indexed individually and stored in
// its own table rather than as an embedded list.
type BlurbPhoto struct {
	ID      database.RecordId `json:"id"`
	BlurbId BlurbId           `json:"blurb_id"`
	PhotoId PhotoId           `json:"photo_id"`
}

func (p *BlurbPhoto) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (p *BlurbPhoto) SetOwner(userId database.UserId) {}

func (p *BlurbPhoto) Type() string {
	return "blurb_photo"
}

func (p *BlurbPhoto) GetId() database.RecordId {
	return p.ID
}

func (p *BlurbPhoto) SetId(id database.RecordId) {
	p.ID = id
}

func (p *BlurbPhoto) StaticallyValid() error {
	return nil
}

func (p *BlurbPhoto) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Blurb{}, p.BlurbId.RecordId()); err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &Photo{}, p.PhotoId.RecordId())
}

func (p *BlurbPhoto) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (p *BlurbPhoto) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (p *BlurbPhoto) NewRecord() database.CrudRecord {
	return new(BlurbPhoto)
}

// BlurbReaction is a child-table record that assigns a single Reaction from a
// User to a Blurb. This is the normalized replacement for the former inline
// `Blurb.Reactions` slice, enabling the relationship to be queried/indexed
// individually and stored in its own table rather than as an embedded list.
type BlurbReaction struct {
	ID           database.RecordId `json:"id"`
	BlurbId      BlurbId           `json:"blurb_id"`
	UserId       database.UserId   `json:"user_id"`
	ReactionType reactionType      `json:"reaction_type"`
}

func (r *BlurbReaction) GetOwner() database.UserId {
	return r.UserId
}

func (r *BlurbReaction) SetOwner(userId database.UserId) {
	r.UserId = userId
}

func (r *BlurbReaction) Type() string {
	return "blurb_reaction"
}

func (r *BlurbReaction) GetId() database.RecordId {
	return r.ID
}

func (r *BlurbReaction) SetId(id database.RecordId) {
	r.ID = id
}

func (r *BlurbReaction) StaticallyValid() error {
	return r.ReactionType.StaticallyValid()
}

func (r *BlurbReaction) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &User{}, r.UserId.RecordId()); err != nil {
		return err
	}
	blurb, err := database.GetExistingRecordById(ctx, db, &Blurb{}, r.BlurbId.RecordId())
	if err != nil {
		return err
	}
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, blurb.Season.RecordId())
	if err != nil {
		return err
	}

	// The reacting user must be a participant of the blurb's season. This
	// mirrors the participant check that the custom /react routes enforce, so
	// the generic CRUD surface can't be used to bypass it.
	isParticipant, err := season.IsUserIdASeasonParticipant(ctx, db, r.UserId)
	if err != nil {
		return err
	}
	if !isParticipant {
		return fmt.Errorf("user %s is not a participant in season %s", r.UserId, season.ID)
	}
	return nil
}

func (r *BlurbReaction) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (r *BlurbReaction) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{r.UserId}
}

func (r *BlurbReaction) NewRecord() database.CrudRecord {
	return new(BlurbReaction)
}

// getPhotoRows returns all BlurbPhoto relationship rows assigned to this blurb.
func (b *Blurb) getPhotoRows(ctx context.Context, db database.Provider) ([]*BlurbPhoto, error) {
	filter := func(_ context.Context, p *BlurbPhoto) bool {
		return p.BlurbId == b.ID
	}
	return database.GetAllWhere[*BlurbPhoto](ctx, db, filter)
}

// GetPhotos reassembles the blurb_photo child rows into the former inline
// Blurb.Photos slice shape.
func (b *Blurb) GetPhotos(ctx context.Context, db database.Provider) ([]PhotoId, error) {
	rows, err := b.getPhotoRows(ctx, db)
	if err != nil {
		return nil, err
	}
	ids := make([]PhotoId, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.PhotoId)
	}
	return ids, nil
}

// getReactionRows returns all BlurbReaction relationship rows assigned to this
// blurb.
func (b *Blurb) getReactionRows(ctx context.Context, db database.Provider) ([]*BlurbReaction, error) {
	filter := func(_ context.Context, r *BlurbReaction) bool {
		return r.BlurbId == b.ID
	}
	return database.GetAllWhere[*BlurbReaction](ctx, db, filter)
}

// GetReactions reassembles the blurb_reaction child rows into the former
// inline Blurb.Reactions slice shape.
func (b *Blurb) GetReactions(ctx context.Context, db database.Provider) (ReactionList, error) {
	rows, err := b.getReactionRows(ctx, db)
	if err != nil {
		return nil, err
	}
	reactions := make(ReactionList, 0, len(rows))
	for _, row := range rows {
		reactions = append(reactions, &Reaction{UserId: row.UserId, Type: row.ReactionType})
	}
	return reactions, nil
}
