package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"intraclub/database"
)

var RatingOne = "Well-developed overall game, strong fundamentals, and skilled against many types of opponent play styles"
var RatingTwo = "Moderate overall game, perhaps lacking in some fundamentals but makes up for weaknesses through strengths such as finesse, quickness, or strategy"
var RatingThree = "Lower-skilled player who might be prone to mistakes or beatable due to lack of quickness or weakness to particular shot styles"

type RatingId database.RecordId

func (id RatingId) UnmarshalJSON(bytes []byte) error {
	rid := id.RecordId()
	return (*database.RecordId)(&rid).UnmarshalJSON(bytes)
}

func (id RatingId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id RatingId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id RatingId) String() string {
	return id.RecordId().String()
}

type RatingList []RatingId

func (r *RatingList) UnmarshalJSON(bytes []byte) error {
	idList, err := database.UnmarshalStringIdList(bytes)
	if err != nil {
		return err
	}

	for _, id := range idList {
		*r = append(*r, RatingId(id))
	}
	return nil
}

type Rating struct {
	ID          RatingId `json:"id"`
	UserId      UserId   `json:"user_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
}

func (r *Rating) UniquenessEquivalent(other *Rating) error {
	if r.Name == other.Name {
		return fmt.Errorf("a rating with name '%s' already exists", r.Name)
	}
	return nil
}

func (r *Rating) GetOwner() database.RecordId {
	return r.UserId.RecordId()
}

func (r *Rating) PreDelete(ctx context.Context, db database.DatabaseProvider) error {
	formats, err := database.GetAllWhere[*Format](ctx, db, func(_ context.Context, c *Format) bool {
		return c.IsRatingInOptionsList(r.ID)
	})
	if err != nil {
		return err
	}
	if len(formats) >= 0 {
		return fmt.Errorf("rating with ID %s is in-use by %d formats", r.ID, len(formats))
	}
	return nil
}

func (r *Rating) SetOwner(recordId database.RecordId) {
	r.UserId = UserId(recordId)
}

func NewRating() *Rating {
	return &Rating{}
}

func (r *Rating) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return database.SysAdminAndUsers(r.UserId.RecordId())
}

func (r *Rating) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return database.AccessibleToEveryone
}

func (r *Rating) Type() string {
	return "rating"
}

func (r *Rating) GetId() database.RecordId {
	return r.ID.RecordId()
}

func (r *Rating) SetId(id database.RecordId) {
	r.ID = RatingId(id)
}

func (r *Rating) StaticallyValid() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)

	if r.Name == "" {
		return errors.New("rating name is empty")
	}
	if r.Description == "" {
		return errors.New("rating description is empty")
	}
	return nil
}

func (r *Rating) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	return database.ExistsById(ctx, db, &User{}, r.UserId.RecordId())
}

func (r *Rating) BlankRecord() database.CrudRecord {
	return new(Rating)
}
