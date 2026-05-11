package model

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"intraclub/database"
)

// RulesetId is a unique identifier for the Ruleset type
type RulesetId database.RecordId

func (id RulesetId) UnmarshalJSON(bytes []byte) error {
	rid := id.RecordId()
	return (*database.RecordId)(&rid).UnmarshalJSON(bytes)
}

func (id RulesetId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id RulesetId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id RulesetId) String() string {
	return id.RecordId().String()
}

// Empty indicates that this RulesetId is unset / has an invalid value
func (id RulesetId) Empty() bool {
	return id.RecordId() == database.InvalidRecordId
}

type Ruleset struct {

	// ID is a unique RulesetId for this Ruleset
	ID RulesetId `json:"id"`

	// Name is an optional descriptive name for this Ruleset
	Name string `json:"name"`

	// Revision is the current revision of this Ruleset.
	// This will get incremented whenever we Amend or Fork
	// this Ruleset.
	Revision int `json:"revision"`

	// SupersededBy (if set to a non RulesetId.Empty value) is
	// the RulesetId of the Ruleset Revision that supersedes
	// this current Ruleset. If this value is non-empty, then
	// this Ruleset has been Archived.
	SupersededBy RulesetId `json:"superseded_by"`

	// Date is the time.Time when we either created or made a
	// RuleAmendment for this Ruleset
	Date time.Time `json:"date"`

	// Owner is the UserId who owns this Ruleset. Another User
	// cannot edit this Ruleset, but may always Fork it into a
	// new Ruleset if they would like to modify it for their
	// own Season, for example.
	Owner database.UserId `json:"owner"`
}

func (r *Ruleset) PreUpdate(ctx context.Context, db database.DatabaseProvider, existingValues database.CrudRecord) error {
	return errors.New("rulesets may not be directly modified (use Ruleset.Amend instead)")
}

// Archived is a bool which indicates this is not the most
// up-to-date version of this Ruleset. An Archived ruleset
// is no longer eligible for a RuleAmendment. Instead, to
// update an Archived Ruleset, we must Fork it.
func (r *Ruleset) Archived() bool {
	return !r.SupersededBy.Empty()
}

// NewRuleset creates a new *Ruleset with default values
func NewRuleset() *Ruleset {
	return &Ruleset{
		ID:           0,
		Name:         "",
		Revision:     0,
		SupersededBy: 0,
		Date:         time.Now(),
		Owner:        0,
	}
}

func (r *Ruleset) Type() string {
	return "ruleset"
}

func (r *Ruleset) GetId() database.RecordId {
	return r.ID.RecordId()
}

func (r *Ruleset) SetId(id database.RecordId) {
	r.ID = RulesetId(id)
}

func (r *Ruleset) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.UserId {
	return database.SysAdminAndUsers(r.Owner)
}

func (r *Ruleset) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.UserId {
	return database.AccessibleToEveryone
}

func (r *Ruleset) SetOwner(userId database.UserId) {
	r.Owner = userId
}

func (r *Ruleset) GetOwner() database.UserId {
	return r.Owner
}

func (r *Ruleset) StaticallyValid() error {
	if r.Name == "" {
		return fmt.Errorf("name must not be empty")
	}

	if r.Revision < 0 {
		// revision must be non-negative. Revision 0 is
		// the empty Ruleset before we add any RuleSections
		return fmt.Errorf("revision cannot be negative")
	}

	return nil
}

// CountSections returns the number of RuleSection records associated with this ruleset.
func (r *Ruleset) CountSections(ctx context.Context, db database.DatabaseProvider) (int, error) {
	sectionRelations, err := r.GetSectionRelations(ctx, db)
	if err != nil {
		return 0, err
	}
	return len(sectionRelations), nil
}

// GetSections returns all RuleSection IDs for this ruleset in order,
// based on the SectionIndex stored in the RulesetSection join table.
func (r *Ruleset) GetSections(ctx context.Context, db database.DatabaseProvider) ([]RuleSectionId, error) {
	sectionRelations, err := r.GetSectionRelations(ctx, db)
	if err != nil {
		return nil, err
	}
	result := make([]RuleSectionId, 0, len(sectionRelations))
	for _, sr := range sectionRelations {
		result = append(result, sr.SectionId)
	}
	return result, nil
}

