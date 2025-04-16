package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"net/http"
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
	c.JSON(http.StatusOK, gin.H{common.ResourceKey: getScoreCountingTypes()})
}

type ScoringStructureId common.RecordId

func (id ScoringStructureId) UnmarshalJSON(bytes []byte) error {
	return id.RecordId().UnmarshalJSON(bytes)
}

func (id ScoringStructureId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id ScoringStructureId) RecordId() common.RecordId {
	return common.RecordId(id)
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
	fmt.Println(string(bytes))
	err := json.Unmarshal(bytes, &s2)
	if err != nil {
		return err
	}
	for _, id := range s2 {
		recordId, err := common.RecordIdFromString(id)
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
	Owner UserId `json:"owner"`

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

	// SecondaryScoringStructures is a list of ScoringStructure
	// references that are used as a secondary mechanism to increment
	// the WinConditionCountingType. For example in a standard tennis
	// scoring structure, the primary win condition is winning 2 out
	// of 3 sets. But to win a _set_, you must first win a requisite
	// number of _games_, i.e. first to 6, win-by-two
	//
	// You could theoretically make the scoring even further nested
	// by specifying that _games_ must be won by winning a requisite
	// number of _points_, but this requires very active score-keeping
	// during a match and does not provide a lot of extra value for
	// the most part (i.e. a 6-0, 6-0 match does not gain much explanatory
	// context by recording that individual games were typically won
	// from a 40-15 or 40-love score)
	SecondaryScoringStructures ScoringStructureList `json:"secondary_scoring_structures"`
}

func (c *ScoringStructure) UniquenessEquivalent(other *ScoringStructure) error {
	if c.Name == other.Name {
		return fmt.Errorf("scoring structure with name %s already exists", other.Name)
	}
	return nil
}

func (c *ScoringStructure) GetOwner() common.RecordId {
	return c.Owner.RecordId()
}

func NewScoringStructure() *ScoringStructure {
	return &ScoringStructure{}
}

func (c *ScoringStructure) Type() string {
	return "scoring_structure"
}

func (c *ScoringStructure) GetId() common.RecordId {
	return c.ID.RecordId()
}

func (c *ScoringStructure) SetId(id common.RecordId) {
	c.ID = ScoringStructureId(id)
}

func (c *ScoringStructure) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return common.SysAdminAndUsers(c.Owner.RecordId())
}

func (c *ScoringStructure) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (c *ScoringStructure) SetOwner(recordId common.RecordId) {
	c.Owner = UserId(recordId)
}

func (c *ScoringStructure) MaximumScoreCountingUnitsPlayed() (int, error) {
	if c.WinCondition.HasInstantWinThreshold() {
		// if we have an instant win at e.g. 3 sets, we can play at most (3 * 2) - 1 = 5 total sets
		return (c.WinCondition.InstantWinThreshold * 2) - 1, nil
	}

	normalWinThreshold := (c.WinCondition.WinThreshold * 2) - 1

	if c.WinCondition.WinByTwoOrMore() {
		if c.IsComposite() {
			return normalWinThreshold, fmt.Errorf("composite scoring structure does not support win-by-two-or-more constraint without instant win threshold")
		} else {
			return -1, nil
		}
	}
	return normalWinThreshold, nil
}

func (c *ScoringStructure) IsComposite() bool {
	return len(c.SecondaryScoringStructures) > 0
}

func (c *ScoringStructure) StaticallyValid() error {
	// make sure the win condition counting type is legitimate
	err := c.WinConditionCountingType.StaticallyValid()
	if err != nil {
		return err
	}

	err = c.WinCondition.StaticallyValid()
	if err != nil {
		return err
	}

	if c.IsComposite() {

		if c.WinConditionCountingType == Point {
			return fmt.Errorf("cannot use point-based win condition in a composite scoring structure")
		}

		// get the maximum number of win-condition scoring units that we might play.
		// e.g. in a first-to-2 sets "standard tennis" scoring structure the max
		// amount of sets you can play is 3. In a "first to 10 points, straight up"
		// scoring structure, the maximum amount of total points would be in a 10 to 9
		// victory, so 19 total points.
		maxUnits, err := c.MaximumScoreCountingUnitsPlayed()
		if err != nil {
			return err
		}

		l := len(c.SecondaryScoringStructures)
		if l != maxUnits {

			// we must have the same length of secondary scoring structures as the max amount of
			// main score-counting units in the scoring win-condition scoring configuration. For
			// example, if we can play a max number of 3 sets in this scoring structure, we must
			// have a way to score all three of those sets using a ScoringStructure reference.
			return fmt.Errorf("secondary scoring structures length is %d, but we can play %d max %ss in this structure", l, maxUnits, c.WinConditionCountingType)
		}
	}
	return nil
}

func (c *ScoringStructure) DynamicallyValid(db common.DatabaseProvider) error {
	for _, id := range c.SecondaryScoringStructures {
		secondary, err := common.GetExistingRecordById(db, &ScoringStructure{}, id.RecordId())
		if err != nil {
			return err
		}

		if secondary.WinConditionCountingType != c.WinConditionCountingType.Secondary() {
			return fmt.Errorf("cannot use %s-based secondary scoring structure in %s-based win condition", secondary.WinConditionCountingType, c.WinConditionCountingType)
		}

	}
	return common.ExistsById(db, &User{}, c.Owner.RecordId())
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
