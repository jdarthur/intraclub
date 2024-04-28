package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

type Facility struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Address     string             `json:"address" bson:"address"`
	Name        string             `json:"name" bson:"name"`
	Courts      int                `json:"courts" bson:"courts"`
	LayoutImage string             `json:"layout_image" bson:"layout_image"`
}

func (f *Facility) RecordType() string {
	return "facility"
}

func (f *Facility) OneRecord() common.CrudRecord {
	return new(Facility)
}

type listOfFacilities []*Facility

func (l listOfFacilities) Length() int {
	return len(l)
}

func (l listOfFacilities) Get(index int) common.CrudRecord {
	return l[index]
}

func (f *Facility) ListOfRecords() common.ListOfCrudRecords {
	return listOfFacilities{}
}

func (f *Facility) SetId(id primitive.ObjectID) {
	f.ID = id
}

func (f *Facility) GetId() primitive.ObjectID {
	return f.ID
}

func (f *Facility) ValidateStatic() error {

	if f.Name == "" {
		return errors.New("name is required")
	}

	if f.Address == "" {
		return errors.New("address is required")
	}

	if f.Courts == 0 {
		return errors.New("facility must have at least 1 court")
	}

	return nil
}

func (f *Facility) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {

	if f.LayoutImage != "" {
		err := common.CheckExistenceOrErrorByStringId(db, &Image{}, f.LayoutImage)
		if err != nil {
			return err
		}
	}

	return nil
}
