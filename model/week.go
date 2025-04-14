package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"sort"
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
	ID      WeekId
	DraftId DraftId
	Date    time.Time
	Note    string
}

func (w *Week) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

func (w *Week) PreDelete(db common.DatabaseProvider) error {
	draft, exists, err := common.GetOneById(db, &Draft{}, w.DraftId.RecordId())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("cannot delete week assigned to existing draft %s exists", w.DraftId)
	} else {
		err = w.DeleteAssignedAvailabilities(db)
		if err != nil {
			return err
		}

	}
	if draft != nil {
		return errors.New("draft is already deleted")
	}
	return nil
}

func (w *Week) DeleteAssignedAvailabilities(db common.DatabaseProvider) error {
	// get all availability records assigned to this Week
	availabilities, err := common.GetAllWhere(db, &Availability{}, func(c *Availability) bool {
		return c.WeekId == w.ID
	})
	if err != nil {
		return err
	}

	// delete each assigned availability
	for _, a := range availabilities {
		_, _, err = common.DeleteOneById(db, &Availability{}, a.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Week) PreUpdate(db common.DatabaseProvider, existingValues common.CrudRecord) error {
	weekInDatabase := existingValues.(*Week)
	if w.DraftId != weekInDatabase.DraftId {
		return fmt.Errorf("Draft ID %s cannot be changed\n", w.DraftId)
	}
	return nil
}

func (w *Week) SetOwner(recordId common.RecordId) {
	// don't need to do anything as Week records have
	// ownership automatically inferred / enforced by the
	// values of the SeasonId field
}

func NewWeek() *Week {
	return &Week{}
}

func (w *Week) EditableBy(db common.DatabaseProvider) []common.RecordId {

	draft, err := common.GetExistingRecordById(db, &Draft{}, w.DraftId.RecordId())
	if err != nil {
		fmt.Printf("Error getting draft in week.EditableBy(): %s", err)
		return nil
	}

	season, err := draft.GetSeason(db)
	if err != nil {
		fmt.Printf("Error getting season in week.EditableBy(): %s", err)
		return nil
	}

	// if Season is set up, then this Week is editable by any of the Season
	// commissioners. Otherwise, it is only editable by the Draft owner
	if season != nil {
		EditableBySeason(db, season.ID)
	}
	return []common.RecordId{draft.Owner.RecordId()}
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

	// draft ID must be set for the week
	err := common.ExistsById(db, &Draft{}, w.DraftId.RecordId())
	if err != nil {
		return err
	}

	return nil
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

func (w *Week) GetNextWeek(allWeeks []*Week) (*Week, error) {
	for i, week := range allWeeks {
		if week.ID == w.ID {
			if i < len(allWeeks)-1 {
				return allWeeks[i+1], nil
			} else {
				return nil, nil
			}
		}
	}
	return nil, fmt.Errorf("week %s was not found in provided weeks list", w.ID)
}

// GetWeeksForDraft gets all the Week records associated with a Draft,
// sorted in ascending order by Week.Date
func GetWeeksForDraft(db common.DatabaseProvider, id DraftId) ([]*Week, error) {
	// get all weeks with matching draft ID
	allWeeks, err := common.GetAllWhere(db, &Week{}, func(c *Week) bool {
		return c.DraftId == id
	})
	if err != nil {
		return nil, err
	}

	// sort the weeks by date and return
	sort.Slice(allWeeks, func(i, j int) bool {
		return allWeeks[i].Date.Before(allWeeks[j].Date)
	})
	return allWeeks, nil
}

func (w *Week) PushBackDefault(db common.DatabaseProvider) error {
	allWeeks, err := GetWeeksForDraft(db, w.DraftId)
	if err != nil {
		return err
	}

	weekToPush := w
	var nextWeek *Week
	for weekToPush != nil {

		// get the week after the current week
		nextWeek, err = w.GetNextWeek(allWeeks)
		if err != nil {
			return err
		}

		var newDate time.Time
		if nextWeek != nil {
			// if this is not the last week in the list,
			// then we will push this week's playing date
			// to the value from the subsequent week
			newDate = nextWeek.Date
		} else {
			// otherwise, we will just push the date by
			// one week by default and the admin can
			// change the value if necessary.
			newDate = w.Date.AddDate(0, 0, 7)
		}

		// update this week's date with the new week
		weekToPush.Date = newDate
		err = common.UpdateOne(db, weekToPush)
		if err != nil {
			return err
		}

		if nextWeek != nil {
			// move next week's availabilities to this week as we have
			// changed this week's playing date to next week's value
			err = w.MoveAvailabilities(db, nextWeek)
			if err != nil {
				return err
			}
		} else {
			// if we are on the last week, we can just delete all of
			// this week's availabilities as the date has changed to
			// a week to which no one could have added availability yet
			return w.DeleteAssignedAvailabilities(db)
		}

		weekToPush = nextWeek
	}
	return nil
}

func (w *Week) MoveAvailabilities(db common.DatabaseProvider, pushedTo *Week) error {
	// delete all the availabilities for this week
	err := w.DeleteAssignedAvailabilities(db)
	if err != nil {
		return err
	}

	nextWeekAvailabilities, err := common.GetAllWhere(db, &Availability{}, func(c *Availability) bool {
		return c.WeekId == pushedTo.ID
	})
	if err != nil {
		return err
	}

	for _, a := range nextWeekAvailabilities {
		a.WeekId = w.ID
		err = common.UpdateOne(db, a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Week) PushBackTo(newDate time.Time) {
	// push back this week

	// push back all subsequent weeks to the date of the next week

	// push back the last week to 1 week after its original date by default

	// change all availabilities corresponding to this week to the availability
	// of the week that this was pushed back into

}
