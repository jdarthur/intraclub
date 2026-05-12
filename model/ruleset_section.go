package model

import (
	"context"
	"time"

	"intraclub/database"
)

// RulesetSection is a join table record that links a Ruleset to its RuleSections.
// The SectionIndex field maintains the ordering of sections within a ruleset.
type RulesetSection struct {
	ID           database.RecordId `json:"id"`
	RulesetId    RulesetId         `json:"ruleset_id"`
	SectionId    RuleSectionId     `json:"section_id"`
	SectionIndex int               `json:"section_index"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as RulesetSection has no specific owner.
func (r *RulesetSection) GetOwner() database.UserId {
	return database.InvalidUserId
}

// SetOwner is a no-op as RulesetSection has no specific owner.
func (r *RulesetSection) SetOwner(userId database.UserId) {}

// Type returns the record type identifier for RulesetSection.
func (r *RulesetSection) Type() string {
	return "ruleset_section"
}

// GetId returns the unique identifier for this RulesetSection record.
func (r *RulesetSection) GetId() database.RecordId {
	return r.ID
}

// SetId sets the unique identifier for this RulesetSection record.
func (r *RulesetSection) SetId(id database.RecordId) {
	r.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for RulesetSection.
func (r *RulesetSection) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that both the referenced Ruleset and RuleSection records exist.
func (r *RulesetSection) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Ruleset{}, r.RulesetId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &RuleSection{}, r.SectionId.RecordId()); err != nil {
		return err
	}
	return nil
}

func (r *RulesetSection) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (r *RulesetSection) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

// Timestamps returns the create and update timestamps for this RulesetSection record.
func (r *RulesetSection) Timestamps() (time.Time, time.Time, *time.Time) {
	return r.CreatedAt, r.UpdatedAt, nil
}

// SetCreatedAt sets the creation timestamp for this RulesetSection record.
func (r *RulesetSection) SetCreatedAt(createdAt time.Time) {
	r.CreatedAt = createdAt
}

// SetUpdatedAt sets the update timestamp for this RulesetSection record.
func (r *RulesetSection) SetUpdatedAt(updatedAt time.Time) {
	r.UpdatedAt = updatedAt
}

func (r *RulesetSection) NewRecord() database.CrudRecord {
	return new(RulesetSection)
}
