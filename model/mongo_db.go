package model

import (
	"context"
	"errors"
	"reflect"
	"time"

	"intraclub/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDb struct {
	Hostname   string
	Username   string
	Password   string
	Connection *mongo.Database
}

func (m *MongoDb) GetAll(ctx context.Context, recordType database.CrudRecord) ([]database.CrudRecord, error) {
	return m.GetAllWhere(ctx, recordType, nil)
}

func (m *MongoDb) GetAllWhere(ctx context.Context, recordType database.CrudRecord, where database.WhereFunc) ([]database.CrudRecord, error) {

	res, err := m.Connection.Collection(recordType.Type()).Find(ctx, nil)
	if err != nil {
		return nil, err
	}

	blank := recordType.BlankRecord()
	sliceType := reflect.SliceOf(reflect.TypeOf(blank))
	slice := reflect.MakeSlice(sliceType, 0, 0)
	ptr := reflect.New(sliceType)
	ptr.Elem().Set(slice)

	err = res.All(ctx, ptr.Interface())
	if err != nil {
		return nil, err
	}

	output := make([]database.CrudRecord, 0, ptr.Elem().Len())
	sliceVal := ptr.Elem()
	for i := 0; i < sliceVal.Len(); i++ {
		record := sliceVal.Index(i).Elem().Interface().(database.CrudRecord)
		if where == nil || where(ctx, record) {
			output = append(output, record)
		}
	}

	return output, nil
}

var IntraclubMongoDatabase = "intraclub"

func (m *MongoDb) GetOne(ctx context.Context, record database.CrudRecord) (object database.CrudRecord, exists bool, err error) {

	blank := record.BlankRecord()

	res := m.Connection.Collection(record.Type()).FindOne(ctx, byId(record.GetId()))
	if res.Err() != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, false, nil
		}
		return nil, false, err
	}

	err = res.Decode(blank)
	if err != nil {
		return nil, false, err
	}
	return blank, true, nil

}

func (m *MongoDb) Create(ctx context.Context, object database.CrudRecord) (database.CrudRecord, error) {

	object.SetId(database.NewRecordId())

	_, err := m.Connection.Collection(object.Type()).InsertOne(ctx, object)
	if err != nil {
		return nil, err
	}

	return object, nil

}

func (m *MongoDb) Update(ctx context.Context, object database.CrudRecord) error {

	v, err := m.Connection.Collection(object.Type()).UpdateOne(ctx, byId(object.GetId()), bson.M{"$set": object})
	if err != nil {
		return err
	}

	if v.MatchedCount == 0 {
		return errors.New("matched count was 0")
	}

	return nil
}

func (m *MongoDb) Delete(ctx context.Context, record database.CrudRecord) error {

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

func byId(id database.RecordId) bson.M {
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

func NewMongoDbProvider(url, username, password string) database.DatabaseProvider {
	return &MongoDb{
		Hostname: url,
		Username: username,
		Password: password,
	}
}
