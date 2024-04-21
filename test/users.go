package test

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"intraclub/model"
	"math/rand"
	"testing"
)

// some dummy Player records for 1s, 2s and 3s

var TomEasum = &model.User{
	ID:        primitive.NewObjectID(),
	FirstName: "Tom",
	LastName:  "Easum",
	Email:     "tom@easum.com",
}

var EthanMoland = &model.User{
	FirstName: "Ethan",
	LastName:  "Moland",
	Email:     "ethan@moland.com",
}

var AndyLascik = &model.User{
	FirstName: "Andy",
	LastName:  "Lascik",
	Email:     "andy@lascik.com",
}

var JdArthur = &model.User{
	FirstName: "JD",
	LastName:  "Arthur",
	Email:     "jd@arthur.com",
}

var ChrisBoehm = &model.User{
	FirstName: "Chris",
	LastName:  "Boehm",
	Email:     "chris@boehm.com",
}

var NormTaffet = &model.User{
	FirstName: "Norm",
	LastName:  "Taffet",
	Email:     "norm@taffet.com",
}

var TomerWagshal = &model.User{
	FirstName: "Tomer",
	LastName:  "Wagshal",
	Email:     "tomer@wagshal.com",
}

var PaulCohen = &model.User{
	FirstName: "Paul",
	LastName:  "Cohen",
	Email:     "paul@cohen.com",
}

var KevinCampbell = &model.User{
	FirstName: "Kevin",
	LastName:  "Campbell",
	Email:     "kevin@campbell.com",
}

var UnitTestUsers = []*model.User{
	TomEasum,
	EthanMoland,
	AndyLascik,
	JdArthur,
	ChrisBoehm,
	NormTaffet,
	TomerWagshal,
	PaulCohen,
	KevinCampbell,
}

func newUser(t *testing.T) *model.User {
	user := &model.User{
		IsAdmin:   false,
		FirstName: "test",
		LastName:  "test",
		Email:     fmt.Sprintf("test%d@test.com", rand.Int()),
	}

	created, err := common.Create(common.GlobalDbProvider, user)
	if err != nil {
		t.Fatalf("error creating user: %s\n", err)
	}
	return created.(*model.User)
}
