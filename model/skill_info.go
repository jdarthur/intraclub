package model

import (
	"intraclub/common"
)

type LeagueType string

const (
	LeagueTypeALTA = "ALTA"
	LeagueTypeUSTA = "USTA"
	LeagueTypeT2   = "T2"
)

var LeagueTypes = []LeagueType{
	LeagueTypeALTA, LeagueTypeUSTA, LeagueTypeT2,
}

// SkillInfo is a record type that is saved on a User record
// and stores information about this User's tennis skill level.
// For example "I last played around line 3 on so-and-so's C-4
// Men's seniors team in 2024"
type SkillInfo struct {
	ID             common.RecordId `json:"id" bson:"_id"`          // unique ID, only queryable field (entry IDs are stored in a list on the User collection)
	UserId         UserId          `json:"user_id" bson:"user_id"` // ID of the User that this record belongs to
	LeagueType     LeagueType      `json:"league_type"`            // Type of league (USTA / ALTA / T2)
	MostRecentYear int             `json:"most_recent_year"`       // Year that you played on this team, e.g. 2024
	Captain        string          `json:"captain"`                // Captain of the team for search convenience, e.g. John Smith
	Level          string          `json:"level"`                  // level of the team, e.g. ALTA level "C-4 Seniors"
	Line           string          `json:"line"`                   // line that you played at, e.g. line 1 or line 5
}

func (s *SkillInfo) GetOwner() common.RecordId {
	return s.UserId.RecordId()
}

func (s *SkillInfo) Type() string {
	//TODO implement me
	panic("implement me")
}

func (s *SkillInfo) GetId() common.RecordId {
	//TODO implement me
	panic("implement me")
}

func (s *SkillInfo) SetId(id common.RecordId) {
	//TODO implement me
	panic("implement me")
}

func (s *SkillInfo) EditableBy(db common.DatabaseProvider) []common.RecordId {
	//TODO implement me
	panic("implement me")
}

func (s *SkillInfo) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	//TODO implement me
	panic("implement me")
}

func (s *SkillInfo) SetOwner(recordId common.RecordId) {
	//TODO implement me
	panic("implement me")
}

func (s *SkillInfo) StaticallyValid() error {
	//TODO implement me
	panic("implement me")
}

func (s *SkillInfo) DynamicallyValid(db common.DatabaseProvider) error {
	//TODO implement me
	panic("implement me")
}

func IsValidLeagueType(leagueType LeagueType) bool {
	switch leagueType {
	case LeagueTypeALTA:
		return true
	case LeagueTypeUSTA:
		return true
	case LeagueTypeT2:
		return true
	default:
		return false
	}
}

func GetAllCaptains(db common.DatabaseProvider) ([]string, error) {
	v, err := common.GetAll(db, &SkillInfo{})
	if err != nil {
		return nil, err
	}

	output := make([]string, 0)
	for _, info := range v {
		output = append(output, info.Captain)
	}
	return output, nil
}
