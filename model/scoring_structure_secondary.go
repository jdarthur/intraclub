package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

// ScoringStructureSecondary is a join-table record that links a (composite)
// ScoringStructure to one of its secondary ScoringStructure references. It is
// the fully normalized replacement for the former inline
// `ScoringStructure.SecondaryScoringStructures` slice (see model/scoring_structure.go),
// following the same join-table pattern as FormatRating, FormatLine, etc.
//
// SecondaryIndex preserves the ordering of the secondary scoring structures,
// which maps positionally onto the maximum number of playable win-condition
// units (see ScoringStructure.StaticallyValid/DynamicallyValid). A natural
// unique constraint on (ScoringStructureId, SecondaryIndex) prevents two
// secondaries from sharing a position.
type ScoringStructureSecondary struct {
	ID                          database.RecordId  `json:"id"`
	ScoringStructureId          ScoringStructureId `json:"scoring_structure_id"`
	SecondaryScoringStructureId ScoringStructureId `json:"secondary_scoring_structure_id"`
	SecondaryIndex              int                `json:"secondary_index"`
}

func (s *ScoringStructureSecondary) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (s *ScoringStructureSecondary) SetOwner(userId database.UserId) {}

func (s *ScoringStructureSecondary) Type() string {
	return "scoring_structure_secondary"
}

func (s *ScoringStructureSecondary) GetId() database.RecordId {
	return s.ID
}

func (s *ScoringStructureSecondary) SetId(id database.RecordId) {
	s.ID = id
}

func (s *ScoringStructureSecondary) StaticallyValid() error {
	return nil
}

func (s *ScoringStructureSecondary) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &ScoringStructure{}, s.ScoringStructureId.RecordId()); err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &ScoringStructure{}, s.SecondaryScoringStructureId.RecordId())
}

func (s *ScoringStructureSecondary) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (s *ScoringStructureSecondary) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (s *ScoringStructureSecondary) NewRecord() database.CrudRecord {
	return new(ScoringStructureSecondary)
}

// UniquenessEquivalent enforces the natural unique constraint on
// (ScoringStructureId, SecondaryIndex): a scoring structure may only have one
// secondary at each position.
func (s *ScoringStructureSecondary) UniquenessEquivalent(other *ScoringStructureSecondary) error {
	if s.ScoringStructureId == other.ScoringStructureId && s.SecondaryIndex == other.SecondaryIndex {
		return fmt.Errorf("scoring structure %s already has a secondary at index %d", s.ScoringStructureId, s.SecondaryIndex)
	}
	return nil
}
