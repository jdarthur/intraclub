package model

import (
	"context"
	"errors"
	"strings"
	"time"

	"intraclub/database"
)

// OrganizationId is a wrapper around the common.RecordId type
// which refers specifically to the primary key for the Organization
// struct. Other records referring to this type (as opposed to
// common.RecordId) allows better code navigation, enabling us
// to automatically determine which structs depend on Organization.
type OrganizationId database.RecordId

func (id OrganizationId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id OrganizationId) String() string {
	return id.RecordId().String()
}

func (id OrganizationId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id *OrganizationId) UnmarshalJSON(bytes []byte) error {
	rid := id.RecordId()
	if err := (*database.RecordId)(&rid).UnmarshalJSON(bytes); err != nil {
		return err
	}
	*id = OrganizationId(rid)
	return nil
}

// Organization is an evergreen, flat roster of registered users grouped
// together season-over-season (e.g. "Martin's Landing Men"). It is owned
// by the creating UserId, but is publicly accessible to all users so that
// memberships can be read by anyone.
//
// An Organization must have a non-empty Name, which is unique across all
// organizations (to prevent duplicate records). The membership list is a
// flat list of User records via the organization_member join table.
type Organization struct {
	ID        OrganizationId `json:"id"`       // Unique ID for this Organization
	UserId    database.UserId `json:"user_id"` // ID of the User who owns the record
	Name      string          `json:"name"`    // Unique name for the Organization
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at"`
}

func (o *Organization) GetOwner() database.UserId {
	return o.UserId
}

func (o *Organization) UniquenessEquivalent(other *Organization) error {
	if o.Name == other.Name {
		return errors.New("duplicate record for organization name")
	}
	return nil
}

// NewOrganization allocates a new *Organization record. Calling this function
// (as opposed to doing e.g. `v := &Organization{}`) allows us to easily
// navigate to all the points in the code which allocate a new Organization.
func NewOrganization() *Organization {
	return &Organization{}
}

// SetOwner assigns the owner of this common.CrudRecord
func (o *Organization) SetOwner(userId database.UserId) {
	o.UserId = userId
}

// EditableBy returns a list of common.UserId values who are allowed
// to edit (or possibly delete) this common.CrudRecord
func (o *Organization) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	// This record can only be edited by the owner (mirrors Facility).
	return database.SysAdminAndUsers(o.UserId)
}

// AccessibleTo returns a list of common.UserId values who are allowed
// to view this record (in this instance, all users, regardless of their
// authentication status)
func (o *Organization) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

// Type is the database table name for this record
func (o *Organization) Type() string {
	return "organization"
}

// GetId returns a unique ID for this record
func (o *Organization) GetId() database.RecordId {
	return o.ID.RecordId()
}

// SetId sets a unique ID for this record
func (o *Organization) SetId(id database.RecordId) {
	o.ID = OrganizationId(id)
}

// StaticallyValid validates this record against the record-specific
// business logic rules without requiring the caller to provide a
// common.DatabaseProvider for database validation
func (o *Organization) StaticallyValid() error {
	o.Name = strings.TrimSpace(o.Name)

	if o.Name == "" {
		return errors.New("organization name is empty")
	}
	return nil
}

// DynamicallyValid validates this record against the record-specific
// business logic rules using a common.DatabaseProvider to validate e.g.
// individual ID values for existence, ownership constraints, etc.
func (o *Organization) DynamicallyValid(ctx context.Context, db database.Provider) error {
	return nil
}

// Timestamps returns the create, update and delete timestamps for this
// Organization record.
func (o *Organization) Timestamps() (time.Time, time.Time, *time.Time) {
	return o.CreatedAt, o.UpdatedAt, o.DeletedAt
}

func (o *Organization) SetCreatedAt(createdAt time.Time) {
	o.CreatedAt = createdAt
}

func (o *Organization) SetUpdatedAt(updatedAt time.Time) {
	o.UpdatedAt = updatedAt
}

func (o *Organization) SetDeletedAt(deletedAt *time.Time) {
	o.DeletedAt = deletedAt
}

func (o *Organization) NewRecord() database.CrudRecord {
	return new(Organization)
}
