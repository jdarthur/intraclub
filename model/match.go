package model

import (
	"fmt"
	"intraclub/common"
)

type MatchId common.RecordId

func (id MatchId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id MatchId) String() string {
	return id.RecordId().String()
}

type MatchStatus int

const (
	MatchUnstarted MatchStatus = iota
	MatchInProgress
	MatchWon
	MatchLost
)

func (status MatchStatus) String() string {
	switch status {
	case MatchUnstarted:
		return "Unstarted"
	case MatchInProgress:
		return "In progress"
	case MatchWon:
		return "Won"
	case MatchLost:
		return "Lost"
	default:
		return "Unknown"
	}
}

type CompletedSecondary struct {
	UsValue   int
	ThemValue int
}

func (sc CompletedSecondary) Reverse() CompletedSecondary {
	return CompletedSecondary{
		UsValue:   sc.ThemValue,
		ThemValue: sc.UsValue,
	}
}

func (sc CompletedSecondary) Won() bool {
	return sc.UsValue > sc.ThemValue
}

type Match struct {
	ID             MatchId
	Editors        []UserId
	Opponent       MatchId
	Structure      ScoringStructureId
	MainValue      int
	SecondaryValue int
	WinOverride    bool
	Status         MatchStatus
	_structure     *ScoringStructure `json:"-" bson:"-"`
	_completed     []CompletedSecondary
}

func NewMatch() *Match {
	return &Match{}
}

func (s *Match) Type() string {
	return "match"
}

func (s *Match) GetId() common.RecordId {
	return s.ID.RecordId()
}

func (s *Match) SetId(id common.RecordId) {
	s.ID = MatchId(id)
}

func (s *Match) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return UserIdListToRecordIdList(s.Editors)
}

func (s *Match) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (s *Match) SetOwner(recordId common.RecordId) {
	s.Editors = append(s.Editors, UserId(recordId))
}

func (s *Match) StaticallyValid() error {
	if len(s.Editors) == 0 {
		return fmt.Errorf("no editors specified")
	}
	if s.MainValue < 0 {
		return fmt.Errorf("main value is negative")
	}
	if s.SecondaryValue < 0 {
		return fmt.Errorf("secondary value is negative")
	}
	return nil
}

func (s *Match) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &ScoringStructure{}, s.Structure.RecordId())
	if err != nil {
		return err
	}

	if s.Opponent != MatchId(common.InvalidRecordId) {
		// check if we have a set value first, so that we can
		// create one score pointing to nothing successfully,
		// then create a second score pointing to the first
		opp, err := common.GetExistingRecordById(db, &Match{}, s.Opponent.RecordId())
		if err != nil {
			return err
		}
		if opp.Opponent != MatchId(common.InvalidRecordId) && opp.Opponent != s.ID {
			return fmt.Errorf("this record's opponent %s is pointing to a different opponent than this record (%s)", s.Opponent, opp.Opponent)
		}
	}

	for _, editor := range s.Editors {
		err = common.ExistsById(db, &User{}, editor.RecordId())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Match) Initialize(db common.DatabaseProvider) error {
	if s._structure == nil {
		v, err := common.GetExistingRecordById(db, &ScoringStructure{}, s.Structure.RecordId())
		if err != nil {
			return err
		}
		s._structure = v
	}
	return nil
}

func (s *Match) Victorious(opp *Match) bool {
	if s.WinOverride == true {
		return true
	}
	if s._structure == nil {
		panic("match._structure is not initialized")
	}

	return s._structure.WinningScore(s.MainValue, opp.MainValue, true)
}

func (s *Match) WonSecondary(opp *Match) bool {
	if s.WinOverride == true {
		return true
	}
	if s._structure == nil {
		panic("match._structure is not initialized")
	}

	return s._structure.WinningScore(s.SecondaryValue, opp.SecondaryValue, false)
}

func (s *Match) MarkStatus(db common.DatabaseProvider, newStatus MatchStatus, opp *Match) error {
	s.Status = newStatus

	oppStatus := MatchInProgress
	if newStatus == MatchWon {
		oppStatus = MatchLost
	}

	opp.Status = oppStatus
	return common.UpdateOne(db, opp)
}

func (s *Match) AddCompletedSecondary(db common.DatabaseProvider) error {
	completedValue := CompletedSecondary{
		UsValue: s.SecondaryValue,
	}

	opp, err := common.GetExistingRecordById(db, &Match{}, s.Opponent.RecordId())
	if err != nil {
		return err
	}
	completedValue.ThemValue = opp.SecondaryValue
	s._completed = append(s._completed, completedValue)

	opp._completed = append(opp._completed, completedValue.Reverse())
	err = common.UpdateOne(db, opp)
	if err != nil {
		return err
	}

	return common.UpdateOne(db, s)
}

func (s *Match) IncrementSecondary(db common.DatabaseProvider) error {
	s.SecondaryValue += 1

	if s.Status != MatchUnstarted {
		s.Status = MatchWon
	}

	opp, err := s.GetOpponent(db)
	if err != nil {
		return err
	}

	if s.WonSecondary(opp) {
		err := s.AddCompletedSecondary(db)
		if err != nil {
			return err
		}

		s.MainValue += 1
		s.SecondaryValue = 0
		if s.Victorious(opp) {
			err = s.MarkStatus(db, MatchWon, opp)
			if err != nil {
				return err
			}
		} else {
			err = s.ResetSecondaryForOpponent(db, opp)
			if err != nil {
				return err
			}
		}
	}
	return common.UpdateOne(db, s)
}

func (s *Match) ResetSecondaryForOpponent(db common.DatabaseProvider, opp *Match) error {
	opp.SecondaryValue = 0
	return common.UpdateOne(db, opp)
}

func (s *Match) GetOpponent(db common.DatabaseProvider) (*Match, error) {
	return common.GetExistingRecordById(db, &Match{}, s.Opponent.RecordId())
}

func (s *Match) String(opp *Match) string {
	if s._structure == nil {
		panic("Match._structure is not initialized!")
	}
	output := fmt.Sprintf("Match %s\n", s.ID)
	output += "------------------------------------\n"
	output += fmt.Sprintf("Status: %s\n", s.Status)
	output += "Results:\n"
	name := s._structure.WinConditionCountingType.String()
	for i, completed := range s._completed {
		won := "won"
		if !completed.Won() {
			won = "lost"
		}
		output += fmt.Sprintf("   %s %d: %d-%d (%s)\n", name, i+1, completed.UsValue, completed.ThemValue, won)
	}

	if !s.Victorious(opp) && !opp.Victorious(s) {
		won := ""
		if s.Victorious(opp) {
			won = " (won)"
		} else if opp.Victorious(s) {
			won = " (lost)"
		}
		output += fmt.Sprintf("   %s %d: %d-%d%s\n", name, len(s._completed)+1, s.SecondaryValue, opp.SecondaryValue, won)
	}
	return output
}
