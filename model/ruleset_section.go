package model

import (
	"intraclub/common"
	"time"
)

// RulesetSection is a join table record that links a Ruleset to its RuleSections.
// The SectionIndex field maintains the ordering of sections within a ruleset.
type RulesetSection struct {
	ID           common.RecordId `json:"id"`
	RulesetId    RulesetId       `json:"ruleset_id"`
	SectionId    RuleSectionId   `json:"section_id"`
	SectionIndex int             `json:"section_index"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as RulesetSection has no specific owner.
func (r *RulesetSection) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

// SetOwner is a no-op as RulesetSection has no specific owner.
func (r *RulesetSection) SetOwner(recordId common.RecordId) {}

// Type returns the record type identifier for RulesetSection.
func (r *RulesetSection) Type() string {
	return "ruleset_section"
}

// GetId returns the unique identifier for this RulesetSection record.
func (r *RulesetSection) GetId() common.RecordId {
	return r.ID
}

// SetId sets the unique identifier for this RulesetSection record.
func (r *RulesetSection) SetId(id common.RecordId) {
	r.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for RulesetSection.
func (r *RulesetSection) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that both the referenced Ruleset and RuleSection records exist.
func (r *RulesetSection) DynamicallyValid(db common.DatabaseProvider) error {
	if err := common.ExistsById(db, &Ruleset{}, r.RulesetId.RecordId()); err != nil {
		return err
	}
	if err := common.ExistsById(db, &RuleSection{}, r.SectionId.RecordId()); err != nil {
		return err
	}
	return nil
}

// AccessibleTo returns everyone as RulesetSection records are public.
func (r *RulesetSection) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

// EditableBy returns only sysadmins as only they can modify RulesetSection records.
func (r *RulesetSection) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.SysAdminRecordId}
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