// GetSectionRelations returns all RulesetSection join table records for this ruleset.
// The results include the SectionIndex which determines the ordering of sections.
func (r *Ruleset) GetSectionRelations(ctx context.Context, db database.DatabaseProvider) ([]*RulesetSection, error) {
	result, err := database.GetAllWhere[*RulesetSection](ctx, db, func(_ context.Context, rs *RulesetSection) bool {
		return rs.RulesetId == r.ID
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(result, func(a, b *RulesetSection) int {
		return a.SectionIndex - b.SectionIndex
	})
	return result, nil
}

func (r *Ruleset) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	// owner must exist
	err := database.ExistsById(ctx, db, &User{}, r.Owner.RecordId())
	if err != nil {
		return err
	}

	// each RuleSection must exist
	sectionRelations, err := r.GetSectionRelations(ctx, db)
	if err != nil {
		return err
	}

	for _, sr := range sectionRelations {
		err = database.ExistsById(ctx, db, &RuleSection{}, sr.SectionId.RecordId())
		if err != nil {
			return err
		}
	}

	// if this Ruleset is superseded by some other RuleSet
	if !r.SupersededBy.Empty() {
		other, err := database.GetExistingRecordById(ctx, db, &Ruleset{}, r.SupersededBy.RecordId())
		if err != nil {
			return err
		}

		if other.Owner != r.Owner {
			return fmt.Errorf("superseded-by owner %s does not match this ruleset's owner %s", other.Owner, r.Owner)
		}

		if other.Revision <= r.Revision {
			return fmt.Errorf("superseded-by record %s has revision %d (must be greater than this ruleset's revision %d)", r.SupersededBy, other.Revision, r.Revision)
		}

		if other.Date.Before(r.Date) {
			return fmt.Errorf("superseded-by record %s has date %s (must be after this ruleset's date %s)", r.SupersededBy, other.Date, r.Date)
		}
	}

	return nil
}

// AddSection creates a new RulesetSection join table record to add a section to this ruleset
// at the specified position. The index parameter determines the ordering of sections.
func (r *Ruleset) AddSection(ctx context.Context, db database.DatabaseProvider, sectionId RuleSectionId, index int) error {
	if err := database.ExistsById(ctx, db, &RuleSection{}, sectionId.RecordId()); err != nil {
		return err
	}

	sectionRelation := &RulesetSection{
		RulesetId:    r.ID,
		SectionId:    sectionId,
		SectionIndex: index,
	}
	_, err := database.CreateOne(ctx, db, sectionRelation)
	return err
}

// RemoveSection removes the RulesetSection join table records for a specific section from this ruleset.
func (r *Ruleset) RemoveSection(ctx context.Context, db database.DatabaseProvider, sectionId RuleSectionId) error {
	relations, err := database.GetAllWhere[*RulesetSection](ctx, db, func(_ context.Context, rs *RulesetSection) bool {
		return rs.RulesetId == r.ID && rs.SectionId == sectionId
	})
	if err != nil {
		return err
	}
	for _, rel := range relations {
		_, _, err = database.DeleteOneById(ctx, db, &RulesetSection{}, rel.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// Fork creates a new Ruleset that is a copy of this one, with a new owner.
// The new ruleset gets an incremented revision number and copies all sections
// from the original ruleset. Returns an error if the ruleset has no sections.
func (r *Ruleset) BlankRecord() database.CrudRecord {
	return new(Ruleset)
}

func (r *Ruleset) Fork(ctx context.Context, db database.DatabaseProvider, newUserId database.UserId) (*Ruleset, error) {
	sectionRelations, err := r.GetSectionRelations(ctx, db)
	if err != nil {
		return nil, err
	}
	if len(sectionRelations) == 0 {
		return nil, fmt.Errorf("cannot fork an empty ruleset")
	}

	r2 := NewRuleset()
	r2.ID = RulesetId(database.InvalidRecordId)
	r2.Name = r.Name
	r2.Revision = r.Revision + 1
	r2.SupersededBy = r.SupersededBy
	r2.Date = r.Date
	r2.Owner = newUserId

	newRuleset, err := database.CreateOne(ctx, db, r2)
	if err != nil {
		return nil, err
	}

	// Copy the section relations to the new ruleset
	for i, sr := range sectionRelations {
		newSectionRelation := &RulesetSection{
			RulesetId:    newRuleset.ID,
			SectionId:    sr.SectionId,
			SectionIndex: i,
		}
		_, err := database.CreateOne(ctx, db, newSectionRelation)
		if err != nil {
			return nil, err
		}
	}

	return newRuleset, nil
}

type RuleSectionId database.RecordId

func (id RuleSectionId) UnmarshalJSON(bytes []byte) error {
	rid := id.RecordId()
	return (*database.RecordId)(&rid).UnmarshalJSON(bytes)
}

func (id RuleSectionId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id RuleSectionId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id RuleSectionId) String() string {
	return id.RecordId().String()
}

func (id RuleSectionId) Empty() bool {
	return id.RecordId() == database.InvalidRecordId
}

type RuleSection struct {
	ID       RuleSectionId   `json:"section_id"`
	Parent   RulesetId       `json:"parent"`
	Title    string          `json:"title"`
	Markdown string          `json:"markdown"`
	Owner    database.UserId `json:"owner"`
}

func (section *RuleSection) Type() string {
	return "rule_section"
}

func (section *RuleSection) GetId() database.RecordId {
	return section.ID.RecordId()
}

func (section *RuleSection) SetId(id database.RecordId) {
	section.ID = RuleSectionId(id)
}

func (section *RuleSection) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.UserId {
	return database.SysAdminAndUsers(section.Owner)
}

func (section *RuleSection) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.UserId {
	return database.AccessibleToEveryone
}

func (section *RuleSection) SetOwner(userId database.UserId) {
	section.Owner = userId
}

func (section *RuleSection) GetOwner() database.UserId {
	return section.Owner
}

func (section *RuleSection) StaticallyValid() error {
	// trim whitespace characters from the ends of title
	section.Title = strings.TrimSpace(section.Title)

	// title is optional, so we don't need to validate that
	// there is a non-empty value in that field

	// trim whitespace characters from the edges of the contents
	section.Markdown = strings.TrimSpace(section.Markdown)

	// section contents cannot be empty as this would be useless
	if section.Markdown == "" {
		return fmt.Errorf("section contents are empty")
	}
	return nil
}

func (section *RuleSection) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	err := database.ExistsById(ctx, db, &Ruleset{}, section.Parent.RecordId())
	if err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &User{}, section.Owner.RecordId())
}

func (section *RuleSection) Empty() bool {
	if section.Parent.RecordId() != database.InvalidRecordId {
		return false
	}
	if !section.ID.Empty() {
		return false
	}
	if section.Title != "" {
		return false
	}
	if section.Markdown != "" {
		return false
	}
	return true
}

func (section *RuleSection) Equals(other *RuleSection) bool {
	if section.Title != other.Title {
		return false
	}
	if section.Markdown != other.Markdown {
		return false
	}
	return true
}

func (section *RuleSection) BlankRecord() database.CrudRecord {
	return new(RuleSection)
}
