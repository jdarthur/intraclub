package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"intraclub/api"
	"intraclub/database"

	"github.com/gin-gonic/gin"
)

type ScoreCountingType int

func (s ScoreCountingType) StaticallyValid() error {
	if s >= InvalidScoreCountingType {
		return fmt.Errorf("invalid score counting type: %d", s)
	}
	return nil
}

const (
	Point ScoreCountingType = iota
	Game
	Set
	InvalidScoreCountingType
)

func (s ScoreCountingType) String() string {
	switch s {
	case Point:
		return "point"
	case Game:
		return "game"
	case Set:
		return "set"
	default:
		return "invalid"
	}
}

// Label is the same as ScoreCountingType.String, except with
// capitalized values. This is used in the UI to populate a dropdown
func (s ScoreCountingType) Label() string {
	switch s {
	case Point:
		return "Point"
	case Game:
		return "Game"
	case Set:
		return "Set"
	default:
		return "InvalidScoreCountingType"
	}
}

func (s ScoreCountingType) Secondary() ScoreCountingType {
	switch s {
	case Game:
		return Point
	case Set:
		return Game
	default:
		return InvalidScoreCountingType
	}
}

var ScoreCountingTypes = []ScoreCountingType{Point, Game, Set}

func getScoreCountingTypes() []map[string]interface{} {
	output := make([]map[string]interface{}, 0)
	for _, scoreType := range ScoreCountingTypes {
		output = append(output, map[string]interface{}{
			"type": scoreType,
			"name": scoreType.Label(),
		})
	}
	return output
}

func GetScoreCountingTypes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{api.ResourceKey: getScoreCountingTypes()})
}

type ScoringStructureId database.RecordId

func (id *ScoringStructureId) UnmarshalJSON(bytes []byte) error {
	var rid database.RecordId
	if err := rid.UnmarshalJSON(bytes); err != nil {
		return err
	}
	*id = ScoringStructureId(rid)
	return nil
}

func (id ScoringStructureId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id ScoringStructureId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id ScoringStructureId) String() string {
	return id.RecordId().String()
}

type WinCondition struct {
	WinThreshold        int `json:"win_threshold"`
	MustWinBy           int `json:"must_win_by"`
	InstantWinThreshold int `json:"instant_win_threshold"`
}

func (w WinCondition) HasInstantWinThreshold() bool {
	return w.InstantWinThreshold > 0
}

func (w WinCondition) WinByTwoOrMore() bool {
	return w.MustWinBy > 1
}

func (w WinCondition) StaticallyValid() error {
	if w.WinThreshold < 1 {
		return errors.New("win threshold must be >= 1")
	}

	// disallow win-by-zero or e.g. win-by-negative-one
	if w.MustWinBy < 1 {
		return errors.New("must-win-by constraint must be >= 1")
	}

	// disallow e.g. first-to-one, win-by-two
	if w.WinThreshold < w.MustWinBy {
		return errors.New("win threshold cannot be lower than must-win-by constraint")
	}

	if w.HasInstantWinThreshold() {
		// if we have an instant win threshold, it must be at least
		// as large as the main win threshold. Doesn't make sense to
		// have e.g. an instant win at 3 if you don't "win" until you
		// reach 5 points, etc.
		if w.InstantWinThreshold < w.WinThreshold {
			return errors.New("instant-win-at threshold must be >= main win threshold")
		}
		// can't have the instant win threshold the same as the win threshold in e.g. win-by-two
		// constraint. In this situation the win-by-two constraint would be meaningless
		if w.InstantWinThreshold == w.WinThreshold && w.WinByTwoOrMore() {
			return fmt.Errorf("instant-win-at threshold cannot be the same as main win threshold in win-by-%d", w.MustWinBy)
		}
	}
	return nil
}

type ScoringStructureList []ScoringStructureId

func (s *ScoringStructureList) UnmarshalJSON(bytes []byte) error {
	s2 := make([]string, 0)
	err := json.Unmarshal(bytes, &s2)
	if err != nil {
		return err
	}
	for _, id := range s2 {
		recordId, err := database.RecordIdFromString(id)
		if err != nil {
			return err
		}
		*s = append(*s, ScoringStructureId(recordId))
	}
	return nil
}

