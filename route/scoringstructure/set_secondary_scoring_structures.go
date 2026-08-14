package scoringstructure

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for ScoringStructure records.
var BaseRoute = "/scoring_structure"

// SetSecondaryScoringStructuresBody is the request body for setting a
// ScoringStructure's secondary (tie-breaker) scoring structures. The
// `secondary_scoring_structures` list replaces the structure's entire set of
// assigned secondary references (as ScoringStructureSecondary join records),
// preserving order as SecondaryIndex values. An empty list clears the
// structure's secondaries (making it non-composite).
type SetSecondaryScoringStructuresBody struct {
	SecondaryScoringStructures model.ScoringStructureList `json:"secondary_scoring_structures"`
}

// StaticallyValid is a no-op: an empty list is legitimate (a non-composite
// scoring structure). Count and counting-type validity are checked by the
// handler and by the primary structure's DynamicallyValid on save.
func (b *SetSecondaryScoringStructuresBody) StaticallyValid() error {
	return nil
}

// GetSecondaryScoringStructures returns a ScoringStructure's currently-assigned
// secondary scoring structures, ordered by SecondaryIndex (tie-breaker
// position).
type GetSecondaryScoringStructures struct{}

func (c GetSecondaryScoringStructures) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, api.AppendPathId(BaseRoute) + "/secondary_scoring_structures"
}

func (c GetSecondaryScoringStructures) RequestBody() (*SetSecondaryScoringStructuresBody, bool) {
	return &SetSecondaryScoringStructuresBody{}, false
}

func (c GetSecondaryScoringStructures) Handler(req api.Request[*SetSecondaryScoringStructuresBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required to get secondary scoring structures")
	}

	structure, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.ScoringStructure{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	secondaries, err := resolveSecondaryScoringStructures(req.Context, req.DatabaseProvider, structure)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: secondaries}, http.StatusOK, nil
}

// SetSecondaryScoringStructures replaces a ScoringStructure's secondary
// (tie-breaker) scoring structure references with the provided list,
// preserving order. Only the structure's owner (or a SysAdmin) may do so.
type SetSecondaryScoringStructures struct{}

func (c SetSecondaryScoringStructures) Path() (api.HttpMethod, string) {
	return api.HttpMethodPut, api.AppendPathId(BaseRoute) + "/secondary_scoring_structures"
}

func (c SetSecondaryScoringStructures) RequestBody() (*SetSecondaryScoringStructuresBody, bool) {
	return &SetSecondaryScoringStructuresBody{}, true
}

func (c SetSecondaryScoringStructures) Handler(req api.Request[*SetSecondaryScoringStructuresBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required to set secondary scoring structures")
	}

	structure, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.ScoringStructure{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// enforce per-record edit authorization (owner / sysadmin)
	wac := database.NewWithAccessControl[*model.ScoringStructure](req.Context, req.DatabaseProvider, req.Token.UserId)
	if !wac.CanUserEdit(structure) {
		return nil, http.StatusForbidden, errors.New("not authorized to set secondary scoring structures for this scoring structure")
	}

	if err := validateSecondaryCountingTypes(req.Context, req.DatabaseProvider, structure, req.Body.SecondaryScoringStructures); err != nil {
		return nil, http.StatusBadRequest, err
	}
	if err := validateSecondaryCount(structure, req.Body.SecondaryScoringStructures); err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err := structure.SetSecondaryScoringStructures(req.Context, req.DatabaseProvider, req.Body.SecondaryScoringStructures); err != nil {
		return nil, http.StatusBadRequest, err
	}

	secondaries, err := resolveSecondaryScoringStructures(req.Context, req.DatabaseProvider, structure)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return gin.H{api.ResourceKey: secondaries}, http.StatusOK, nil
}

// validateSecondaryCountingTypes checks that every secondary reference uses
// the score-counting type expected for a secondary of this structure (the same
// rule ScoringStructure.DynamicallyValid enforces when the primary is saved),
// so mismatches are rejected here instead of blocking a later primary save.
func validateSecondaryCountingTypes(ctx context.Context, db database.Provider, structure *model.ScoringStructure, ids model.ScoringStructureList) error {
	for _, id := range ids {
		secondary, err := database.GetExistingRecordById(ctx, db, &model.ScoringStructure{}, id.RecordId())
		if err != nil {
			return err
		}
		if secondary.WinConditionCountingType != structure.WinConditionCountingType.Secondary() {
			return fmt.Errorf("cannot use %s-based secondary scoring structure in %s-based win condition", secondary.WinConditionCountingType, structure.WinConditionCountingType)
		}
	}
	return nil
}

// validateSecondaryCount enforces the composite count rule that
// ScoringStructure.DynamicallyValid applies when the primary is saved: a
// composite structure must have exactly as many secondaries as the maximum
// number of score-counting units playable under its win condition. An empty
// list (a non-composite structure) is always valid. Rejecting a wrong count
// here prevents persisting an invalid composite via this endpoint that would
// only fail later when the primary is saved.
func validateSecondaryCount(structure *model.ScoringStructure, ids model.ScoringStructureList) error {
	if len(ids) == 0 {
		return nil
	}
	if structure.WinConditionCountingType == model.Point {
		return fmt.Errorf("cannot use point-based win condition in a composite scoring structure")
	}
	maxUnits, err := structure.MaximumScoreCountingUnitsPlayed(len(ids))
	if err != nil {
		return err
	}
	if len(ids) != maxUnits {
		return fmt.Errorf("secondary scoring structures length is %d, but we can play %d max %ss in this structure", len(ids), maxUnits, structure.WinConditionCountingType)
	}
	return nil
}

// resolveSecondaryScoringStructures fetches the full ScoringStructure records
// for a structure's secondary references, preserving SecondaryIndex order.
func resolveSecondaryScoringStructures(ctx context.Context, db database.Provider, structure *model.ScoringStructure) ([]model.ScoringStructure, error) {
	ids, err := structure.GetSecondaryScoringStructures(ctx, db)
	if err != nil {
		return nil, err
	}

	secondaries := make([]model.ScoringStructure, 0, len(ids))
	for _, id := range ids {
		s, err := database.GetExistingRecordById(ctx, db, &model.ScoringStructure{}, id.RecordId())
		if err != nil {
			return nil, err
		}
		secondaries = append(secondaries, *s)
	}
	return secondaries, nil
}
