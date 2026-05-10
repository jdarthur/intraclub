package common

import (
	"context"
	"fmt"
	"time"
)

// WhereFunc is a function signature used in DatabaseProvider.GetAllWhere
// which provides a CrudRecord and allows the caller to define a function
// to potentially filter the result from the output list. If the WhereFunc
// returns true, the record will be returned in the output.
type WhereFunc func(ctx context.Context, c CrudRecord) bool

// WhereFuncT is a typed version of WhereFunc which is used in the typed
// DatabaseProvider wrapper functions.
type WhereFuncT[T CrudRecord] func(ctx context.Context, c T) bool

// DatabaseProvider is a generic interface to allow CRUD operations
// on a CrudRecord in a database without depending on database-specific
// function signatures, logic, etc. This also allows us some nice
// type-safe functions such as GetOneById which don't require a caller
// to do any manual type-casting after querying something from the DB
type DatabaseProvider interface {

	// GetOne returns the CrudRecord with a matching RecordId to the
	// record in the function signature, false if one does not exist,
	// and an error if we encountered one during the DB query
	GetOne(ctx context.Context, record CrudRecord) (CrudRecord, bool, error)

	// GetAll returns a list of all records of a certain CrudRecord type
	GetAll(ctx context.Context, recordType CrudRecord) ([]CrudRecord, error)

	// GetAllWhere returns all records of a certain CrudRecord type
	// which match the filtering logic provided in the given WhereFunc
	GetAllWhere(ctx context.Context, recordType CrudRecord, where WhereFunc) ([]CrudRecord, error)

	// Create creates a new CrudRecord in the database
	Create(ctx context.Context, record CrudRecord) (CrudRecord, error)

	// Update updates an existing CrudRecord by the record's RecordId
	Update(ctx context.Context, record CrudRecord) error

	// Delete deletes a CrudRecord by the record's RecordId (if it exists)
	Delete(ctx context.Context, record CrudRecord) error
}

// GetOneById is a typed version of DatabaseProvider.GetOne that takes a RecordId
// returning the matching record from the DatabaseProvider (if exists), a bool
// indicating whether the target record exists, and an error if one was encountered
// during the query on the DatabaseProvider
func GetOneById[T CrudRecord](ctx context.Context, db DatabaseProvider, record T, id RecordId) (t T, exists bool, err error) {
	record.SetId(id)
	r, exists, err := db.GetOne(ctx, record)
	if err != nil {
		return t, false, err
	}
	if !exists {
		return t, false, nil
	}
	return r.(T), exists, nil
}

func GetExistingRecordById[T CrudRecord](ctx context.Context, db DatabaseProvider, record T, id RecordId) (t T, err error) {
	record.SetId(id)
	r, exists, err := db.GetOne(ctx, record)
	if err != nil {
		return t, err
	}
	if !exists {
		return t, fmt.Errorf("'%s' record with ID '%s' does not exist", record.Type(), id)

	}
	return r.(T), nil
}

// ExistsById checks if a CrudRecord exists with a particular RecordId,
// returning an error if encountered, or if the record does not exist
func ExistsById[T CrudRecord](ctx context.Context, db DatabaseProvider, record T, id RecordId) (err error) {
	record.SetId(id)
	_, exists, err := db.GetOne(ctx, record)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("'%s' record with ID '%s' does not exist", record.Type(), id)
	}
	return nil
}

// FromListOfCrudRecord is a type-parameterized helper function which is used to
// convert from the []CrudRecord returned by the Get All functions on a struct
// implementing DatabaseProvider into a []T for type-safety downstream of the call
func fromListOfCrudRecord[T CrudRecord](list []CrudRecord) ([]T, error) {
	output := make([]T, 0, len(list))
	for _, v := range list {
		output = append(output, v.(T))
	}
	return output, nil
}

// GetAll is a typed version of DatabaseProvider.GetAll that gets all records
// of type T, then converts the resulting []CrudRecord into a []T
func GetAll[T CrudRecord](ctx context.Context, db DatabaseProvider) ([]T, error) {
	var zero T
	v, err := db.GetAll(ctx, zero)
	if err != nil {
		return nil, err
	}
	return fromListOfCrudRecord[T](v)
}

// GetAllWhere is a type-safe wrapper around DatabaseProvider.GetAllWhere,
// returning a []T instead of a []CrudRecord
func GetAllWhere[T CrudRecord](ctx context.Context, db DatabaseProvider, where WhereFuncT[T]) ([]T, error) {
	var zero T
	var w WhereFunc
	if where != nil {
		w = func(_ context.Context, c CrudRecord) bool {
			return where(ctx, c.(T))
		}
	}

	v, err := db.GetAllWhere(ctx, zero, w)
	if err != nil {
		return nil, err
	}
	return fromListOfCrudRecord[T](v)
}

