package model

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"time"
)

type League struct {
	ID           primitive.ObjectID `json:"league_id" bson:"_id"`             // primary object ID for this league
	Name         string             `json:"name" bson:"name"`                 // descriptive name for this league
	Colors       []TeamColor        `json:"colors" bson:"colors"`             // Colors assigns the color scheme of the teams in this league
	Commissioner string             `json:"commissioner" bson:"commissioner"` // The main owner of this league
	Reporters    []string           `json:"reporters" bson:"reporters"`       // Reporters are eligible to create a Blurb for this league
	Facility     string             `json:"facility" bson:"facility"`         // id of the Facility where this league plays
	StartTime    time.Time          `json:"start_time" bson:"start_time"`
	Weeks        []string           `json:"weeks" bson:"weeks"`
	Active       bool               `json:"active,omitempty" bson:"-"`
}

func (l *League) VerifyUpdatable(c common.CrudRecord) (illegalUpdate bool, field string) {
	existingLeague := c.(*League)

	if l.Commissioner != existingLeague.Commissioner {
		return true, "commissioner"
	}

	return false, ""
}

func (l *League) GetUserId() string {
	return l.Commissioner
}

func (l *League) RecordType() string {
	return "league"
}

func (l *League) OneRecord() common.CrudRecord {
	return new(League)
}

type listOfLeagues []*League

func (l listOfLeagues) Get(index int) common.CrudRecord {
	return l[index]
}

func (l listOfLeagues) Length() int {
	return len(l)
}

func (l *League) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfLeagues, 0)
}

func (l *League) SetId(id primitive.ObjectID) {
	l.ID = id
}

func (l *League) GetId() primitive.ObjectID {
	return l.ID
}

func (l *League) ValidateStatic() error {

	for _, color := range l.Colors {
		err := color.ValidateStatic()
		if err != nil {
			return fmt.Errorf("invalid team color %+v: %s", color, err.Error())
		}
	}

	if len(l.Colors) > 1 {
		return l.checkDuplicateColors()
	}

	year, month, day := l.StartTime.Date()

	if year != 1 {
		return errors.New("year must not be set in start time")
	}

	if month != 1 {
		return errors.New("month must not be set in start time")
	}

	if day != 1 {
		return errors.New("day must not be set in start time")
	}

	return nil
}

