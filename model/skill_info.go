package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ID             primitive.ObjectID `json:"id" bson:"_id"`          // unique ID, only queryable field (entry IDs are stored in a list on the User collection)
	UserId         string             `json:"user_id" bson:"user_id"` // ID of the User that this record belongs to
	LeagueType     LeagueType         `json:"league_type"`            // Type of league (USTA / ALTA / T2)
	MostRecentYear int                `json:"most_recent_year"`       // Year that you played on this team, e.g. 2024
	Captain        string             `json:"captain"`                // Captain of the team for search convenience, e.g. John Smith
	Level          string             `json:"level"`                  // level of the team, e.g. ALTA level "C-4 Seniors"
	Line           string             `json:"line"`                   // line that you played at, e.g. line 1 or line 5
}

func (s *SkillInfo) GetUserId() string {
	return s.UserId
}

func (s *SkillInfo) SetUserId(userId string) {
	s.UserId = userId
}

func (s *SkillInfo) RecordType() string {
	return "skillInfo"
}

func (s *SkillInfo) OneRecord() common.CrudRecord {
	return s
}

type listOfSkillInfo []*SkillInfo

func (l listOfSkillInfo) Length() int {
	return len(l)
}

func (l listOfSkillInfo) Get(index int) common.CrudRecord {
	return l[index]
}

func (s *SkillInfo) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfSkillInfo, 0)
}

func (s *SkillInfo) SetId(id primitive.ObjectID) {
	s.ID = id
}

func (s *SkillInfo) GetId() primitive.ObjectID {
	return s.ID
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

func (s *SkillInfo) ValidateStatic() error {
	if !IsValidLeagueType(s.LeagueType) {
		return fmt.Errorf("invalid league type: %s", s.LeagueType)
	}

	return nil
}

func (s *SkillInfo) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {
	return common.CheckExistenceOrErrorByStringId(db, &User{}, s.UserId)
}

func GetAllCaptains(db common.DbProvider) ([]string, error) {
	s := SkillInfo{}
	v, err := db.GetAll(&s)
	if err != nil {
		return nil, err
	}

	captainMap := make(map[string]bool)
	common.ForEachCrudRecord(v, func(record common.CrudRecord) {
		si := record.(*SkillInfo)
		captainMap[si.Captain] = true
	})

	uniqueCaptains := make([]string, 0, len(captainMap))
	for captain := range captainMap {
		uniqueCaptains = append(uniqueCaptains, captain)
	}

	return uniqueCaptains, nil
}
