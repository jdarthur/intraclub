package database

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
)

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
	return RecordId(rand.Uint64() + uint64(len(unavailableRecordIds)))
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
	// A RecordId is 8 bytes = 16 hex chars. Reject anything that isn't exactly
	// that length BEFORE decoding; hex.Decode into a fixed 8-byte buffer would
	// otherwise write past the end of the slice and panic (runtime bounds
	// check) when handed a longer even-length string.
	if len(s) != 16 {
		return InvalidRecordId, fmt.Errorf("invalid record ID %q: must be exactly 16 hex characters", s)
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
