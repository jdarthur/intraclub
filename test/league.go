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

	return &model.League{
		Colors:       make([]model.TeamColor, 0),
		Commissioner: CommissionerId.Hex(),
	}
}