type ScoringStructure struct {
	// ID is a unique ID for this scoring structure.
	// This can be referenced by composite scoring structures
	// or things like Schedule or PlayoffStructure objects
	// which need to reference a particular way that their
	// matches are played out from a scoring perspective
	ID ScoringStructureId `json:"id"`

	// Owner is the UserId who created this ScoringStructure.
	// This is only used to allow deletion / update and to
	// filter on one's own ScoringStructure records
	Owner database.UserId `json:"owner"`

	// Name is a descriptive name for this ScoringStructure
	Name string `json:"name"`

	// WinConditionCountingType is the ScoreCountingType
	// that determines when someone wins in this ScoringStructure.
	// The win condition will occur when a team in a Match
	// gets to a particular number of points, games, or sets
	// won, based on the configuration of this ScoringStructure
	WinConditionCountingType ScoreCountingType `json:"win_condition_counting_type"`

	// WinCondition sets out the thresholds at which a team wins
	// a Match using this ScoringStructure. This includes a main
	// win threshold, a possible must-win-by-X constraint, and a
	// threshold that a team instantly wins at, bypassing the
	// win-by-X constraint
	WinCondition WinCondition `json:"win_condition"`
}

// API/JSON shape decision: the former inline `secondary_scoring_structures`
// (`ScoringStructureList`) field has been removed from `ScoringStructure` and
// normalized into the `ScoringStructureSecondary` join table
// (scoring_structure_secondary collection). In-process reads reassemble the
// ordered references via `ScoringStructure.GetSecondaryScoringStructures`, and
// composite-ness is determined by `ScoringStructure.IsComposite`. This matches
// the join-table normalization of Format (format_rating / format_line) and
// Draft (draft_rating_cutoff).

func (c *ScoringStructure) UniquenessEquivalent(other *ScoringStructure) error {
	if c.Name == other.Name {
		return fmt.Errorf("scoring structure with name %s already exists", other.Name)
	}
	return nil
}

func (c *ScoringStructure) GetOwner() database.UserId {
	return c.Owner
}

func NewScoringStructure() *ScoringStructure {
	return &ScoringStructure{}
}

func (c *ScoringStructure) Type() string {
	return "scoring_structure"
}

func (c *ScoringStructure) GetId() database.RecordId {
	return c.ID.RecordId()
}

func (c *ScoringStructure) SetId(id database.RecordId) {
	c.ID = ScoringStructureId(id)
}

func (c *ScoringStructure) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return database.SysAdminAndUsers(c.Owner)
}

func (c *ScoringStructure) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (c *ScoringStructure) SetOwner(userId database.UserId) {
	c.Owner = userId
}

