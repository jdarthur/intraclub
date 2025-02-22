package model

import (
	"errors"
	"intraclub/common"
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
	ID      BlurbId
	Title   string
	Content string
	Photos  []PhotoId
	UserId  UserId
	Season  SeasonId
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

	err := common.ExistsById(db, &User{}, b.UserId.RecordId())
	if err != nil {
		return err
	}

	err = common.ExistsById(db, &Season{}, b.Season.RecordId())
	if err != nil {
		return err
	}

	for _, id := range b.Photos {
		if err := common.ExistsById(db, &Photo{}, id.RecordId()); err != nil {
		}
	}

	return nil
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