// DeleteOneById deletes a CrudRecord from the given DatabaseProvider by the provided RecordId,
// and returns the record back to the caller (if it existed), or an error if encountered during
// the query / delete operations on the DatabaseProvider
func DeleteOneById[T CrudRecord](ctx context.Context, db DatabaseProvider, record T, id RecordId) (t T, exists bool, err error) {

	// check if a record with the given RecordId exists for the CrudRecord type
	record.SetId(id)
	r, exists, err := db.GetOne(ctx, record)
	if err != nil {
		return t, false, err
	}

	// return no error but exists==false if there is no CrudRecord with that ID
	if !exists {
		return t, false, nil
	}

	// run post-create logic if the record type implements it
	o, ok := r.(PreDelete)
	if ok {
		err = o.PreDelete(ctx, db)
		if err != nil {
			return t, false, err
		}
	}

	// delete the record from the DatabaseProvider if it does exist currently
	err = db.Delete(ctx, r)
	if err != nil {
		return t, false, err
	}

	// run post-delete logic if the record type implements it
	pd, ok := r.(PostDelete)
	if ok {
		err = pd.PostDelete(ctx, db)
		if err != nil {
			return t, false, err
		}
	}

	// return the now-deleted record back to the caller
	return r.(T), exists, nil
}

// CreateOne validates that a CrudRecord is statically and dynamically
// valid, creates a new RecordId for the new record's primary key, and
// saves the record to the given DatabaseProvider. If the record has
// any post-create logic to run (via implementing the PostCreate interface)
// this logic is also run, returning any errors encountered along the way
func CreateOne[T CrudRecord](ctx context.Context, db DatabaseProvider, record T) (t T, err error) {

	// create a new record ID
	recordId := NewRecordId()
	record.SetId(recordId)

	setCreateTimeStampIfApplicable(record)

	// validate that this record meets all the constraints of its type
	err = Validate(ctx, db, record)
	if err != nil {
		return t, err
	}

	err = ValidateUniqueConstraint(ctx, db, record)
	if err != nil {
		return t, err
	}

	v, err := db.Create(ctx, record)
	if err != nil {
		return t, err
	}

	// run post-create logic if the record type implements it
	o, ok := v.(PostCreate)
	if ok {
		err = o.PostCreate(ctx, db)
		if err != nil {
			return t, err
		}
	}

	// return created object back to the caller (so they can get the
	// newly-create RecordId that was generated on creation)
	return v.(T), nil
}

// UpdateOne takes an updated CrudRecord, validates that the update
// does not violate any type-specific constraints, updates it in the
// given DatabaseProvider and runs any post-create logic that the
// type implements, returning any errors encountered along the way
func UpdateOne[T CrudRecord](ctx context.Context, db DatabaseProvider, record T) (err error) {
	// fetch existing record for timestamp preservation and PreUpdate hook
	v, exists, err := GetOneById(ctx, db, record, record.GetId())
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s record with ID %s does not exist", record.Type(), record.GetId())
	}

	// set update timestamp before validating (in case validation checks for timestamp values)
	setUpdateTimeStampIfApplicable(record, v)

	// validate that this record meets all the constraints of its type
	err = Validate(ctx, db, record)
	if err != nil {
		return err
	}

	err = ValidateUniqueConstraint(ctx, db, record)
	if err != nil {
		return err
	}

	pu, ok := any(record).(PreUpdate)
	if ok {
		err = pu.PreUpdate(ctx, db, v)
		if err != nil {
			return err
		}
	}
	// update the record in the DB
	err = db.Update(ctx, record)
	if err != nil {
		return err
	}

	// run post-update logic if implemented by the type
	o, ok := any(record).(PostUpdate)
	if ok {
		err = o.PostUpdate(ctx, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func setCreateTimeStampIfApplicable[T CrudRecord](record T) {
	ts, ok := any(record).(TimestampedRecord)
	if ok {
		ts.SetCreateTimestamp(time.Now())
	}
}

// setUpdateTimeStampIfApplicable will check if the record is a TimestampedRecord, and if
// so, it will set the update timestamp. This function will also overwrite the created
// timestamp on the update body with the value from the existing record, so that the
// created timestamp remains immutable during an update operation
func setUpdateTimeStampIfApplicable[T CrudRecord](updatedRecord, existingRecord T) {
	ts, ok := any(updatedRecord).(TimestampedRecord)
	if !ok {
		return
	}

	existingCreated, _ := any(existingRecord).(TimestampedRecord).GetTimeStamps()
	ts.SetCreateTimestamp(existingCreated)
	ts.SetUpdateTimestamp(time.Now())
}