// GetSecondaryScoringStructures returns this scoring structure's secondary
// scoring structure references, ordered by ScoringStructureSecondary.SecondaryIndex.
func (c *ScoringStructure) GetSecondaryScoringStructures(ctx context.Context, db database.Provider) (ScoringStructureList, error) {
	rows, err := database.GetAllWhere[*ScoringStructureSecondary](ctx, db, func(_ context.Context, s *ScoringStructureSecondary) bool {
		return s.ScoringStructureId == c.ID
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(rows, func(a, b *ScoringStructureSecondary) int {
		return a.SecondaryIndex - b.SecondaryIndex
	})
	list := make(ScoringStructureList, 0, len(rows))
	for _, r := range rows {
		list = append(list, r.SecondaryScoringStructureId)
	}
	return list, nil
}

// SetSecondaryScoringStructures replaces this scoring structure's secondary
// references with the provided list, preserving order as SecondaryIndex values
// in the ScoringStructureSecondary join table. The scoring structure must
// already have an assigned ID (i.e. be created).
func (c *ScoringStructure) SetSecondaryScoringStructures(ctx context.Context, db database.Provider, list ScoringStructureList) error {
	existing, err := database.GetAllWhere[*ScoringStructureSecondary](ctx, db, func(_ context.Context, s *ScoringStructureSecondary) bool {
		return s.ScoringStructureId == c.ID
	})
	if err != nil {
		return err
	}
	for _, s := range existing {
		if _, _, err := database.DeleteOneById(ctx, db, s, s.ID); err != nil {
			return err
		}
	}
	for i, id := range list {
		if _, err := database.CreateOne(ctx, db, &ScoringStructureSecondary{
			ScoringStructureId:          c.ID,
			SecondaryScoringStructureId: id,
			SecondaryIndex:              i,
		}); err != nil {
			return err
		}
	}
	return nil
}

// IsComposite reports whether this scoring structure has any secondary scoring
// structures (i.e. it is a composite scoring structure).
func (c *ScoringStructure) IsComposite(ctx context.Context, db database.Provider) (bool, error) {
	list, err := c.GetSecondaryScoringStructures(ctx, db)
	if err != nil {
		return false, err
	}
	return len(list) > 0, nil
}

func (c *ScoringStructure) MaximumScoreCountingUnitsPlayed(secondaryCount int) (int, error) {
	if c.WinCondition.HasInstantWinThreshold() {
		// if we have an instant win at e.g. 3 sets, we can play at most (3 * 2) - 1 = 5 total sets
		return (c.WinCondition.InstantWinThreshold * 2) - 1, nil
	}

	normalWinThreshold := (c.WinCondition.WinThreshold * 2) - 1

	if c.WinCondition.WinByTwoOrMore() {
		if secondaryCount > 0 {
			return normalWinThreshold, fmt.Errorf("composite scoring structure does not support win-by-two-or-more constraint without instant win threshold")
		} else {
			return -1, nil
		}
	}
	return normalWinThreshold, nil
}

// StaticallyValid validates this scoring structure's scalar fields (win
// condition counting type and win condition). Composite-specific validation
// (which requires querying the secondary scoring structure relationships) is
// performed by DynamicallyValid.
func (c *ScoringStructure) StaticallyValid() error {
	// make sure the win condition counting type is legitimate
	err := c.WinConditionCountingType.StaticallyValid()
	if err != nil {
		return err
	}
	return c.WinCondition.StaticallyValid()
}

func (c *ScoringStructure) DynamicallyValid(ctx context.Context, db database.Provider) error {
	secondary, err := c.GetSecondaryScoringStructures(ctx, db)
	if err != nil {
		return err
	}

	// validate each secondary reference: it must exist and use the expected
	// score-counting type for a secondary of this structure
	for _, id := range secondary {
		secondaryStructure, err := database.GetExistingRecordById(ctx, db, &ScoringStructure{}, id.RecordId())
		if err != nil {
			return err
		}

		if secondaryStructure.WinConditionCountingType != c.WinConditionCountingType.Secondary() {
			return fmt.Errorf("cannot use %s-based secondary scoring structure in %s-based win condition", secondaryStructure.WinConditionCountingType, c.WinConditionCountingType)
		}
	}

	// composite-specific validation
	if len(secondary) > 0 {
		if c.WinConditionCountingType == Point {
			return fmt.Errorf("cannot use point-based win condition in a composite scoring structure")
		}

		maxUnits, err := c.MaximumScoreCountingUnitsPlayed(len(secondary))
		if err != nil {
			return err
		}

		if len(secondary) != maxUnits {
			// we must have the same length of secondary scoring structures as the max amount of
			// main score-counting units in the scoring win-condition scoring configuration. For
			// example, if we can play a max number of 3 sets in this scoring structure, we must
			// have a way to score all three of those sets using a ScoringStructure reference.
			return fmt.Errorf("secondary scoring structures length is %d, but we can play %d max %ss in this structure", len(secondary), maxUnits, c.WinConditionCountingType)
		}
	}

	return database.ExistsById(ctx, db, &User{}, c.Owner.RecordId())
}

func (c *ScoringStructure) WinningScore(myScore, yourScore int) bool {
	diff := myScore - yourScore

	// check against instant-win threshold if applicable
	if c.WinCondition.HasInstantWinThreshold() && myScore >= c.WinCondition.InstantWinThreshold {
		return true
	}

	// if we haven't hit the instant win threshold, check if we have hit the
	// main win threshold and cleared the win-by-X constraint
	if myScore >= c.WinCondition.WinThreshold && diff >= c.WinCondition.MustWinBy {
		return true
	}

	return false
}

// PostDelete cascades deletion to this scoring structure's
// scoring_structure_secondary join rows. Without this, deleting a scoring
// structure would orphan those rows (see #97).
func (c *ScoringStructure) PostDelete(ctx context.Context, db database.Provider) error {
	secondaries, err := database.GetAllWhere[*ScoringStructureSecondary](ctx, db, func(_ context.Context, s *ScoringStructureSecondary) bool {
		return s.ScoringStructureId == c.ID
	})
	if err != nil {
		return err
	}
	for _, s := range secondaries {
		if _, _, err := database.DeleteOneById(ctx, db, s, s.ID); err != nil {
			return err
		}
	}
	return nil
}

func (c *ScoringStructure) NewRecord() database.CrudRecord {
	return new(ScoringStructure)
}
