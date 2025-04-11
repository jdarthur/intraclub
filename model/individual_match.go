package model

import (
	"fmt"
	"intraclub/common"
)

type IndividualMatchId common.RecordId

func (id IndividualMatchId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id IndividualMatchId) String() string {
	return id.RecordId().String()
}

type IndividualMatchStatus int

const (
	MatchUnstarted IndividualMatchStatus = iota
	MatchInProgress
	MatchWon
	MatchLost
)

func (status IndividualMatchStatus) String() string {
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

type IndividualMatch struct {
	ID             IndividualMatchId
	Editors        []UserId
	Opponent       IndividualMatchId
	Structure      ScoringStructureId
	MainValue      int
	SecondaryValue int
	WinOverride    bool
	Status         IndividualMatchStatus
	_structure     *ScoringStructure   `json:"-" bson:"-"`
	_subStructures []*ScoringStructure `json:"-" bson:"-"`
	_completed     []CompletedSecondary
}

func NewMatch() *IndividualMatch {
	return &IndividualMatch{}
}

func (s *IndividualMatch) Type() string {
	return "match"
}

func (s *IndividualMatch) GetId() common.RecordId {
	return s.ID.RecordId()
}

func (s *IndividualMatch) SetId(id common.RecordId) {
	s.ID = IndividualMatchId(id)
}

func (s *IndividualMatch) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return UserIdListToRecordIdList(s.Editors)
}

func (s *IndividualMatch) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (s *IndividualMatch) SetOwner(recordId common.RecordId) {
	s.Editors = append(s.Editors, UserId(recordId))
}

func (s *IndividualMatch) StaticallyValid() error {
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

func (s *IndividualMatch) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &ScoringStructure{}, s.Structure.RecordId())
	if err != nil {
		return err
	}

	if s.Opponent != IndividualMatchId(common.InvalidRecordId) {
		// check if we have a set value first, so that we can
		// create one score pointing to nothing successfully,
		// then create a second score pointing to the first
		opp, err := common.GetExistingRecordById(db, &IndividualMatch{}, s.Opponent.RecordId())
		if err != nil {
			return err
		}
		if opp.Opponent != IndividualMatchId(common.InvalidRecordId) && opp.Opponent != s.ID {
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

func (s *IndividualMatch) Initialize(db common.DatabaseProvider) error {
	if s._structure == nil {
		v, err := common.GetExistingRecordById(db, &ScoringStructure{}, s.Structure.RecordId())
		if err != nil {
			return err
		}
		s._structure = v

		// if the scoring structure is composite, we need to retrieve all
		// the sub-structures referenced by it as well
		if v.IsComposite() {
			for _, id := range v.SecondaryScoringStructures {
				sub, err := common.GetExistingRecordById(db, &ScoringStructure{}, id.RecordId())
				if err != nil {
					return err
				}
				s._subStructures = append(s._subStructures, sub)
			}
		}
	}
	return nil
}

func (s *IndividualMatch) Victorious(opp *IndividualMatch) bool {
	if s.WinOverride == true {
		return true
	}
	if s._structure == nil {
		panic("match._structure is not initialized")
	}

	return s._structure.WinningScore(s.MainValue, opp.MainValue)
}

func (s *IndividualMatch) WonSecondary(opp *IndividualMatch) bool {
	if s.WinOverride == true {
		return true
	}
	if s._structure == nil {
		panic("match._structure is not initialized")
	}
	if !s._structure.IsComposite() {
		panic("match._structure is not composite, WonSecondary does not make sense to call")
	} else {
		if len(s._subStructures) == 0 {
			panic("match._subStructures is not initialized")
		}
	}

	currentSubstructure := s._subStructures[len(s._completed)]
	return currentSubstructure.WinningScore(s.SecondaryValue, opp.SecondaryValue)
}

func (s *IndividualMatch) MarkStatus(db common.DatabaseProvider, newStatus IndividualMatchStatus, opp *IndividualMatch) error {
	s.Status = newStatus

	oppStatus := MatchInProgress
	if newStatus == MatchWon {
		oppStatus = MatchLost
	}

	opp.Status = oppStatus
	return common.UpdateOne(db, opp)
}

func (s *IndividualMatch) AddCompletedSecondary(db common.DatabaseProvider) error {
	completedValue := CompletedSecondary{
		UsValue: s.SecondaryValue,
	}

	opp, err := common.GetExistingRecordById(db, &IndividualMatch{}, s.Opponent.RecordId())
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

func (s *IndividualMatch) IncrementSecondary(db common.DatabaseProvider) error {
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
		}
		err = s.ResetSecondaryForOpponent(db, opp)
		if err != nil {
			return err
		}
	}
	return common.UpdateOne(db, s)
}

func (s *IndividualMatch) ResetSecondaryForOpponent(db common.DatabaseProvider, opp *IndividualMatch) error {
	opp.SecondaryValue = 0
	return common.UpdateOne(db, opp)
}

func (s *IndividualMatch) GetOpponent(db common.DatabaseProvider) (*IndividualMatch, error) {
	return common.GetExistingRecordById(db, &IndividualMatch{}, s.Opponent.RecordId())
}

func (s *IndividualMatch) String(opp *IndividualMatch) string {
	if s._structure == nil {
		panic("IndividualMatch._structure is not initialized!")
	}
	output := fmt.Sprintf("IndividualMatch %s\n", s.ID)
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

func (s *IndividualMatch) GetSecondaryPointTotal() int {
	output := 0
	output += s.SecondaryValue
	for _, v := range s._completed {
		output += v.UsValue
	}
	return output
}
