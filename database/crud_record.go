package database

import (
	"time"
)

type CrudRecord interface {
	Type() string
	GetId() RecordId
	SetId(RecordId)
	Authorizable
	DatabaseValidatable
	NewRecord() CrudRecord
}

type TimestampedRecord interface {
	GetTimeStamps() (created, updated time.Time)
	SetCreateTimestamp(time.Time) (oldValue time.Time)
	SetUpdateTimestamp(time.Time) (oldValue time.Time)
}
