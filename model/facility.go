package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"strings"
)

// FacilityId is a wrapper around the common.RecordId type
// which refers specifically to the primary key for the Facility
// struct. Other records referring to this type (as opposed to
// common.RecordId) allows better code navigation, enabling us
// to automatically determine which structs depend on Facility
type FacilityId common.RecordId

func (id FacilityId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id FacilityId) String() string {
	return id.RecordId().String()
}

// Facility is a physical location where a Season is played.
// It is owned by a particular UserId, but is publicly
// accessible to all users (so that multiple seasons may
// share the same fixed FacilityId).
//
// A Facility must have a Name, Address, and a non-zero
// NumberOfCourts. It may also have a
type Facility struct {
	ID             FacilityId // Unique ID for this Facility
	UserId         UserId     // ID of the User who owns the record
	Name           string     // Unique name for the Facility (to prevent duplicate records)
	Address        string     // Unique street address for the Facility (to prevent duplicate records)
	NumberOfCourts int        // Number of courts available at the Facility
	LayoutPhoto    PhotoId    // ID of a Photo showing the layout of the Facility (i.e. orientation of courts, parking, etc.)
}

func (f *Facility) UniquenessEquivalent(other *Facility) error {
	if f.Name == other.Name {
		return fmt.Errorf("duplicate record for facility name")
	}
	if f.Address == other.Address {
		return fmt.Errorf("duplicate record for facility address")
	}
	return nil
}

// NewFacility allocates a new *Facility record. Calling this function
// (as opposed to doing e.g. `v := &Facility{}`) allows us to easily
// navigate to all the points in the code which allocate a new Facility
func NewFacility() *Facility {
	return &Facility{}
}

// PreDelete validates that this Facility is not in use by any existing
// Season. If it is assigned to a Season, then it may not be deleted as the
// Facility information is viewable and potentially important to the participants
func (f *Facility) PreDelete(db common.DatabaseProvider) error {
	inUse, err := f.IsFacilityInUse(db)
	if err != nil {
		return err
	}
	if inUse {
		return fmt.Errorf("facility %s is in use", f.ID)
	}
	return nil
}

// SetOwner assigns the owner of this common.CrudRecord
func (f *Facility) SetOwner(recordId common.RecordId) {
	f.UserId = UserId(recordId)
}

// EditableBy returns a list of common.RecordId values who are allowed
// to edit (or possibly delete) this common.CrudRecord
func (f *Facility) EditableBy(common.DatabaseProvider) []common.RecordId {
	// This record can only be edited by the owner. It should
	// probably be created once and reused many times without
	// modification, so it is unlikely that updates will occur
	// very often. It also may not be deleted after assignment
	// to a particular season (as described in PreDelete)
	return common.SysAdminAndUsers(f.UserId.RecordId())
}

// AccessibleTo returns a list of common.RecordId values who are allowed
// to view this record (in this instance, all users, regardless of their
// authentication status)
func (f *Facility) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

// Type is the database table name for this record
func (f *Facility) Type() string {
	return "facility"
}

// GetId returns a unique ID for this record
func (f *Facility) GetId() common.RecordId {
	return f.ID.RecordId()
}

// SetId sets a unique ID for this record
func (f *Facility) SetId(id common.RecordId) {
	f.ID = FacilityId(id)
}

// StaticallyValid validates this record against the record-specific
// business logic rules without requiring the caller to provide a
// common.DatabaseProvider for database validation
func (f *Facility) StaticallyValid() error {
	f.Name = strings.TrimSpace(f.Name)
	f.Address = strings.TrimSpace(f.Address)

	if f.Name == "" {
		return errors.New("facility name is empty")
	}
	if f.NumberOfCourts <= 0 {
		return errors.New("facility number of courts must be greater than zero")
	}
	if f.Address == "" {
		return errors.New("facility address is empty")
	}
	return nil
}

// DynamicallyValid validates this record against the record-specific
// business logic rules using a common.DatabaseProvider to validate e.g.
// individual ID values for existence, ownership constraints, etc.
func (f *Facility) DynamicallyValid(db common.DatabaseProvider) error {
	if f.LayoutPhoto != 0 {
		return common.ExistsById(db, &Photo{}, f.LayoutPhoto.RecordId())
	}
	return nil
}

// IsFacilityInUse checks if this Facility is assigned to any Season records.
// If so, it will be illegal to delete the record (see PreDelete for more info)
func (f *Facility) IsFacilityInUse(db common.DatabaseProvider) (bool, error) {
	seasons, err := f.GetSeasonsForFacility(db)
	if err != nil {
		return false, err
	}
	return len(seasons) > 0, nil
}

// GetSeasonsForFacility gets all the Season records which have this Facility
// assigned. This is used for convenience purposes, e.g. in the UI to provide
// a link to navigate to a season from the single-Facility page.
func (f *Facility) GetSeasonsForFacility(db common.DatabaseProvider) ([]*Season, error) {
	// get all Seasons where Season.Facility == this FacilityId
	filter := func(c *Season) bool { return c.Facility == f.ID }
	seasons, err := common.GetAllWhere(db, &Season{}, filter)
	if err != nil {
		return nil, err
	}
	return seasons, nil
}
