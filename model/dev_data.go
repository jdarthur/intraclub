package model

import (
	"intraclub/common"
	"sort"
)

func SeedDevData() {
	user := seedDevUsers()
	seedDevScoringStructures(user.ID)
	seedDevRatings(user.ID)
	seedDevFormat(user.ID)
}

func seedDevUsers() *User {
	user1 := NewUser()
	user1.FirstName = "JD"
	user1.LastName = "Arthur"
	user1.Email = "jdarthur@gatech.edu"

	v, err := common.CreateOne(common.GlobalDatabaseProvider, user1)
	if err != nil {
		panic(err)
	}
	return v
}

func seedDevScoringStructures(u UserId) {
	scoringStructure := NewScoringStructure()
	scoringStructure.Name = "Tennis standard set"
	scoringStructure.Owner = u
	scoringStructure.WinConditionCountingType = Game
	scoringStructure.WinCondition = WinCondition{
		WinThreshold:        6,
		MustWinBy:           2,
		InstantWinThreshold: 7,
	}

	v, err := common.CreateOne(common.GlobalDatabaseProvider, scoringStructure)
	if err != nil {
		panic(err)
	}

	scoringStructure2 := NewScoringStructure()
	scoringStructure2.Name = "Tennis standard match"
	scoringStructure2.Owner = u
	scoringStructure2.WinConditionCountingType = Set
	scoringStructure2.WinCondition = WinCondition{
		WinThreshold: 2,
		MustWinBy:    1,
	}
	scoringStructure2.SecondaryScoringStructures = ScoringStructureList{
		v.ID, v.ID, v.ID,
	}

	_, err = common.CreateOne(common.GlobalDatabaseProvider, scoringStructure2)
	if err != nil {
		panic(err)
	}
}

var MensOne = "Men's 1"
var MensTwo = "Men's 2"
var MensThree = "Men's 3"

func seedDevRatings(u UserId) {
	r := NewRating()
	r.UserId = u
	r.Name = MensOne
	r.Description = RatingOne

	_, err := common.CreateOne(common.GlobalDatabaseProvider, r)
	if err != nil {
		panic(err)
	}

	r = NewRating()
	r.UserId = u
	r.Name = MensTwo
	r.Description = RatingTwo

	_, err = common.CreateOne(common.GlobalDatabaseProvider, r)
	if err != nil {
		panic(err)
	}

	r = NewRating()
	r.UserId = u
	r.Name = MensThree
	r.Description = RatingThree

	_, err = common.CreateOne(common.GlobalDatabaseProvider, r)
	if err != nil {
		panic(err)
	}
}

func seedDevFormat(u UserId) {
	ratings, err := common.GetAll[*Rating](common.GlobalDatabaseProvider)
	if err != nil {
		panic(err)
	}

	format := NewFormat()
	format.UserId = u
	format.Name = "Men's Intraclub"

	sort.Slice(ratings, func(i, j int) bool {
		return ratings[i].Name < ratings[j].Name
	})

	for i, rating := range ratings {
		// add possible ratings
		format.PossibleRatings = append(format.PossibleRatings, rating.ID)

		// create lines
		for _, rating2 := range ratings[i:] {
			format.Lines = append(format.Lines, FormatLine{
				Player1Rating: rating.ID,
				Player2Rating: rating2.ID,
			})
		}

	}

	_, err = common.CreateOne(common.GlobalDatabaseProvider, format)
	if err != nil {
		panic(err)
	}
}
