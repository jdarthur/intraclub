package model

import (
	"context"
	"fmt"
	"intraclub/common"
	"strings"
	"time"
)

func getRuleAmendmentContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = cancel
	return ctx
}

type RuleAmendmentType int

const (
	RuleAmendmentTypeAddSection RuleAmendmentType = iota
	RuleAmendmentTypeRemoveSection
	RuleAmendmentTypeModifySection
	RuleAmendmentTypeReorderSection
	RuleAmendmentTypeInvalid
)

func (r RuleAmendmentType) StaticallyValid() error {
	if r < 0 || r >= RuleAmendmentTypeInvalid {
		return fmt.Errorf("invalid rule amendment type: %d", r)
	}
	return nil
}

func (r *Ruleset) Amend(ctx context.Context, db common.DatabaseProvider, a *RuleAmendment) (new *Ruleset, err error) {
	err = r.ValidateAmendment(ctx, db, a)
	if err != nil {
		return nil, err
	}

	if a.Type == RuleAmendmentTypeAddSection {
		return r.HandleAddSection(getRuleAmendmentContext(), db, a)
	} else if a.Type == RuleAmendmentTypeRemoveSection {
		return r.HandleRemoveSection(getRuleAmendmentContext(), db, a)
	} else if a.Type == RuleAmendmentTypeModifySection {
		return r.HandleModifySection(getRuleAmendmentContext(), db, a)
	} else if a.Type == RuleAmendmentTypeReorderSection {
		return r.HandleReorderSection(getRuleAmendmentContext(), db, a)
	}

	return nil, fmt.Errorf("unhandled rule amendment type: %d", a.Type)
}

func (r *Ruleset) HandleAddSection(ctx context.Context, db common.DatabaseProvider, a *RuleAmendment) (new *Ruleset, err error) {
	// update the parent and the owner in the new section to add
	a.NewSection.Parent = r.ID
	a.NewSection.Owner = r.Owner

	// create a new section
	v, err := common.CreateOne(ctx, db, &a.NewSection)
	if err != nil {
		return nil, err
	}

	// Get existing section relations
	existingRelations, err := r.GetSectionRelations(ctx, db)
	if err != nil {
		return nil, err
	}

	// Build new section list based on After field
	var newRelations []*RulesetSection
	if a.After.Empty() {
		// add new section to front
		newRelations = append(newRelations, &RulesetSection{
			RulesetId:    r.ID,
			SectionId:    v.ID,
			SectionIndex: 0,
		})
		for i, sr := range existingRelations {
			srCopy := *sr
			srCopy.SectionIndex = i + 1
			newRelations = append(newRelations, &srCopy)
		}
	} else {
		// add existing records, putting the new record in place after the target
		targetIndex := -1
		for i, sr := range existingRelations {
			if sr.SectionId == a.After {
				targetIndex = i
				break
			}
		}
		if targetIndex == -1 {
			return nil, fmt.Errorf("after section %s not found", a.After)
		}

		// Add sections before target
		for i := 0; i <= targetIndex; i++ {
			newRelations = append(newRelations, existingRelations[i])
		}

		// Add new section
		newSectionRelation := &RulesetSection{
			RulesetId:    r.ID,
			SectionId:    v.ID,
			SectionIndex: targetIndex + 1,
		}
		_, err = common.CreateOne(ctx, db, newSectionRelation)
		if err != nil {
			return nil, err
		}
		newRelations = append(newRelations, newSectionRelation)

		// Add sections after target
		for i := targetIndex + 1; i < len(existingRelations); i++ {
			newRelations = append(newRelations, existingRelations[i])
		}
	}

	// Handle the amendment - update section relations
	newRuleset, err := r.HandleAmendment(getRuleAmendmentContext(), db, newRelations)
	if err != nil {
		fmt.Printf("Error handling amendment: %s\n", err)
		fmt.Printf("Deleting newly-created section after failed amendment %s\n", v.ID)
		_, _, err = common.DeleteOneById(ctx, db, &RuleSection{}, v.ID.RecordId())
		if err != nil {
			return nil, fmt.Errorf("error deleting newly-created section %s after amendment: %s", v.ID, err.Error())
		}
	}
	return newRuleset, nil
}

