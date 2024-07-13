package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

type Facility struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Address     string             `json:"address" bson:"address"`
	Name        string             `json:"name" bson:"name"`
	Courts      int                `json:"courts" bson:"courts"`
	LayoutImage string             `json:"layout_image" bson:"layout_image"`
	UserId      string             `json:"-" bson:"user_id"`
}

func (f *Facility) SetUserId(userId string) {
	f.UserId = userId
}

func (f *Facility) GetUserId() string {
	return f.UserId
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
		return common.ApiError{
			References: []any{"Name"},
			Code:       common.FieldIsRequired,
		}
	}

	if f.Address == "" {
		return common.ApiError{
			References: []any{"Address"},
			Code:       common.FieldIsRequired,
		}
	}

	if f.Courts == 0 {
		return common.ApiError{
			Code: common.FacilityMustHaveAtLeastOneCourt,
		}
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

	return common.CheckExistenceOrErrorByStringId(db, &User{}, f.UserId)
}
