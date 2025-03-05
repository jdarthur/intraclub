package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"time"
)

type WeekId common.RecordId

func (id WeekId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id WeekId) String() string {
	return id.RecordId().String()
}

type Week struct {
	ID       WeekId
	SeasonId SeasonId
	Date     time.Time
}

func (w *Week) SetOwner(recordId common.RecordId) {
	// don't need to do anything as Week records have
	// ownership automatically inferred / enforced by the
	// values of the SeasonId field
}

func (w *Week) PostCreate(db common.DatabaseProvider) error {
	season, err := GetSeason(db, w.SeasonId)
	if err != nil {
		return err
	}
	season.Weeks = append(season.Weeks)
	return common.UpdateOne(db, season)
}

func NewWeek() *Week {
	return &Week{}
}

func (w *Week) EditableBy(db common.DatabaseProvider) []common.RecordId {
	season, err := GetSeason(db, w.SeasonId)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return UserIdListToRecordIdList(season.Commissioners)
}

func (w *Week) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.EveryoneRecordId}
}

func (w *Week) StaticallyValid() error {
	if w.Date.IsZero() {
		return errors.New("date is zero")
	}
	return nil
}

func (w *Week) DynamicallyValid(db common.DatabaseProvider) error {
	return common.ExistsById(db, &Season{}, w.SeasonId.RecordId())
}

func (w *Week) Type() string {
	return "week"
}

func (w *Week) GetId() common.RecordId {
	return w.ID.RecordId()
}

func (w *Week) SetId(id common.RecordId) {
	w.ID = WeekId(id)
}
