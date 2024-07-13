package test

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"intraclub/model"
)

var CommissionerId = primitive.ObjectID{}

func newLeague() *model.League {

	commish := &model.User{
		FirstName: "Jim",
		LastName:  "Bibby",
		Email:     "jim@bibby.com",
	}

	u, err := common.Create(common.GlobalDbProvider, commish)
	if err != nil {
		panic(err)
	}

	CommissionerId = u.(*model.User).ID
	return NewLeagueWithCommish(commish)
}

func NewLeagueWithCommish(commish *model.User) *model.League {

	return &model.League{
		Facility:     newFacility(commish).ID.Hex(),
		Colors:       make([]model.TeamColor, 0),
		Commissioner: commish.ID.Hex(),
	}
}

func newFacility(user *model.User) *model.Facility {
	facility := &model.Facility{
		Address:     "1221 Riverside Rd., Roswell, GA 30076",
		Name:        "Martin's Landing River Club",
		Courts:      9,
		LayoutImage: "",
		UserId:      user.GetUserId(),
	}

	v, err := common.Create(common.GlobalDbProvider, facility)
	if err != nil {
		panic(err)
	}

	return v.(*model.Facility)
}
