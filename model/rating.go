package model

import (
	"errors"
	"intraclub/common"
	"strings"
)

type RatingId common.RecordId

func (id RatingId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id RatingId) String() string {
	return id.RecordId().String()
}

type Rating struct {
	ID          RatingId
	UserId      UserId
	Name        string
	Description string
}

func (r *Rating) SetOwner(recordId common.RecordId) {
	r.UserId = UserId(recordId)
}

func NewRating() *Rating {
	return &Rating{}
}

func (r *Rating) EditableBy(common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{r.UserId.RecordId()}
}

func (r *Rating) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (r *Rating) Type() string {
	return "rating"
}

func (r *Rating) GetId() common.RecordId {
	return r.ID.RecordId()
}

func (r *Rating) SetId(id common.RecordId) {
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

func (r *Rating) DynamicallyValid(db common.DatabaseProvider) error {
	return nil
}
