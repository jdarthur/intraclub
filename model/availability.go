package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (l listOfAvailability) Get(index int) common.CrudRecord {
	return l[index]
}

func (l listOfAvailability) Length() int {
	return len(l)
}

type Availability struct {
	Id        primitive.ObjectID `json:"availability_id" bson:"_id"`
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

func (a *Availability) SetId(id primitive.ObjectID) {
	a.Id = id
}

func (a *Availability) GetId() primitive.ObjectID {
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

func (a *Availability) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {

	err := common.CheckExistenceOrErrorByStringId(db, &Week{}, a.WeekId)
	if err != nil {
		return err
	}

	err = common.CheckExistenceOrErrorByStringId(db, &User{}, a.UserId)
	if err != nil {
		return err
	}

	return nil
}
