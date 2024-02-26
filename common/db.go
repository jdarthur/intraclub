package common

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Collection struct {
	PrettyName string
	DbName     string
}

type Database struct {
	Connection *mongo.Database
}

var Db Database

type Record interface {
	GetId() primitive.ObjectID
	SetId(id primitive.ObjectID)
	Collection() Collection
}

func (d Database) GetOne(value Record) (exists bool, err error) {
	ctx, cancel := defaultTimeout()
	defer cancel()

	res := d.Connection.Collection(value.Collection().DbName).FindOne(ctx, byId(value.GetId(), ""))
	if res.Err() != nil {
		return false, res.Err()
	}

	err = res.Decode(value)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (d Database) GetAll(c Collection, value []Record) error {

	ctx, cancel := defaultTimeout()
	defer cancel()

	res, err := d.Connection.Collection(c.DbName).Find(ctx, value)
	if err != nil {
		return err
	}

	err = res.Decode(value)
	if err != nil {
		return err
	}

	return nil
}

func (d Database) Create(value Record) error {
	existingId := value.GetId()
	if !existingId.IsZero() {
		return errors.New("ID value must not be set when creating a new record")
	}

	newId := primitive.NewObjectID()
	value.SetId(newId)

	ctx, cancel := defaultTimeout()
	defer cancel()

	exists, err := d.GetOne(value)
	if err != nil {
		return err
	}
	if exists {
		return errors.New(fmt.Sprintf("Record with ID %s already exists", newId))
	}

	_, err = d.Connection.Collection(value.Collection().DbName).InsertOne(ctx, value)
	if err != nil {
		return err
	}

	return nil
}

func (d Database) Update(value Record) error {
	// if record does not exist, return 404

	// update the DB record according to the new values
	return nil
}

func (d Database) Delete(value Record) error {
	// if record does not exist, return 404

	// delete the record in value.Collection() where ID matches value.GetId()
	return nil
}

func byId(id primitive.ObjectID, username string) bson.M {
	return bson.M{"_id": id}
}

func defaultTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
