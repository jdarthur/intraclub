package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"strings"
)

type FacilityId common.RecordId

func (id FacilityId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id FacilityId) String() string {
	return id.RecordId().String()
}

type Facility struct {
	ID             FacilityId
	UserId         UserId
	Name           string
	Address        string
	NumberOfCourts int
	LayoutPhoto    common.RecordId
}

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

func (f *Facility) SetOwner(recordId common.RecordId) {
	f.UserId = UserId(recordId)
}

func (f *Facility) EditableBy(common.DatabaseProvider) []common.RecordId {
	return common.SysAdminAndUsers(f.UserId.RecordId())
}

func (f *Facility) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func NewFacility() *Facility {
	return &Facility{}
}

func (f *Facility) Type() string {
	return "facility"
}

func (f *Facility) GetId() common.RecordId {
	return f.ID.RecordId()
}

func (f *Facility) SetId(id common.RecordId) {
	f.ID = FacilityId(id)
}

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

// Enforce that Facility.Name and Facility.Address fields are unique in the DB
func facilityAlreadyExistsWithValues(f, other *Facility) bool {
	if f.ID == other.ID {
		return false // don't enforce uniqueness against self
	}
	if f.Name == other.Name {
		return true
	}
	if f.Address == other.Address {
		return true
	}
	return false
}

func (f *Facility) DynamicallyValid(db common.DatabaseProvider, existing common.DatabaseValidatable) error {

	checkAgainst := f
	if existing != nil {
		checkAgainst = existing.(*Facility)
	}

	f2 := func(c *Facility) bool {
		return facilityAlreadyExistsWithValues(checkAgainst, c)
	}
	records, err := common.GetAllWhere(db, &Facility{}, f2)
	if err != nil {
		return err
	}

	if len(records) != 0 {
		if records[0].Name == f.Name {
			return errors.New("facility with provided name already exists")
		} else if records[0].Address == f.Address {
			return errors.New("facility with provided address already exists")
		}
		panic("unhandled facility value collision")
	}

	if f.LayoutPhoto != 0 {
		return common.ExistsById(db, &Photo{}, f.LayoutPhoto)
	}
	return nil
}

func (f *Facility) IsFacilityInUse(db common.DatabaseProvider) (bool, error) {
	seasons, err := f.GetSeasonsForFacility(db)
	if err != nil {
		return false, err
	}
	return len(seasons) > 0, nil
}

func (f *Facility) GetSeasonsForFacility(db common.DatabaseProvider) ([]*Season, error) {
	// get all Seasons where Season.Facility == this FacilityId
	filter := func(c *Season) bool { return c.Facility == f.ID }
	seasons, err := common.GetAllWhere(db, &Season{}, filter)
	if err != nil {
		return nil, err
	}
	return seasons, nil
}
