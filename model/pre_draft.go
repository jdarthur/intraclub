package model

import (
	"fmt"
	"intraclub/common"
)

type PreDraftGrade struct {
	ID       common.RecordId // unique ID of this PreDraftGrade
	PlayerId UserId          // ID of the User who is being graded
	DraftId  DraftId         // ID of the Draft that this PreDraftGrade pertains to
	GraderId UserId          // ID of the User providing a Grade
	Rating   RatingId
}

func NewPreDraftGrade() *PreDraftGrade {
	return &PreDraftGrade{}
}

func (p *PreDraftGrade) Type() string {
	return "pre_draft_grade"
}

func (p *PreDraftGrade) GetId() common.RecordId {
	return p.ID
}

func (p *PreDraftGrade) SetId(id common.RecordId) {
	p.ID = id
}

func (p *PreDraftGrade) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{p.GraderId.RecordId()}
}

func (p *PreDraftGrade) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (p *PreDraftGrade) StaticallyValid() error {
	return nil
}

func (p *PreDraftGrade) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &User{}, p.PlayerId.RecordId())
	if err != nil {
		return err
	}

	err = common.ExistsById(db, &User{}, p.GraderId.RecordId())
	if err != nil {
		return err
	}

	draft, exists, err := common.GetOneById(db, &Draft{}, p.DraftId.RecordId())
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("draft %d does not exist", p.DraftId.RecordId())
	}

	format, exists, err := common.GetOneById(db, &Format{}, draft.Format.RecordId())
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("draft format %d does not exist", draft.Format.RecordId())
	}

	if !format.IsRatingValidForFormat(p.Rating) {
		return fmt.Errorf("rating %s is not a valid rating for draft's format %s", p.Rating, draft.Format.RecordId())
	}
	return nil
}
