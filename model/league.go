package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

type League struct {
	ID           primitive.ObjectID `json:"league_id" bson:"_id"`
	Colors       []TeamColor        `json:"colors" bson:"colors"`
	Commissioner string             `json:"commissioner" bson:"commissioner"`
}

func (l *League) VerifyUpdatable(c common.CrudRecord) (illegalUpdate bool, field string) {
	existingLeague := c.(*League)

	if l.Commissioner != existingLeague.Commissioner {
		return true, "commissioner"
	}

	return false, ""
}

func (l *League) GetUserId() string {
	return l.Commissioner
}

func (l *League) RecordType() string {
	return "league"
}

func (l *League) OneRecord() common.CrudRecord {
	return new(League)
}

type listOfLeagues []*League

func (l listOfLeagues) Get(index int) common.CrudRecord {
	return l[index]
}

func (l listOfLeagues) Length() int {
	return len(l)
}

func (l *League) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfLeagues, 0)
}

func (l *League) SetId(id primitive.ObjectID) {
	l.ID = id
}

func (l *League) GetId() primitive.ObjectID {
	return l.ID
}

func (l *League) ValidateStatic() error {

	for _, color := range l.Colors {
		err := color.ValidateStatic()
		if err != nil {
			return fmt.Errorf("invalid team color %+v: %s", color, err.Error())
		}
	}

	if len(l.Colors) > 1 {
		return l.CheckDuplicateColors()
	}
	return nil
}

// CheckDuplicateColors validates that each TeamColor in the League.Colors list
// has a unique color name and hex code
func (l *League) CheckDuplicateColors() error {
	for i, color := range l.Colors[:len(l.Colors)-1] {
		for j, color2 := range l.Colors[i+1:] {
			if color.Name == color2.Name {
				return fmt.Errorf("duplicate color name at index %d / %d", i, i+j+1)
			} else if color.Hex == color2.Hex {
				return fmt.Errorf("duplicate color hex code at index %d / %d", i, i+j+1)
			}
		}
	}
	return nil
}

func (l *League) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {

	err := common.CheckExistenceOrErrorByStringId(common.GlobalDbProvider, &User{}, l.Commissioner)
	if err != nil {
		return err
	}

	return nil
}
