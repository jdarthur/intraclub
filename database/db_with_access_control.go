package database

import (
	"context"
	"fmt"
)

type WithAccessControl[T CrudRecord] struct {
	Database          Provider
	AccessControlUser UserId
}

func NewWithAccessControl[T CrudRecord](ctx context.Context, db Provider, accessControlUser UserId) *WithAccessControl[T] {
	return &WithAccessControl[T]{
		Database:          db,
		AccessControlUser: accessControlUser,
	}
}

var SysAdminCheck func(ctx context.Context, db Provider, userId UserId) (bool, error)

func (w *WithAccessControl[T]) CanUserAccess(record T) bool {
	list := record.AccessibleTo(context.Background(), w.Database)
	if len(list) == 0 {
		fmt.Println("Access list is empty")
		return false
	}

	for _, userId := range list {
		if userId == w.AccessControlUser || userId == EveryoneUserId {
			return true
		}
	}

	return w.CanUserEdit(record)
}

func (w *WithAccessControl[T]) CanUserEdit(record T) bool {
	list := record.EditableBy(context.Background(), w.Database)
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
				if !cod.CanOnlyDelete(context.Background(), w.Database, w.AccessControlUser) {
					// if EditableBy has a "user X can only delete" constraint, but this
					// user doesn't have that constraint, we can edit.
					return true
				}
			}
		}

		if userId == SysAdminUserId {
			isSysAdminEditable = true
		}
	}

	if isSysAdminEditable && SysAdminCheck != nil {
		if editIsConstrained && cod.CanOnlyDelete(context.Background(), w.Database, SysAdminUserId) {
			return false
		}
		isSysAdmin, err := SysAdminCheck(context.Background(), w.Database, w.AccessControlUser)
		if err != nil {
			fmt.Println("error checking for sys admin", err)
			return false
		}

		return isSysAdmin
	}

	fmt.Println("User cannot edit")
	return false
}

func (w *WithAccessControl[T]) GetAll(ctx context.Context) ([]T, error) {
	filter := func(_ context.Context, c T) bool { return w.CanUserAccess(c) }
	return GetAllWhere[T](ctx, w.Database, filter)
}

// GetOneById retrieves a CrudRecord by RecordId
func (w *WithAccessControl[T]) GetOneById(ctx context.Context, record T, id RecordId) (t T, exists bool, err error) {
	t, exists, err = GetOneById(ctx, w.Database, record, id)
	if err != nil {
		return t, false, err
	}

	if !exists || !w.CanUserAccess(t) {
		return t, false, nil
	}

	return t, true, nil
}

func (w *WithAccessControl[T]) DeleteOneById(ctx context.Context, record T, id RecordId) (t T, exists bool, err error) {
	t, exists, err = GetOneById(ctx, w.Database, record, id)
	if err != nil {
		return t, false, err
	}

	if !exists || !w.CanUserEdit(t) {
		fmt.Println("Does not exist or cannot edit")
		return t, false, nil
	}

	fmt.Println("deleting record", id)
	return DeleteOneById(ctx, w.Database, record, id)
}

func (w *WithAccessControl[T]) UpdateOneById(ctx context.Context, record T) (err error) {
	// check that a record exists with this ID. we will do this again in
	// the call to UpdateOne function below, but we need the common validate
	// and post-update logic that that function provides here.
	t, exists, err := GetOneById(ctx, w.Database, record, record.GetId())
	if err != nil {
		return err
	}

	// validate the record exists and the AccessControlUser can edit it
	if !exists || !w.CanUserEdit(t) {
		return fmt.Errorf("%s with ID %s was not found\n", record.Type(), record.GetId())
	}

	// make sure the owner is set on the record after we have validated that
	// this user can edit
	if record.GetOwner() == InvalidUserId {
		record.SetOwner(t.GetOwner())
	}

	// update it, validating the type while doing so
	return UpdateOne(ctx, w.Database, record)
}
