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

// UserId is a first-class type representing a user identifier in the system.
// It is used throughout the authorization and access control subsystems.
type UserId RecordId

func (id UserId) RecordId() RecordId {
	return RecordId(id)
}

func (id UserId) String() string {
	return RecordId(id).String()
}

func (id UserId) MarshalJSON() ([]byte, error) {
	return RecordId(id).MarshalJSON()
}

func (id *UserId) UnmarshalJSON(bytes []byte) error {
	v, err := RecordIdFromString(strings.Trim(string(bytes), "\""))
	if err != nil {
		return err
	}
	*id = UserId(v)
	return nil
}

// InvalidRecordId is a special record ID that indicates a value hasn't been set
var InvalidRecordId = RecordId(0)

// InvalidUserId is a special user ID that indicates a value hasn't been set
var InvalidUserId = UserId(0)

// EveryoneUserId indicates that a CrudRecord is accessible by everyone
var EveryoneUserId = UserId(1)
var AccessibleToEveryone = []UserId{EveryoneUserId}

// SysAdminUserId indicates that a CrudRecord is accessible / editable by users
// with the SystemAdministrator role applied to their User record
var SysAdminUserId = UserId(2)

func SysAdminAndUsers(users ...UserId) []UserId {
	userIds := make([]UserId, 0, 1+len(users))
	userIds = append(userIds, SysAdminUserId)
	userIds = append(userIds, users...)
	return userIds
}

// unavailableRecordIds is a list of RecordId values that cannot be set
// in NewRecordId because they have a special meaning in the auth/access logic
var unavailableRecordIds = []RecordId{
	InvalidRecordId, RecordId(EveryoneUserId), RecordId(SysAdminUserId),
}

type Authorizable interface {
	EditableBy(ctx context.Context, db DatabaseProvider) []UserId
	AccessibleTo(ctx context.Context, db DatabaseProvider) []UserId
	SetOwner(userId UserId)
	GetOwner() UserId
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
	CanOnlyDelete(ctx context.Context, db DatabaseProvider, userId UserId) bool
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
