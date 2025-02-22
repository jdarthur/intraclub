package model

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
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

func (m *MongoDb) GetAll(recordType common.CrudRecord) ([]common.CrudRecord, error) {
	return m.GetAllWhere(recordType, nil)
}

func (m *MongoDb) GetAllWhere(recordType common.CrudRecord, where common.WhereFunc) ([]common.CrudRecord, error) {
	ctx, cancel := defaultTimeout()
	defer cancel()

	res, err := m.Connection.Collection(recordType.Type()).Find(ctx, nil)
	if err != nil {
		return nil, err
	}

	records := ListOfCrudRecords(recordType)
	err = res.All(ctx, records)
	if err != nil {
		return nil, err
	}

	output := make([]common.CrudRecord, 0)
	for _, record := range records {
		if where == nil || where(record) {
			output = append(output, record)
		}
	}

	return output, nil
}

var IntraclubMongoDatabase = "intraclub"

func (m *MongoDb) GetOne(record common.CrudRecord) (object common.CrudRecord, exists bool, err error) {

	ctx, cancel := defaultTimeout()
	defer cancel()

	res := m.Connection.Collection(record.Type()).FindOne(ctx, byId(record.GetId()))
	if res.Err() != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, false, nil
		}
		return nil, false, err
	}

	err = res.Decode(record)
	if err != nil {
		return nil, false, err
	}
	return record, true, nil

}

func (m *MongoDb) Create(object common.CrudRecord) (common.CrudRecord, error) {

	ctx, cancel := defaultTimeout()
	defer cancel()

	object.SetId(common.NewRecordId())

	_, err := m.Connection.Collection(object.Type()).InsertOne(ctx, object)
	if err != nil {
		return nil, err
	}

	return object, nil

}

func (m *MongoDb) Update(object common.CrudRecord) error {
	ctx, cancel := defaultTimeout()
	defer cancel()

	v, err := m.Connection.Collection(object.Type()).UpdateOne(ctx, byId(object.GetId()), bson.M{"$set": object})
	if err != nil {
		return err
	}

	if v.MatchedCount == 0 {
		return errors.New("matched count was 0")
	}

	return nil
}

func (m *MongoDb) Delete(record common.CrudRecord) error {

	ctx, cancel := defaultTimeout()
	defer cancel()

	deleted, err := m.Connection.Collection(record.Type()).DeleteOne(ctx, byId(record.GetId()))
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

func byId(id common.RecordId) bson.M {
	return bson.M{"_id": id.String()}
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

func NewMongoDbProvider(url, username, password string) common.DatabaseProvider {
	return &MongoDb{
		Hostname: url,
		Username: username,
		Password: password,
	}
}

func ListOfCrudRecords[T common.CrudRecord](record T) []T {
	return make([]T, 0)
}
