package model

import (
	"context"
	"sort"
	"time"

	"intraclub/database"
)

func SeedDevData(db database.Provider) {
	users := seedDevUsers(db)
	// In dev mode the seeded jdarthur@gatech.edu user is the system
	// administrator so the dev / e2e flows (which log in as that user) can
	// exercise sysadmin-only write paths (e.g. season late additions).
	if err := users[0].AssignRole(getDevContext(), db, SystemAdministrator); err != nil {
		panic(err)
	}
	owner := users[0]
	seedDevScoringStructures(db, owner.ID)
	seedDevRatings(db, owner.ID)
	seedDevFormat(db, owner.ID)
}

func getDevContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = cancel
	return ctx
}

func seedDevUsers(db database.Provider) []*User {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devUsers := []struct {
		FirstName string
		LastName  string
		Email     string
	}{
		{"JD", "Arthur", "jdarthur@gatech.edu"},
		{"Avery", "Chen", "avery.chen@example.com"},
		{"Bailey", "Nguyen", "bailey.nguyen@example.com"},
		{"Cameron", "Patel", "cameron.patel@example.com"},
		{"Dana", "Kim", "dana.kim@example.com"},
		{"Elliot", "Garcia", "elliot.garcia@example.com"},
		{"Frankie", "Rossi", "frankie.rossi@example.com"},
		{"Grayson", "Walsh", "grayson.walsh@example.com"},
		{"Harper", "Okafor", "harper.okafor@example.com"},
		{"Imani", "Silva", "imani.silva@example.com"},
	}

	// jdarthur@gatech.edu is the first user (and the dev/e2e login account).
	created := make([]*User, 0, len(devUsers))
	for _, u := range devUsers {
		user := NewUser()
		user.FirstName = u.FirstName
		user.LastName = u.LastName
		user.Email = EmailAddress(u.Email)

		createdUser, err := database.CreateOne(ctx, db, user)
		if err != nil {
			panic(err)
		}
		created = append(created, createdUser)
	}
	return created
}

func seedDevScoringStructures(db database.Provider, u database.UserId) {
	ctx := getDevContext()
	scoringStructure := NewScoringStructure()
	scoringStructure.Name = "Tennis standard set"
	scoringStructure.Owner = u
	scoringStructure.WinConditionCountingType = Game
	scoringStructure.WinCondition = WinCondition{
		WinThreshold:        6,
		MustWinBy:           2,
		InstantWinThreshold: 7,
	}

	v, err := database.CreateOne(ctx, db, scoringStructure)
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
	v2, err := database.CreateOne(ctx, db, scoringStructure2)
	if err != nil {
		panic(err)
	}

	err = v2.SetSecondaryScoringStructures(ctx, db, ScoringStructureList{
		v.ID, v.ID, v.ID,
	})
	if err != nil {
		panic(err)
	}
}

var MensOne = "Men's 1"
var MensTwo = "Men's 2"
var MensThree = "Men's 3"

func seedDevRatings(db database.Provider, u database.UserId) {
	ctx := getDevContext()
	r := NewRating()
	r.UserId = u
	r.Name = MensOne
	r.Description = RatingOne

	_, err := database.CreateOne(ctx, db, r)
	if err != nil {
		panic(err)
	}

	r = NewRating()
	r.UserId = u
	r.Name = MensTwo
	r.Description = RatingTwo

	_, err = database.CreateOne(ctx, db, r)
	if err != nil {
		panic(err)
	}

	r = NewRating()
	r.UserId = u
	r.Name = MensThree
	r.Description = RatingThree

	_, err = database.CreateOne(ctx, db, r)
	if err != nil {
		panic(err)
	}
}

func seedDevFormat(db database.Provider, u database.UserId) {
	ctx := getDevContext()
	ratings, err := database.GetAll[*Rating](ctx, db)
	if err != nil {
		panic(err)
	}

	format := NewFormat()
	format.UserId = u
	format.Name = "Men's Intraclub"

	sort.Slice(ratings, func(i, j int) bool {
		return ratings[i].Name < ratings[j].Name
	})

	possibleRatings := make(RatingList, 0, len(ratings))
	lines := make([]FormatLine, 0)
	for i, rating := range ratings {
		// add possible ratings
		possibleRatings = append(possibleRatings, rating.ID)

		// create lines
		for _, rating2 := range ratings[i:] {
			lines = append(lines, FormatLine{
				Player1Rating: rating.ID,
				Player2Rating: rating2.ID,
			})
		}
	}

	created, err := database.CreateOne(ctx, db, format)
	if err != nil {
		panic(err)
	}
	if err := created.SetPossibleRatings(ctx, db, possibleRatings); err != nil {
		panic(err)
	}
	if err := created.SetLines(ctx, db, lines); err != nil {
		panic(err)
	}
}
