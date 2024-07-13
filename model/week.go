package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"time"
)

type Week struct {
	ID           primitive.ObjectID `json:"week_id" bson:"_id"`
	Date         YyyyMmDdDate       `json:"date" bson:"date"`                   // date when this week was actually played
	OriginalDate YyyyMmDdDate       `json:"original_date" bson:"original_date"` // date when this week was originally scheduled to play (e.g. before a rain day)
	UserId       string             `json:"-" bson:"user_id"`                   // user ID of the user that created this week (the league commissioner)
}

func (w *Week) GetUserId() string {
	return w.UserId
}

func (w *Week) SetUserId(userId string) {
	w.UserId = userId
}

func (w *Week) RecordType() string {
	return "week"
}

func (w *Week) OneRecord() common.CrudRecord {
	return new(Week)
}

type listOfWeeks []*Week

func (l listOfWeeks) Get(index int) common.CrudRecord {
	return l[index]
}

func (l listOfWeeks) Length() int {
	return len(l)
}

func (w *Week) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfWeeks, 0)
}

func (w *Week) SetId(id primitive.ObjectID) {
	w.ID = id
}

func (w *Week) GetId() primitive.ObjectID {
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

func (w *Week) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {
	return nil
}

func (w *Week) PushBack(weeks int) {
	w.Date.Time = w.Date.Time.Add(time.Duration(weeks) * oneWeek)
}

func (w *Week) OnDelete(db common.DbProvider) error {

	leagues, err := GetCommissionedLeaguesByUserId(db, w.UserId)
	if err != nil {
		return err
	}

	for _, league := range leagues {
		weekInLeague := false
		for _, weekId := range league.Weeks {
			if weekId == w.ID.Hex() {
				weekInLeague = true
				break
			}
		}

		if weekInLeague {
			newWeekIds := make([]string, 0)
			for _, weekId := range league.Weeks {
				if weekId != w.ID.Hex() {
					newWeekIds = append(newWeekIds, weekId)
				}
			}

			league.Weeks = newWeekIds
			err = common.Update(db, league)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
