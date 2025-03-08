package model

type ScoreCountingType int

const (
	Point ScoreCountingType = iota
	Game
	Set
	NotApplicable
)

type WinCondition int

type ScoringStructure struct {
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
	SecondaryScoreWinsAt       int

	// SecondaryScoreMustWinBy is a constraint that delays the win condition for the
	// SecondaryScoreCountingType until a team has X amount of that type compared to the
	// other team. For example, scoring might be played to 11, but with a win-by-two constraint
	SecondaryScoreMustWinBy    int

	// SecondaryScoreInstantWinAt is a threshold that, when reached, causes a team to instantly
	// reach the SecondaryScoreCountingType win condition. For example, this can be used to e.g,
	// disregard a "first-to-seven, win-by-two" constraint when either team hits 10
	SecondaryScoreInstantWinAt int
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
