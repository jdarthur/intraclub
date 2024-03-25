package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"intraclub/common"
	"time"
)

type MongoDb struct {
	Hostname   string
	Username   string
	Password   string
	Connection *mongo.Database
}

func (m *MongoDb) GetAll(record common.CrudRecord) (objects interface{}, err error) {
	ctx, cancel := defaultTimeout()
	defer cancel()

	res, err := m.Connection.Collection(record.RecordType()).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	output := record.ListOfRecords()
	err = res.Decode(output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (m *MongoDb) GetOne(record common.CrudRecord) (object common.CrudRecord, exists bool, err error) {

	ctx, cancel := defaultTimeout()
	defer cancel()

	res, err := m.Connection.Collection(record.RecordType()).Find(ctx, byId(record.GetId()))
	if err != nil {
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
	//TODO implement me
	panic("implement me")
}

func (m *MongoDb) Update(object common.CrudRecord) error {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDb) Delete(record common.CrudRecord) error {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDb) Disconnect() error {
	//TODO implement me
	panic("implement me")
}

func defaultTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func byId(id string) bson.M {
	return bson.M{"_id": id}
}

func (m *MongoDb) Connect() error {

	//conn, err := mongo.Connect(defaultTimeout(), options.Client())

	//TODO implement me
	panic("implement me")
}

func NewMongoDbProvider(url, username, password string) common.DbProvider {
	return &MongoDb{
		Hostname: url,
		Username: username,
		Password: password,
	}
}