func (l *League) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {

	err := common.CheckExistenceOrErrorByStringId(common.GlobalDbProvider, &User{}, l.Commissioner)
	if err != nil {
		return err
	}

	for i, userId := range l.Reporters {
		err := common.CheckExistenceOrErrorByStringId(common.GlobalDbProvider, &User{}, userId)
		if err != nil {
			return fmt.Errorf("error with reporter at index %d: %s", i, err)
		}
	}

	err = common.CheckExistenceOrErrorByStringId(common.GlobalDbProvider, &Facility{}, l.Facility)
	if err != nil {
		return err
	}

	for _, w := range l.Weeks {
		err := common.CheckExistenceOrErrorByStringId(db, &Week{}, w)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckDuplicateColors validates that each TeamColor in the League.Colors list
// has a unique color name and hex code
func (l *League) checkDuplicateColors() error {
	for i, color := range l.Colors[:len(l.Colors)-1] {
		for j, color2 := range l.Colors[i+1:] {
			if color.Name == color2.Name {
				return fmt.Errorf("duplicate color name at index %d / %d", i, i+j+1)
			} else if color.Hex == color2.Hex {
				return fmt.Errorf("duplicate color hex code at index %d / %d", i, i+j+1)
			}
		}
	}

	return nil
}

func (l *League) GetWeeks(provider common.DbProvider) ([]*Week, error) {

	weeks := make([]*Week, 0)
	for _, w := range l.Weeks {

		weekId, err := primitive.ObjectIDFromHex(w)
		if err != nil {
			return nil, err
		}

		search := &Week{ID: weekId}
		week, exists, err := provider.GetOne(search)
		if !exists {
			return nil, common.RecordDoesNotExist(&Week{})
		}
		if err != nil {
			return nil, err
		}
		weeks = append(weeks, week.(*Week))
	}

	return weeks, nil
}

var oneWeek = time.Hour * 24 * 7

var TimeFormat = "2006-01-02"

// RainDelayOn takes a Season and pushes all of the weeks back to the next
// week's date. This is usually 7 days later, but when there is a holiday in
// between weeks, this logic will correctly move the week around the holiday
// by just switching the week's Date to the next Week in the list (e.g. 14 days later)
func (l *League) RainDelayOn(provider common.DbProvider, weekId string) error {

	startWeek := -1
	weeks, err := l.GetWeeks(provider)
	if err != nil {
		return err
	}

	weekObjId, err := primitive.ObjectIDFromHex(weekId)
	if err != nil {
		return err
	}

	// check that this week ID is actually present in the Season
	for i, week := range weeks {
		if weekObjId == week.ID {
			startWeek = i
			break
		}
	}

	// if this week ID was not found, then we don't need to continue
	if startWeek == -1 {
		return fmt.Errorf("week with ID %s was not found in league %s", weekId, l.ID)
	}

	weeksAffected := weeks[startWeek:]

	for i, week := range weeksAffected {

		// if this is the last week, we will push back a manual time period
		if i == len(weeksAffected)-1 {

			// check if the next week is a holiday maybe

			week.PushBack(1)

			// if this is the last week of the season, we will not have a way
			// to auto-update availability records for the next week. In this case,
			// we will just have to delete all of them and rely on the players to
			// update their availability for the next week in the availability page.

			err = DeleteAllAvailabilityForWeek(provider, week)
			if err != nil {
				return err
			}
		} else {

			// if this isn't the last week, just use the date of the next entry in the season

			nextWeek := weeksAffected[i+1]
			week.Date = nextWeek.Date

			// get all of the availability records for the rained out week and replace them with
			// the records for the following week

			err = UpdateAvailabilityForWeek(provider, week, nextWeek)
			if err != nil {
				return err
			}
		}

		err = provider.Update(week)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateAvailabilityForWeek(db common.DbProvider, rainedOutWeek, nextWeekInSeason *Week) error {

	rainedOutAvailability, err := GetAllAvailabilityByWeekId(db, rainedOutWeek)
	if err != nil {
		return err
	}

	nextWeekAvailability, err := GetAllAvailabilityByWeekId(db, nextWeekInSeason)
	if err != nil {
		return err
	}

	// loop through all of the availability entries that we had in the database for the
	// week that got rained out. We will auto-update the availability for each user to the
	// following week's availability when a rain-out happens so that we don't have out-of
	// date information
	for _, a := range rainedOutAvailability {

		// get next week's availability for this user, if it exists

		availabilityNextWeekForThisUser, exists := GetAvailabilityForUserId(nextWeekAvailability, a.UserId)
		if exists {

			// if it exists, we will set the availability for this week ID to the availability that we
			// had already set for this user for the week afterward.

			a.Available = availabilityNextWeekForThisUser.Available
			err = common.Update(db, a)
			if err != nil {
				return err
			}
		} else {

			// if no availability existed for the next week, we will delete this availability record as we
			// do not know if this user will be available anymore for the updated week after the rain-out

			err = common.Delete(db, a)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteAllAvailabilityForWeek finds all of the Availability records for a particular
// Week using the provided common.DbProvider and deletes the records
func DeleteAllAvailabilityForWeek(db common.DbProvider, week *Week) error {
	availability, err := GetAllAvailabilityByWeekId(db, week)
	if err != nil {
		return err
	}

	for _, a := range availability {

		err = common.Delete(db, a)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetAllAvailabilityByWeekId(db common.DbProvider, week *Week) ([]*Availability, error) {

	rainedOutWeekAvailability, err := common.GetAllWhere(db, &Availability{}, map[string]interface{}{"week_id": week.ID.Hex()})
	if err != nil {
		return nil, err
	}

	output := make([]*Availability, 0)
	for i := 0; i < rainedOutWeekAvailability.Length(); i++ {
		a := rainedOutWeekAvailability.Get(i)
		output = append(output, a.(*Availability))
	}

	return output, nil
}

func GetAvailabilityForUserId(availability []*Availability, userId string) (*Availability, bool) {
	for _, av := range availability {
		if av.UserId == userId {
			return av, true
		}
	}

	return nil, false
}

func (l *League) GetTeamsForLeague(db common.DbProvider) ([]*Team, error) {
	v, err := common.GetAllWhere(db, &Team{}, map[string]interface{}{"league_id": l.ID})
	if err != nil {
		return nil, err
	}

	return v.(listOfTeams), nil
}

// IsActive returns true if it the current date is at least 1 day
// past the last Week in this League.
func (l *League) IsActive(db common.DbProvider) (bool, error) {
	weeks, err := l.GetWeeks(db)
	if err != nil {
		return false, err
	}

	lastWeekDate := time.Time{}
	for _, week := range weeks {
		if week.Date.After(lastWeekDate) {
			lastWeekDate = week.Date
		}
	}

	return time.Now().After(lastWeekDate.Add(time.Hour * 24)), nil
}
