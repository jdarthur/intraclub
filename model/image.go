package model

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

type Extension string

var PNG = Extension("png")
var JPG = Extension("jpg")
var GIF = Extension("gif")
var WEBP = Extension("webp")

var ValidExtensions = []Extension{PNG, JPG, GIF, WEBP}

type Image struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"` // unique ID for  this image
	UserId    string             `json:"user_id"`       // user ID of the person who uploaded this image
	Name      string             `json:"name"`          // name of the image
	Extension string             `json:"extension"`     // extension of the image (must be in ValidExtensions)
	Content   []byte             `json:"content"`       // raw bytes of the image
	AltText   string             `json:"alt_text"`      // optional alt text for this image that displays on mouseover
}

func (i *Image) RecordType() string {
	return "image"
}

func (i *Image) OneRecord() common.CrudRecord {
	return new(Image)
}

type listOfImages []*Image

func (l listOfImages) Length() int {
	return len(l)
}

func (l listOfImages) Get(index int) common.CrudRecord {
	return l[index]
}

func (i *Image) ListOfRecords() common.ListOfCrudRecords {
	return listOfImages{}
}

func (i *Image) SetId(id primitive.ObjectID) {
	i.ID = id
}

func (i *Image) GetId() primitive.ObjectID {
	return i.ID
}

func (i *Image) ValidateStatic() error {
	if i.Content == nil {
		return errors.New("content must not be empty")
	}

	if i.Name == "" {
		return errors.New("name must not be empty")
	}

	if i.UserId == "" {
		return errors.New("user ID must not be empty")
	}

	return i.validateExtension()
}

func (i *Image) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {
	return common.CheckExistenceOrErrorByStringId(db, &User{}, i.UserId)
}

func (i *Image) validateExtension() error {
	if i.Extension == "" {
		return errors.New("extension must not be empty")
	}

	valid := false
	for _, ext := range ValidExtensions {
		if i.Extension == string(ext) {
			valid = true
		}
	}

	if !valid {
		return fmt.Errorf("invalid extension: '%s'" + i.Extension)
	}

	return nil

}
