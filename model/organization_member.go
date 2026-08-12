package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"intraclub/database"
)

// OrganizationMember is a join table record that links a User to an
// Organization's evergreen membership roster. Membership is flat (no roles),
// and a single (Organization, User) pair may appear at most once.
type OrganizationMember struct {
	ID             database.RecordId `json:"id"`
	OrganizationId OrganizationId    `json:"organization_id"`
	UserId         database.UserId   `json:"user_id"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	DeletedAt      *time.Time        `json:"deleted_at"`
}

// GetOwner returns InvalidUserId as OrganizationMember has no specific owner.
func (m *OrganizationMember) GetOwner() database.UserId {
	return database.InvalidUserId
}

// SetOwner is a no-op as OrganizationMember has no specific owner.
func (m *OrganizationMember) SetOwner(userId database.UserId) {}

// AccessibleTo returns everyone as OrganizationMember records are public.
func (m *OrganizationMember) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (m *OrganizationMember) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	// Memberships may be edited by the organization's owner (or sysadmin).
	// Resolve the owning user from the referenced organization.
	org, exists, err := database.GetOneById(ctx, db, &Organization{}, m.OrganizationId.RecordId())
	if err != nil || !exists {
		return []database.UserId{database.SysAdminUserId}
	}
	return database.SysAdminAndUsers(org.UserId)
}

// Type returns the record type identifier for OrganizationMember.
func (m *OrganizationMember) Type() string {
	return "organization_member"
}

// GetId returns the unique identifier for this OrganizationMember record.
func (m *OrganizationMember) GetId() database.RecordId {
	return m.ID
}

// SetId sets the unique identifier for this OrganizationMember record.
func (m *OrganizationMember) SetId(id database.RecordId) {
	m.ID = id
}

// UniquenessEquivalent guards against a duplicate (organization, user)
// membership pair.
func (m *OrganizationMember) UniquenessEquivalent(other *OrganizationMember) error {
	if m.OrganizationId == other.OrganizationId && m.UserId == other.UserId {
		return fmt.Errorf("duplicate organization membership for organization %s user %s",
			m.OrganizationId, m.UserId)
	}
	return nil
}

// StaticallyValid validates the record's field-level constraints without
// requiring a database provider.
func (m *OrganizationMember) StaticallyValid() error {
	if m.OrganizationId == 0 {
		return errors.New("organization member organization is empty")
	}
	if m.UserId == 0 {
		return errors.New("organization member user is empty")
	}
	return nil
}

// DynamicallyValid verifies that both the referenced Organization and User
// records exist.
func (m *OrganizationMember) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Organization{}, m.OrganizationId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &User{}, m.UserId.RecordId()); err != nil {
		return err
	}
	return nil
}

// Timestamps returns the create, update and delete timestamps for this
// OrganizationMember record.
func (m *OrganizationMember) Timestamps() (time.Time, time.Time, *time.Time) {
	return m.CreatedAt, m.UpdatedAt, m.DeletedAt
}

func (m *OrganizationMember) SetCreatedAt(createdAt time.Time) {
	m.CreatedAt = createdAt
}

func (m *OrganizationMember) SetUpdatedAt(updatedAt time.Time) {
	m.UpdatedAt = updatedAt
}

func (m *OrganizationMember) SetDeletedAt(deletedAt *time.Time) {
	m.DeletedAt = deletedAt
}

func (m *OrganizationMember) NewRecord() database.CrudRecord {
	return new(OrganizationMember)
}
