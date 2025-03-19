package model

import (
	"fmt"
	"intraclub/common"
)

type ScoreCountingType int

func (s ScoreCountingType) StaticallyValid() error {
	if s >= Invalid {
		return fmt.Errorf("invalid score counting type: %d", s)
	}
	return nil
}

const (
	Point ScoreCountingType = iota
	Game
	Set
	NotApplicable
	Invalid
)

func (s ScoreCountingType) String() string {
	switch s {
	case Point:
		return "Point"
	case Game:
		return "Game"
	case Set:
		return "Set"
	case NotApplicable:
		return "NotApplicable"
	default:
		return "Invalid"
	}
}

type ScoringStructureId common.RecordId

func (id ScoringStructureId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id ScoringStructureId) String() string {
	return id.RecordId().String()
}

type ScoringStructure struct {
	// ID is a unique identifier for this ScoringStructure
	ID ScoringStructureId

	Owner UserId

	// WinConditionCountingType is the ScoreCountingType that determines who wins
	// in this ScoringStructure.
	WinConditionCountingType ScoreCountingType

	// MainScoreWinsAt determines where a team must get to in the WinConditionCountingType
	// in order to win in this ScoringStructure. They must also satisfy the MainScoreMustWinBy
	// threshold in order to reach the win condition.
	MainScoreWinsAt int

	// MainScoreMustWinBy determines the value that a team must beat the other team by in
	// order to trigger the win condition, e.g. a win-by-two constraint
	MainScoreMustWinBy int

	// A team wins automatically if they reach this number, for example to short-circuit
	// a win-by-two constraint for sudden-death purposes
	MainScoreInstantWinAt int

	// SecondaryScoreCountingType is the ScoreCountingType that is used to increment
	// the WinConditionCountingType, if applicable. For example, you may trigger the
	// win condition if you win 3 games to 11 points, or win 2 sets each played to 6 games
	SecondaryScoreCountingType ScoreCountingType

	// SecondaryScoreWinsAt is the threshold that a team must reach in order
	// to increment the main ScoreCountingType (as long as they also satisfy
	// the SecondaryScoreMustWinBy constraint)
	SecondaryScoreWinsAt int

	// SecondaryScoreMustWinBy is a constraint that delays the win condition for the
	// SecondaryScoreCountingType until a team has X amount of that type compared to the
	// other team. For example, scoring might be played to 11, but with a win-by-two constraint
	SecondaryScoreMustWinBy int

	// SecondaryScoreInstantWinAt is a threshold that, when reached, causes a team to instantly
	// reach the SecondaryScoreCountingType win condition. For example, this can be used to e.g,
	// disregard a "first-to-seven, win-by-two" constraint when either team hits 10
	SecondaryScoreInstantWinAt int
}

func (s *ScoringStructure) Type() string {
	return "scoring_structure"
}

func (s *ScoringStructure) GetId() common.RecordId {
	return s.ID.RecordId()
}

func (s *ScoringStructure) SetId(id common.RecordId) {
	s.ID = ScoringStructureId(id)
}

func (s *ScoringStructure) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return common.SysAdminAndUsers(s.Owner.RecordId())
}

func (s *ScoringStructure) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (s *ScoringStructure) SetOwner(recordId common.RecordId) {
	s.Owner = UserId(recordId)
}

func (s *ScoringStructure) StaticallyValid() error {
	err := s.WinConditionCountingType.StaticallyValid()
	if err != nil {
		return err
	}
	err = s.SecondaryScoreCountingType.StaticallyValid()
	if err != nil {
		return err
	}

	err = s.validateWinIntegers(s.MainScoreWinsAt, s.MainScoreMustWinBy, s.MainScoreInstantWinAt, true)
	if err != nil {
		return err
	}

	if s.IsComposite() {
		return s.validateWinIntegers(s.SecondaryScoreWinsAt, s.SecondaryScoreMustWinBy, s.SecondaryScoreInstantWinAt, false)
	}
	return nil
}

func (s *ScoringStructure) validateWinIntegers(winsAt, winBy, instantWin int, isMainScore bool) error {
	descriptor := "main"
	if !isMainScore {
		descriptor = "secondary"
	}

	if instantWin > 0 && instantWin < winsAt {
		return fmt.Errorf("%s instant win (%d) is less than %s wins-at (%d)", descriptor, instantWin, descriptor, winsAt)
	}
	if winsAt <= 0 {
		return fmt.Errorf("%s wins-at must be > 0 (got %d)", descriptor, winsAt)
	}
	if winBy <= 0 {
		return fmt.Errorf("%s win-by value must be > 0 (got %d)", descriptor, winBy)
	}
	if instantWin == winsAt && winBy > 1 {
		return fmt.Errorf("%s instant win cannot be the same as %s wins-at in win-by-%d", descriptor, descriptor, winBy)
	}
	return nil
}

func (s *ScoringStructure) WinningScore(myScore, yourScore int, isMain bool) bool {
	diff := myScore - yourScore

	if isMain {
		// check against main winning threshold if (isMain)
		if s.MainScoreInstantWinAt > 0 && myScore >= s.MainScoreInstantWinAt {
			return true
		}
		if myScore >= s.MainScoreWinsAt && diff >= s.MainScoreMustWinBy {
			return true
		}
		return false
	}
	// otherwise check against secondary
	if s.SecondaryScoreInstantWinAt > 0 && myScore >= s.SecondaryScoreInstantWinAt {
		return true
	}
	if myScore >= s.SecondaryScoreWinsAt && diff >= s.SecondaryScoreMustWinBy {
		return true
	}
	return false

}

func (s *ScoringStructure) DynamicallyValid(db common.DatabaseProvider) error {
	return common.ExistsById(db, &User{}, s.Owner.RecordId())
}

func (s *ScoringStructure) IsComposite() bool {
	return s.SecondaryScoreCountingType != NotApplicable
}

var TennisScoringStructure = ScoringStructure{
	WinConditionCountingType:   Set,
	MainScoreWinsAt:            2,
	MainScoreMustWinBy:         1,
	SecondaryScoreCountingType: Game,
	SecondaryScoreWinsAt:       6,
	SecondaryScoreMustWinBy:    2,
	SecondaryScoreInstantWinAt: 7,
}

var ThreeOutOfFiveGamesTo11 = ScoringStructure{
	WinConditionCountingType:   Game,
	MainScoreWinsAt:            3,
	MainScoreMustWinBy:         1,
	SecondaryScoreCountingType: Point,
	SecondaryScoreWinsAt:       11,
	SecondaryScoreMustWinBy:    2,
}
