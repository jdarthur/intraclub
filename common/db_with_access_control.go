package common

import "fmt"

type WithAccessControl[T CrudRecord] struct {
	Database          DatabaseProvider
	AccessControlUser RecordId
}

func NewWithAccessControl[T CrudRecord](db DatabaseProvider, accessControlUser RecordId) *WithAccessControl[T] {
	return &WithAccessControl[T]{
		Database:          db,
		AccessControlUser: accessControlUser,
	}
}

var SysAdminCheck func(db DatabaseProvider, c RecordId) (bool, error)

func (w WithAccessControl[T]) CanUserAccess(record T) bool {
	list := record.AccessibleTo(w.Database)
	if len(list) == 0 {
		fmt.Println("Access list is empty")
		return false
	}

	for _, userId := range list {
		if userId == w.AccessControlUser || userId == EveryoneRecordId {
			return true
		}
	}

	return w.CanUserEdit(record)
}

func (w WithAccessControl[T]) CanUserEdit(record T) bool {
	list := record.EditableBy(w.Database)
	if len(list) == 0 {
		fmt.Println("Editable-by list is empty")
		return false
	}

	cod, editIsConstrained := any(record).(CanOnlyDelete)

	isSysAdminEditable := false
	for _, userId := range list {
		if userId == w.AccessControlUser {
			if !editIsConstrained {
				// if EditableBy has no "user X can only delete" constraint, we can edit
				return true
			} else {
				if !cod.CanOnlyDelete(w.Database, w.AccessControlUser) {
					// if EditableBy has a "user X can only delete" constraint, but this
					// user doesn't have that constraint, we can edit.
					return true
				}
			}
		}

		if userId == SysAdminRecordId {
			isSysAdminEditable = true
		}
	}

	if isSysAdminEditable && SysAdminCheck != nil {
		if editIsConstrained && cod.CanOnlyDelete(w.Database, SysAdminRecordId) {
			return false
		}
		isSysAdmin, err := SysAdminCheck(w.Database, w.AccessControlUser)
		if err != nil {
			fmt.Println("error checking for sys admin", err)
			return false
		}

		return isSysAdmin
	}

	return false
}

func (w WithAccessControl[T]) GetAll(recordType T) ([]T, error) {
	filter := func(c T) bool { return w.CanUserAccess(c) }
	return GetAllWhere[T](w.Database, recordType, filter)
}

// GetOneById retrieves a CrudRecord by RecordId
func (w WithAccessControl[T]) GetOneById(record T, id RecordId) (t T, exists bool, err error) {
	t, exists, err = GetOneById(w.Database, record, id)
	if err != nil {
		return t, false, err
	}

	if !exists || !w.CanUserAccess(t) {
		return t, false, nil
	}

	return t, true, nil
}

func (w WithAccessControl[T]) DeleteOneById(record T, id RecordId) (t T, exists bool, err error) {
	t, exists, err = GetOneById(w.Database, record, id)
	if err != nil {
		return t, false, err
	}

	if !exists || !w.CanUserEdit(t) {
		return t, false, nil
	}

	return DeleteOneById(w.Database, record, id)
}

func (w WithAccessControl[T]) UpdateOneById(record T) (err error) {
	// check that a record exists with this ID. we will do this again in
	// the call to UpdateOne function below, but we need the common validate
	// and post-update logic that that function provides here.
	t, exists, err := GetOneById(w.Database, record, record.GetId())
	if err != nil {
		return err
	}

	// validate the record exists and the AccessControlUser can edit it
	if !exists || !w.CanUserEdit(t) {
		return fmt.Errorf("%s with ID %s was not found\n", record.Type(), record.GetId())
	}

	// update it, validating the type while doing so
	return UpdateOne(w.Database, record)
}
