package model

import (
	"fmt"
	"intraclub/common"
	"math/rand"
	"testing"
)

func newStoredDraft(t *testing.T, db common.DatabaseProvider, commissioner UserId) *Draft {
	draft := NewDraft()
	draft.Owner = commissioner
	draft.Available = []UserId{commissioner}
	draft.Format = newDefaultStoredFormat(t, db).ID

	v, err := common.CreateOne(db, draft)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func newRandomDraft(t *testing.T, db common.DatabaseProvider, playerCount, teamCount int) *Draft {
	draft := NewDraft()
	commissioner := newStoredUser(t, db)
	draft.Owner = commissioner.ID
	draft.Format = newDefaultStoredFormat(t, db).ID

	for i := 0; i < teamCount; i++ {
		user := newStoredUser(t, db)
		team := newStoredTeam(t, db, user.ID)
		tca := TeamCaptainAssigment{TeamId: team.ID, CaptainId: user.ID}
		draft.Captains = append(draft.Captains, tca)
		draft.Available = append(draft.Available, user.ID)
	}

	for i := 0; i < playerCount; i++ {
		user := newStoredUser(t, db)
		draft.Available = append(draft.Available, user.ID)
	}

	var err error
	draft, err = common.CreateOne(db, draft)
	if err != nil {
		t.Fatal(err)
	}
	return draft
}

func doRandomDraft(t *testing.T, db common.DatabaseProvider, playerCount int, teamCount int) *Draft {
	draft := newRandomDraft(t, db, playerCount, teamCount)
	for _, v := range draft.Captains {
		err := draft.Select(db, v.CaptainId)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < playerCount; i++ {
		selectRandomAvailable(t, db, draft, UserId(0))
	}
	return draft
}

func selectRandomAvailable(t *testing.T, db common.DatabaseProvider, draft *Draft, captain UserId) {
	available := draft.GetAllAvailableToSelect(captain)
	index := rand.Intn(len(available))
	err := draft.Select(db, available[index])
	if err != nil {
		t.Fatal(err)
	}
}

func selectRandomAvailableByCaptain(t *testing.T, db common.DatabaseProvider, draft *Draft, captain UserId) {
	available := draft.GetAllAvailableToSelect(captain)
	index := rand.Intn(len(available))
	err := draft.SelectByCaptain(db, available[index], captain)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRandomDraft(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := doRandomDraft(t, db, 100, 4)

	if len(draft.GetAllAvailableToSelect(draft.Captains[0].CaptainId)) != 0 {
		t.Fatal("Expected no available users left to draft")
	}
	fmt.Printf("%+v\n", draft)
}

func TestCaptainIsNotInDraftList(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 0, 4)

	draft.Available = []UserId{draft.Owner}
	err := draft.DynamicallyValid(db, nil)
	if err == nil {
		t.Fatal("Expected draft without captain ID in list to be invalid")
	}
	fmt.Println(err)
}

func TestCaptainsCanOnlyBeSelfDrafted(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)
	err := draft.SelectByCaptain(db, draft.Captains[1].CaptainId, draft.Captains[0].CaptainId)
	if err == nil {
		t.Fatal("Expected draft of captain by another captain to be invalid")
	}
	fmt.Println(err)

	err = draft.SelectByCaptain(db, draft.Captains[0].CaptainId, draft.Captains[0].CaptainId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCaptainsIsNotOnTheClock(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)
	err := draft.SelectByCaptain(db, draft.Captains[1].CaptainId, draft.Captains[1].CaptainId)
	if err == nil {
		t.Fatal("Expected selection by captain not on the clock to be invalid")
	}
	fmt.Println(err)
}

func TestSnakeSelection(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 3)

	captain1 := draft.Captains[0].CaptainId
	captain2 := draft.Captains[1].CaptainId
	captain3 := draft.Captains[2].CaptainId

	selectRandomAvailableByCaptain(t, db, draft, captain1)
	selectRandomAvailableByCaptain(t, db, draft, captain2)
	selectRandomAvailableByCaptain(t, db, draft, captain3)
	selectRandomAvailableByCaptain(t, db, draft, captain3)
	selectRandomAvailableByCaptain(t, db, draft, captain2)
	selectRandomAvailableByCaptain(t, db, draft, captain1)
}
