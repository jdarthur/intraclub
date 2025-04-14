package common

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type RecordId uint64

func (r RecordId) UnmarshalJSON(bytes []byte) error {
	var err error
	r, err = RecordIdFromString(strings.Trim(string(bytes), "\""))
	if err != nil {
		return err
	}

	fmt.Println(err)
	return nil
}

func (r RecordId) MarshalJSON() ([]byte, error) {
	str := "\""
	str += r.String()
	str += "\""
	return []byte(str), nil
}

func NewRecordId() RecordId {
	return RecordId(rand.Uint64() + uint64(len(unavailableRecordIds)))
}

func (r RecordId) Uint64() uint64 {
	return uint64(r)
}

func (r RecordId) String() string {
	return fmt.Sprintf("%016x", r.Uint64())
}

func (r RecordId) ValidRecordId() bool {
	return uint64(r) > uint64(len(unavailableRecordIds))
}

func RecordIdFromString(s string) (RecordId, error) {
	if s == "" {
		return InvalidRecordId, nil
	}
	b := make([]byte, 8)
	n, err := hex.Decode(b, []byte(s))
	if err != nil {
		return InvalidRecordId, err
	}
	if n != 8 {
		return InvalidRecordId, fmt.Errorf("short read on hex.Decode")
	}
	return RecordId(binary.BigEndian.Uint64(b)), nil
}

// InvalidRecordId is a special record ID that indicates a value hasn't been set
var InvalidRecordId = RecordId(0)

// EveryoneRecordId indicates that a CrudRecord is accessible by everyone
var EveryoneRecordId = RecordId(1)
var AccessibleToEveryone = []RecordId{EveryoneRecordId}

// SysAdminRecordId indicated that a CrudRecord is accessible / editable by users
// the model.SystemAdministrator role applied to their model.User record
var SysAdminRecordId = RecordId(2)

func SysAdminAndUsers(users ...RecordId) []RecordId {
	recordIds := make([]RecordId, 0, 1+len(users))
	recordIds = append(recordIds, SysAdminRecordId)
	recordIds = append(recordIds, users...)
	return recordIds
}

// unavailableRecordIds is a list of RecordId values that cannot be set
// in NewRecordId because they have a special meaning in the auth/access logic
var unavailableRecordIds = []RecordId{
	InvalidRecordId, EveryoneRecordId, SysAdminRecordId,
}

type CrudRecord interface {
	Type() string
	GetId() RecordId
	SetId(RecordId)
	EditableBy(db DatabaseProvider) []RecordId
	AccessibleTo(db DatabaseProvider) []RecordId
	SetOwner(recordId RecordId)
	GetOwner() RecordId
	DatabaseValidatable
}

type PostCreate interface {
	PostCreate(db DatabaseProvider) error // function to call post-create
}

type PreUpdate interface {
	PreUpdate(db DatabaseProvider, existingValues CrudRecord) error // function to call pre-update
}

type CanOnlyDelete interface {
	CrudRecord
	CanOnlyDelete(db DatabaseProvider, userId RecordId) bool
}

type PostUpdate interface {
	PostUpdate(db DatabaseProvider) error // function to call post-update
}

type PreDelete interface {
	PreDelete(db DatabaseProvider) error // function to call pre-delete
}

type PostDelete interface {
	PostDelete(db DatabaseProvider) error // function to call post-delete
}

type TimestampedRecord interface {
	GetTimeStamps() (created, updated time.Time)
	SetCreateTimestamp(time.Time) (oldValue time.Time)
	SetUpdateTimestamp(time.Time) (oldValue time.Time)
}
