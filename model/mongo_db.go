package model

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"intraclub/common"
	"time"
)

type MongoDb struct {
	Hostname   string
	Username   string
	Password   string
	Connection *mongo.Database
}

var IntraclubMongoDatabase = "intraclub"

func (m *MongoDb) GetAllWhere(record common.CrudRecord, filter map[string]interface{}) (objects common.ListOfCrudRecords, err error) {

	ctx, cancel := defaultTimeout()
	defer cancel()

	res, err := m.Connection.Collection(record.RecordType()).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	output := record.ListOfRecords()
	err = res.All(ctx, &output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (m *MongoDb) GetAll(record common.CrudRecord) (objects common.ListOfCrudRecords, err error) {
	return m.GetAllWhere(record, bson.M{})
}

func (m *MongoDb) GetOne(record common.CrudRecord) (object common.CrudRecord, exists bool, err error) {

	ctx, cancel := defaultTimeout()
	defer cancel()

	res := m.Connection.Collection(record.RecordType()).FindOne(ctx, byId(record.GetId()))
	if res.Err() != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, false, nil
		}
		return nil, false, err
	}

	output := record.OneRecord()
	err = res.Decode(output)
	if err != nil {
		return nil, false, err
	}

	return output, true, nil

}

func (m *MongoDb) Create(object common.CrudRecord) (common.CrudRecord, error) {

	ctx, cancel := defaultTimeout()
	defer cancel()

	object.SetId(primitive.NewObjectID())

	inserted, err := m.Connection.Collection(object.RecordType()).InsertOne(ctx, object)
	if err != nil {
		return nil, err
	}

	object.SetId(inserted.InsertedID.(primitive.ObjectID))

	return object, nil

}

func (m *MongoDb) Update(object common.CrudRecord) error {
	ctx, cancel := defaultTimeout()
	defer cancel()

	v, err := m.Connection.Collection(object.RecordType()).UpdateOne(ctx, byId(object.GetId()), bson.M{"$set": object})
	if err != nil {
		return err
	}

	fmt.Println(v)

	if v.ModifiedCount == 0 {
		return errors.New("updated count was 0")
	}

	return nil
}

func (m *MongoDb) Delete(record common.CrudRecord) error {

	ctx, cancel := defaultTimeout()
	defer cancel()

	deleted, err := m.Connection.Collection(record.RecordType()).DeleteOne(ctx, byId(record.GetId()))
	if err != nil {
		return err
	}

	if deleted.DeletedCount == 0 {
		return errors.New("deleted count was 0")
	}

	return nil
}

func (m *MongoDb) Disconnect() error {
	ctx, cancel := defaultTimeout()
	defer cancel()
	return m.Connection.Client().Disconnect(ctx)
}

func defaultTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func byId(id primitive.ObjectID) bson.M {
	return bson.M{"_id": id}
}

func (m *MongoDb) Connect() error {

	ctx, cancel := defaultTimeout()
	defer cancel()

	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(m.Hostname))
	if err != nil {
		return err
	}

	m.Connection = conn.Database(IntraclubMongoDatabase)
	return nil
}

func NewMongoDbProvider(url, username, password string) common.DbProvider {
	return &MongoDb{
		Hostname: url,
		Username: username,
		Password: password,
	}
}