func (r *Ruleset) HandleRemoveSection(ctx context.Context, db common.DatabaseProvider, a *RuleAmendment) (new *Ruleset, err error) {
	// Get existing section relations
	existingRelations, err := r.GetSectionRelations(ctx, db)
	if err != nil {
		return nil, err
	}

	// Build new section list excluding target
	var newRelations []*RulesetSection
	for _, sr := range existingRelations {
		if sr.SectionId != a.TargetSection {
			newRelations = append(newRelations, sr)
		}
	}

	if len(newRelations) == 0 {
		return nil, fmt.Errorf("deleting section would make ruleset empty")
	}

	// Delete the target section relation
	_, _, err = common.DeleteOneById(ctx, db, &RulesetSection{}, a.TargetSection.RecordId())
	if err != nil {
		return nil, err
	}

	// Handle the amendment
	newRuleset, err := r.HandleAmendment(getRuleAmendmentContext(), db, newRelations)
	return newRuleset, err
}

func (r *Ruleset) HandleModifySection(ctx context.Context, db common.DatabaseProvider, a *RuleAmendment) (new *Ruleset, err error) {

	existing, err := common.GetExistingRecordById(ctx, db, &RuleSection{}, a.TargetSection.RecordId())
	if err != nil {
		return nil, err
	}

	// check if any values were actually changed
	if existing.Equals(&a.NewSection) {
		return nil, fmt.Errorf("new section is not updated")
	}

	if existing.Markdown == a.NewSection.Markdown && existing.Title != a.NewSection.Title {
		// if the title is changed but the markdown content is not changed, we do not
		// need to trigger a new revision. Not a meaningful rule change at that point
		// so there is not a real reason to track it in the revision history. Instead, just
		// update the existing RuleSection and return any errors we encounter during this
		existing.Title = a.NewSection.Title
		err = db.Update(getRuleAmendmentContext(), existing)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// Update the RuleSection content
	existing.Title = a.NewSection.Title
	existing.Markdown = a.NewSection.Markdown
	err = db.Update(getRuleAmendmentContext(), existing)
	if err != nil {
		return nil, err
	}

	// No revision needed for modify without order change, just update the section
	return r, nil
}

func (r *Ruleset) HandleReorderSection(ctx context.Context, db common.DatabaseProvider, a *RuleAmendment) (new *Ruleset, err error) {
	// Get existing section relations
	existingRelations, err := r.GetSectionRelations(ctx, db)
	if err != nil {
		return nil, err
	}

	// Find the target relation to reuse (preserves the existing ID)
	var targetRelation *RulesetSection
	var otherRelations []*RulesetSection
	for _, sr := range existingRelations {
		if sr.SectionId == a.TargetSection {
			targetRelation = sr
		} else {
			otherRelations = append(otherRelations, sr)
		}
	}

	// Build new ordered list by inserting target after the "After" section
	var newRelations []*RulesetSection
	found := false
	for _, sr := range otherRelations {
		newRelations = append(newRelations, sr)
		if sr.SectionId == a.After {
			newRelations = append(newRelations, targetRelation)
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("after section ID %s was not found in ruleset %s", a.After, r.ID)
	}

	// Update section indices for all relations
	for i, sr := range newRelations {
		srCopy := *sr
		srCopy.SectionIndex = i
		err = db.Update(getRuleAmendmentContext(), &srCopy)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Ruleset) HandleAmendment(ctx context.Context, db common.DatabaseProvider, newRelations []*RulesetSection) (newRuleset *Ruleset, err error) {
	// Build section ID list from relations
	newSectionIds := make([]RuleSectionId, 0, len(newRelations))
	for _, sr := range newRelations {
		newSectionIds = append(newSectionIds, sr.SectionId)
	}

	// copy the existing ruleset to a new object with the target sections list.
	// the Copy operation here will reset the timestamp to the current time
	copied := r.Copy(newSectionIds)

	// increment the revision number by one
	copied.Revision += 1

	// create a new ruleset with the updated values
	newRuleset, err = common.CreateOne(ctx, db, copied)
	if err != nil {
		return nil, err
	}

	// Mark this ruleset as superseded by the new ruleset
	r.SupersededBy = newRuleset.ID
	err = db.Update(getRuleAmendmentContext(), r)

	if err != nil {
		// Unusual error state: we have successfully created a superseding
		// Ruleset to this current Ruleset, but we were unable to update the
		// current record to indicate it is superseded by another Ruleset.
		// In this situation we will revert the creation of the new Ruleset
		// and return the error from the update call to the end user
		fmt.Printf("Error updating ruleset after being superseded: %s\n", err)
		fmt.Printf("Deleting newly-created superseding ruleset %s\n", r.ID)

		// If we failed to update this current ruleset, we need to revert the
		// creation of the superseding ruleset
		_, _, err = common.DeleteOneById(ctx, db, &Ruleset{}, newRuleset.ID.RecordId())
		if err != nil {
			return nil, fmt.Errorf("error deleting superseding ruleset %s: %s", newRuleset.ID, err)
		}

		// Return the error from the failed update
		return nil, err
	}

	// Delete all old section relations
	oldRelations, _ := r.GetSectionRelations(ctx, db)
	for _, oldRel := range oldRelations {
		if oldRel.RulesetId == r.ID {
			_, _, _ = common.DeleteOneById(ctx, db, &RulesetSection{}, oldRel.ID)
		}
	}

	// Add all new section relations to the new ruleset
	for i, sr := range newRelations {
		newRuleSectionRelation := &RulesetSection{
			RulesetId:    newRuleset.ID,
			SectionId:    sr.SectionId,
			SectionIndex: i,
		}
		_, err = common.CreateOne(ctx, db, newRuleSectionRelation)
		if err != nil {
			return nil, err
		}
	}

	// Return the new ruleset that supersedes this one
	return newRuleset, nil
}

func (r *Ruleset) Copy(newSectionIds []RuleSectionId) *Ruleset {
	sectionRelations := make([]*RulesetSection, 0, len(newSectionIds))
	for i, sectionId := range newSectionIds {
		sectionRelations = append(sectionRelations, &RulesetSection{
			RulesetId:    r.ID,
			SectionId:    sectionId,
			SectionIndex: i,
		})
	}

	return &Ruleset{
		Name:         r.Name,
		Revision:     r.Revision,
		SupersededBy: r.SupersededBy,
		Date:         time.Now(),
		Owner:        r.Owner,
	}
}

type RuleAmendment struct {
	Type          RuleAmendmentType `json:"type"`
	TargetSection RuleSectionId     `json:"target_section"`
	NewSection    RuleSection       `json:"new_section"`
	After         RuleSectionId     `json:"after"`
}

func (r *RuleAmendment) DynamicallyValid(ctx context.Context, db common.DatabaseProvider) error {
	if r.Type == RuleAmendmentTypeAddSection {
		if r.After.Empty() {
			// if After section ID is empty, we are adding this new section to
			// the beginning of the Ruleset's sections list
			return nil
		}
		return common.ExistsById(ctx, db, &RuleSection{}, r.After.RecordId())
	} else if r.Type == RuleAmendmentTypeRemoveSection {
		return common.ExistsById(ctx, db, &RuleSection{}, r.TargetSection.RecordId())
	} else if r.Type == RuleAmendmentTypeModifySection {
		return common.ExistsById(ctx, db, &RuleSection{}, r.TargetSection.RecordId())
	} else if r.Type == RuleAmendmentTypeReorderSection {

		if r.After.Empty() {
			// setting r.After to the empty value moves this section to the
			// front of the Ruleset's sections list
			return nil
		} else {
			err := common.ExistsById(ctx, db, &RuleSection{}, r.After.RecordId())
			if err != nil {
				return err
			}
		}

		return common.ExistsById(ctx, db, &RuleSection{}, r.TargetSection.RecordId())
	}
	return nil
}

func (r *RuleAmendment) StaticallyValid() error {
	err := r.Type.StaticallyValid()
	if err != nil {
		return err
	}

	r.NewSection.Title = strings.TrimSpace(r.NewSection.Title)
	r.NewSection.Markdown = strings.TrimSpace(r.NewSection.Markdown)

	if r.Type == RuleAmendmentTypeAddSection {
		if !r.TargetSection.Empty() {
			// cannot supply a target section ID when adding a new section
			return fmt.Errorf("'target section' must be empty when adding a new section")
		}
		if !r.NewSection.Parent.Empty() {
			// user cannot set parent ID in the new section contents (set automatically on amendment)
			return fmt.Errorf("'new section' parent must not be set when adding a new section")
		}
		if !r.NewSection.ID.Empty() {
			// user cannot set rule section ID in the new section contents (set automatically on amendment)
			return fmt.Errorf("'new section' ID must not be set when adding a new section")
		}
		if r.NewSection.Markdown == "" {
			// section contents cannot be empty
			return fmt.Errorf("'new section' contents must not be empty when adding a new section")
		}
	}

	if r.Type == RuleAmendmentTypeRemoveSection {
		if !r.NewSection.Empty() {
			return fmt.Errorf("'new section' contents must be empty when removing a section")
		}
		if !r.After.Empty() {
			return fmt.Errorf("'after section ID' must be empty when removing a section")
		}
	}

	if r.Type == RuleAmendmentTypeModifySection {
		if r.TargetSection.Empty() {
			return fmt.Errorf("'target section' must not be empty when updating a section")
		}
		if r.NewSection.Markdown == "" {
			return fmt.Errorf("'new section' contents must not be empty when updating a section")
		}
		if !r.After.Empty() {
			return fmt.Errorf("'after section ID' must be empty when updating a section")
		}
	}

	if r.Type == RuleAmendmentTypeReorderSection {
		if r.TargetSection == r.After {
			return fmt.Errorf("'target section' must be different than after section")
		}

		if !r.NewSection.Empty() {
			return fmt.Errorf("'new section' contents must be empty when moving a section within a ruleset")
		}
	}

	return nil
}

func (r *Ruleset) GetAmendmentBefore(ctx context.Context, db common.DatabaseProvider, id RuleSectionId) (RuleSectionId, error) {
	sectionRelations, err := r.GetSectionRelations(ctx, db)
	if err != nil {
		return 0, err
	}

	for i, sr := range sectionRelations {
		if sr.SectionId == id {
			if i == 0 {
				return RuleSectionId(common.InvalidRecordId), nil
			} else {
				return sectionRelations[i-1].SectionId, nil
			}
		}
	}
	return 0, fmt.Errorf("rule section ID %s was not found in ruleset %s", id, r.ID)
}

func (r *Ruleset) ValidateAmendment(ctx context.Context, db common.DatabaseProvider, a *RuleAmendment) error {
	err := common.Validate(ctx, db, a)
	if err != nil {
		return err
	}

	if a.Type == RuleAmendmentTypeAddSection {
		// validate that the "after rule X" ID exists in the ruleset
		if a.After.Empty() {
			// if After is empty, we'll add this amendment to the front
			// of the Ruleset's sections list
			return nil
		}
		_, err = r.GetAmendmentBefore(ctx, db, a.After)
		if err != nil {
			return err
		}
	} else if a.Type == RuleAmendmentTypeRemoveSection {
		// validate that the target rule ID exists in the ruleset
		_, err = r.GetAmendmentBefore(ctx, db, a.TargetSection)
		if err != nil {
			return err
		}
	} else if a.Type == RuleAmendmentTypeModifySection {
		// validate that the "after rule X" ID exists in the ruleset
		_, err = r.GetAmendmentBefore(ctx, db, a.TargetSection)
		if err != nil {
			return err
		}
	} else if a.Type == RuleAmendmentTypeReorderSection {
		other, err := r.GetAmendmentBefore(ctx, db, a.TargetSection)
		if err != nil {
			// error if the after-rule-X ID doesn't exist in the ruleset
			return err
		}
		if a.After == other {
			return fmt.Errorf("target section %s remains after section %s (does not reorder)", a.TargetSection, a.After)
		}
	}
	return nil
}
