package model

import (
	"fmt"
	"intraclub/common"
	"time"
)

type Week struct {
	ID           string    `json:"week_id" bson:"week_id"`
	Date         time.Time `json:"date" bson:"date"`                   // date when this week was actually played
	OriginalDate time.Time `json:"original_date" bson:"original_date"` // date when this week was originally scheduled to play (e.g. before a rain day)
}

func (w *Week) RecordType() string {
	return "week"
}

func (w *Week) OneRecord() common.CrudRecord {
	return new(Week)
}

type listOfWeeks []*Week

func (l listOfWeeks) Length() int {
	return len(l)
}

func (w *Week) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfWeeks, 0)
}

func (w *Week) SetId(id string) {
	w.ID = id
}

func (w *Week) GetId() string {
	return w.ID
}

func (w *Week) ValidateStatic() error {
	if w.Date.IsZero() {
		return fmt.Errorf("date field must not be empty")
	}

	if w.OriginalDate.IsZero() {
		return fmt.Errorf("original date field must not be empty")
	}

	return nil
}

func (w *Week) ValidateDynamic(db common.DbProvider) error {
	return nil
}

func (w *Week) PushBack(weeks int) {
	w.Date = w.Date.Add(time.Duration(weeks) * oneWeek)
}
