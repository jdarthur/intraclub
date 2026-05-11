package database

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"strings"
	"time"
)

func init() {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		mrand.Seed(time.Now().UnixNano())
	} else {
		mrand.Seed(int64(binary.BigEndian.Uint64(b)))
	}
}

type RecordId uint64

func (r *RecordId) UnmarshalJSON(bytes []byte) error {
	v, err := RecordIdFromString(strings.Trim(string(bytes), "\""))
	if err != nil {
		return err
	}

	*r = v
	return nil
}

func (r RecordId) MarshalJSON() ([]byte, error) {
	str := "\""
	str += r.String()
	str += "\""
	return []byte(str), nil
}

func NewRecordId() RecordId {
	return RecordId(mrand.Uint64() + uint64(len(unavailableRecordIds)))
}

func (r RecordId) Uint64() uint64 {
	return uint64(r)
}

func (r RecordId) String() string {
	if r == InvalidRecordId {
		return ""
	}
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

func UnmarshalStringIdList(bytes []byte) ([]RecordId, error) {
	l := make([]string, 0)
	err := json.Unmarshal(bytes, &l)
	if err != nil {
		return nil, err
	}
	l2 := make([]RecordId, 0, len(l))
	for _, id := range l {
		recordId, err := RecordIdFromString(id)
		if err != nil {
			return nil, err
		}
		l2 = append(l2, recordId)
	}
	return l2, nil
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

type Authorizable interface {
	EditableBy(ctx context.Context, db DatabaseProvider) []RecordId
	AccessibleTo(ctx context.Context, db DatabaseProvider) []RecordId
	SetOwner(recordId RecordId)
	GetOwner() RecordId
}

type CrudRecord interface {
	Type() string
	GetId() RecordId
	SetId(RecordId)
	Authorizable
	DatabaseValidatable
	BlankRecord() CrudRecord
}

type PostCreate interface {
	PostCreate(ctx context.Context, db DatabaseProvider) error // function to call post-create
}

type PreUpdate interface {
	PreUpdate(ctx context.Context, db DatabaseProvider, existingValues CrudRecord) error // function to call pre-update
}

type CanOnlyDelete interface {
	CrudRecord
	CanOnlyDelete(ctx context.Context, db DatabaseProvider, userId RecordId) bool
}

type PostUpdate interface {
	PostUpdate(ctx context.Context, db DatabaseProvider) error // function to call post-update
}

type PreDelete interface {
	PreDelete(ctx context.Context, db DatabaseProvider) error // function to call pre-delete
}

type PostDelete interface {
	PostDelete(ctx context.Context, db DatabaseProvider) error // function to call post-delete
}

type TimestampedRecord interface {
	GetTimeStamps() (created, updated time.Time)
	SetCreateTimestamp(time.Time) (oldValue time.Time)
	SetUpdateTimestamp(time.Time) (oldValue time.Time)
}
