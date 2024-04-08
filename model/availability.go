package model

import (
	"fmt"
	"intraclub/common"
)

type AvailabilityOption int

const (
	Unavailable = iota
	MaybeAvailable
	Available
)

func (o AvailabilityOption) String() string {
	return [...]string{
		"Unavailable",
		"Maybe",
		"Available",
	}[o]
}

type listOfAvailability []*Availability

func (l listOfAvailability) Length() int {
	return len(l)
}

type Availability struct {
	Id        string             `json:"availability_id" bson:"availability_id"`
	WeekId    string             `json:"week_id" bson:"week_id"`
	UserId    string             `json:"user_id" bson:"user_id"`
	Available AvailabilityOption `json:"available" bson:"available"`
}

func (a *Availability) RecordType() string {
	return "availability"
}

func (a *Availability) OneRecord() common.CrudRecord {
	return new(Availability)
}

func (a *Availability) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfAvailability, 0)
}

func (a *Availability) SetId(id string) {
	a.Id = id
}

func (a *Availability) GetId() string {
	return a.Id
}

func (a *Availability) ValidateStatic() error {
	if a.Available == Unavailable {
		return nil
	} else if a.Available == MaybeAvailable {
		return nil
	} else if a.Available == Available {
		return nil
	}

	return fmt.Errorf("unexpected availability option: %d", a.Available)
}

func (a *Availability) ValidateDynamic(db common.DbProvider) error {

	_, exists, err := db.GetOne(&Week{ID: a.WeekId})
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("invalid week ID '%s'", a.WeekId)
	}

	_, exists, err = db.GetOne(&User{ID: a.UserId})
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("invalid user ID '%s'", a.UserId)
	}

	return nil
}
