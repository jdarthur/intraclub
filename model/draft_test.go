package model

import (
	"fmt"
	"intraclub/common"
	"math/rand"
	"testing"
)

func newDefaultStoredDraft(t *testing.T, db common.DatabaseProvider) *Draft {
	user := newStoredUser(t, db)
	return newStoredDraft(t, db, user.ID)
}

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

func newUninitializedRandomDraft(t *testing.T, db common.DatabaseProvider, playerCount, teamCount int) *Draft {
	draft := NewDraft()
	commissioner := newStoredUser(t, db)
	draft.Owner = commissioner.ID
	draft.Format = newDefaultStoredFormat(t, db).ID

	for i := 0; i < playerCount-teamCount; i++ {
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

func newRandomDraft(t *testing.T, db common.DatabaseProvider, playerCount, teamCount int) *Draft {

	// create an uninitialized draft
	draft := newUninitializedRandomDraft(t, db, playerCount, teamCount)

	// create random captains
	users := make([]UserId, 0, playerCount)
	for i := 0; i < teamCount; i++ {
		user := newStoredUser(t, db)
		users = append(users, user.ID)
	}

	// initialize the team / captain assignments
	err := draft.Initialize(db, users)
	if err != nil {
		t.Fatal(err)
	}

	return draft
}

func completeExistingDraft(t *testing.T, draft *Draft) {
	for _, v := range draft.Captains {
		err := draft.Select(v.CaptainId)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < len(draft.Available)-len(draft.Captains); i++ {
		onTheClock, err := draft.GetCaptainOnTheClock()
		if err != nil {
			t.Fatal(err)
		}
		selectRandomAvailableByCaptain(t, draft, onTheClock)
	}

}

func doRandomDraft(t *testing.T, db common.DatabaseProvider, playerCount int, teamCount int) *Draft {
	draft := newRandomDraft(t, db, playerCount, teamCount)
	completeExistingDraft(t, draft)
	return draft
}

func selectRandomAvailableByCaptain(t *testing.T, draft *Draft, captain UserId) {
	available := draft.GetAllAvailableToSelect(captain)
	index := rand.Intn(len(available))
	err := draft.SelectByCaptain(available[index], captain)
	if err != nil {
		t.Fatal(err)
	}
}

func newCompletedDraft(t *testing.T, db common.DatabaseProvider) (*Draft, *Season) {
	draft := doRandomDraft(t, db, 100, 4)
	facility := newStoredFacility(t, db, draft.Owner)
	season, err := draft.CreateSeason(db, "Test season", facility.ID, NewStartTime(8, 30))
	if err != nil {
		t.Fatal(err)
	}
	return draft, season
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
	err := draft.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected draft without captain ID in list to be invalid")
	}
	fmt.Println(err)
}

func TestCaptainsCanOnlyBeSelfDrafted(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)
	err := draft.SelectByCaptain(draft.Captains[1].CaptainId, draft.Captains[0].CaptainId)
	if err == nil {
		t.Fatal("Expected draft of captain by another captain to be invalid")
	}
	fmt.Println(err)

	err = draft.SelectByCaptain(draft.Captains[0].CaptainId, draft.Captains[0].CaptainId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCaptainsIsNotOnTheClock(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)
	err := draft.SelectByCaptain(draft.Captains[1].CaptainId, draft.Captains[1].CaptainId)
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

	selectRandomAvailableByCaptain(t, draft, captain1)
	selectRandomAvailableByCaptain(t, draft, captain2)
	selectRandomAvailableByCaptain(t, draft, captain3)
	selectRandomAvailableByCaptain(t, draft, captain3)
	selectRandomAvailableByCaptain(t, draft, captain2)
	selectRandomAvailableByCaptain(t, draft, captain1)
}

func TestLastPickDoubleSelection(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 3)
	draft.DraftOrderPattern = DraftOrderPatternLastPickDouble{}

	captain1 := draft.Captains[0].CaptainId
	captain2 := draft.Captains[1].CaptainId
	captain3 := draft.Captains[2].CaptainId

	selectRandomAvailableByCaptain(t, draft, captain1)
	selectRandomAvailableByCaptain(t, draft, captain2)
	selectRandomAvailableByCaptain(t, draft, captain3)
	selectRandomAvailableByCaptain(t, draft, captain3)
	selectRandomAvailableByCaptain(t, draft, captain1)
	selectRandomAvailableByCaptain(t, draft, captain2)
	selectRandomAvailableByCaptain(t, draft, captain2)
	selectRandomAvailableByCaptain(t, draft, captain3)
	selectRandomAvailableByCaptain(t, draft, captain1)

}

func TestDoubleSelectPlayer(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 3)

	captain1 := draft.Captains[0].CaptainId
	captain2 := draft.Captains[1].CaptainId

	selectRandomAvailableByCaptain(t, draft, captain1)
	err := draft.SelectByCaptain(draft.Selections[0], captain2)
	if err == nil {
		t.Fatalf("Expected double-selection of player to be invalid")
	}
	fmt.Println(err)
}

func TestSelectValidButNotDraftableUserId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 3)
	captain1 := draft.Captains[0].CaptainId
	err := draft.SelectByCaptain(newStoredUser(t, db).ID, captain1)
	if err == nil {
		t.Fatalf("Expected double-selection of player to be invalid")
	}
	fmt.Println(err)
}

func TestGetRatingForSelection(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultFormat(t, db)
	draft := newRandomDraft(t, db, 100, 4)
	draft.RatingCutoffs = map[RatingId]int{
		format.PossibleRatings[0]: 20,
		format.PossibleRatings[1]: 40,
		format.PossibleRatings[2]: 70,
	}

	for i := 0; i <= 20; i++ {
		rating := draft.GetRatingForPick(format.PossibleRatings, i)
		if rating != format.PossibleRatings[0] {
			t.Fatalf("Expected rating to be %s, got %s", format.PossibleRatings[0], rating)
		}
	}

	for i := 21; i <= 40; i++ {
		rating := draft.GetRatingForPick(format.PossibleRatings, i)
		if rating != format.PossibleRatings[1] {
			t.Fatalf("Expected rating to be %s, got %s", format.PossibleRatings[1], rating)
		}
	}

	for i := 41; i <= 70; i++ {
		rating := draft.GetRatingForPick(format.PossibleRatings, i)
		if rating != format.PossibleRatings[2] {
			t.Fatalf("Expected rating to be %s, got %s", format.PossibleRatings[2], rating)
		}
	}

	for i := 71; i <= 100; i++ {
		rating := draft.GetRatingForPick(format.PossibleRatings, i)
		if rating != format.PossibleRatings[3] {
			t.Fatalf("Expected rating to be %s, got %s", format.PossibleRatings[3], rating)
		}
	}
}

func TestRatingWithCutoffBelowPrevious(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultFormat(t, db)
	draft := newRandomDraft(t, db, 100, 4)
	draft.RatingCutoffs = map[RatingId]int{
		format.PossibleRatings[0]: 20,
		format.PossibleRatings[1]: 10,
		format.PossibleRatings[2]: 70,
	}

	err := draft.ValidateRatingsCutoff(format.PossibleRatings)
	if err == nil {
		t.Fatal("Expected draft to be invalid")
	}
	fmt.Println(err)
}

func TestRatingCutoffIsZero(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultFormat(t, db)
	draft := newRandomDraft(t, db, 100, 4)
	draft.RatingCutoffs = map[RatingId]int{
		format.PossibleRatings[0]: 0,
		format.PossibleRatings[1]: 10,
		format.PossibleRatings[2]: 70,
	}

	err := draft.ValidateRatingsCutoff(format.PossibleRatings)
	if err == nil {
		t.Fatal("Expected draft to be invalid")
	}
	fmt.Println(err)
}

func TestRatingCutoffIsMissing(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultFormat(t, db)
	draft := newRandomDraft(t, db, 100, 4)
	draft.RatingCutoffs = map[RatingId]int{
		format.PossibleRatings[0]: 5,
		format.PossibleRatings[1]: 10,
	}

	err := draft.ValidateRatingsCutoff(format.PossibleRatings)
	if err == nil {
		t.Fatal("Expected draft to be invalid")
	}
	fmt.Println(err)
}

func TestRatingCutoffForLastRatingIdIsPresent(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultFormat(t, db)
	draft := newRandomDraft(t, db, 100, 4)
	draft.RatingCutoffs = map[RatingId]int{
		format.PossibleRatings[0]: 5,
		format.PossibleRatings[1]: 10,
		format.PossibleRatings[2]: 70,
		format.PossibleRatings[3]: 80,
	}

	err := draft.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected draft to be invalid")
	}
	fmt.Println(err)
}

func TestTeamCaptainAssignmentHasIncorrectCaptainId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)
	draft.Captains[0].CaptainId = draft.Captains[1].CaptainId
	draft.Captains[1].CaptainId = draft.Captains[0].CaptainId

	err := draft.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected draft to be invalid")
	}
	fmt.Println(err)
}

func TestGetRoundAndPick(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)

	round, pick := draft.GetRoundAndPick(8)
	if round != 3 {
		t.Fatalf("Expected round to be 3, got %d", round)
	}
	if pick != 1 {
		t.Fatalf("Expected pick to be 1, got %d", pick)
	}
}

func TestDraftResults(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := doRandomDraft(t, db, 100, 4)

	results, err := draft.GetDraftSelectionsByCaptainId(db, draft.Captains[0].CaptainId)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 25 {
		t.Fatalf("Expected results length to be 25, got %d", len(results))
	}
}

func printOverlappingMembers(team *Team, teams []*Team) int {
	i := 0
	for _, otherTeam := range teams {
		if otherTeam.ID != team.ID {
			for _, member := range team.Members {
				for _, otherMember := range otherTeam.Members {
					if member == otherMember {
						fmt.Printf("Member %s was drafted by teams %s and %s\n", member, team.ID, otherTeam.ID)
						i += 1
					}
				}
			}
		}
	}
	return i
}

func TestTeamAssignmentAfterDraft(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := doRandomDraft(t, db, 100, 4)

	err := draft.AssignDraftedPlayersToTeams(db)
	if err != nil {
		t.Fatal(err)
	}

	teams := make([]*Team, 0)

	for _, assignment := range draft.Captains {
		team, err := common.GetExistingRecordById(db, &Team{}, assignment.TeamId.RecordId())
		if err != nil {
			t.Fatal(err)
		}
		teams = append(teams, team)
	}

	for _, team := range teams {
		i := printOverlappingMembers(team, teams)
		if i != 0 {
			t.Fatalf("Expected team overlapping members to be zero, but got %d", i)
		}
	}
}

func TestDoubleInitializeDraft(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)

	err := draft.Initialize(db, []UserId{})
	if err == nil {
		t.Fatal("Expected draft double-initialize to be invalid")
	}
	fmt.Println(err)
}

func TestInitializeDraftWithInvalidCaptain(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newUninitializedRandomDraft(t, db, 100, 4)

	err := draft.Initialize(db, []UserId{UserId(common.InvalidRecordId)})
	if err == nil {
		t.Fatal("Expected invalid captain ID to be invalid")
	}
	fmt.Println(err)
}

func TestNoAssignedFormat(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)
	draft.Format = FormatId(common.InvalidRecordId)
	err := common.UpdateOne(db, draft)
	if err == nil {
		t.Fatal("Expected draft to be invalid with empty format")
	}

	fmt.Println(err)
}

func TestInvalidAvailablePlayerId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)
	draft.Available = append(draft.Available, UserId(common.InvalidRecordId))
	err := common.UpdateOne(db, draft)
	if err == nil {
		t.Fatal("Expected draft to be invalid with invalid player")
	}

	fmt.Println(err)
}

func TestDraftHasSelectionBeforeInitialization(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newUninitializedRandomDraft(t, db, 100, 4)
	draft.Selections = append(draft.Available, draft.Available[0])
	err := draft.Initialize(db, []UserId{newStoredUser(t, db).ID})
	if err == nil {
		t.Fatal("Expected draft initialization to fail with selections set")
	}

	fmt.Println(err)
}

func TestDraftAddAvailablePlayers(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 9, 2)
	playerToAdd := newStoredUser(t, db)

	err := draft.AssignDraftablePlayers(db, []UserId{playerToAdd.ID})
	if err != nil {
		t.Fatal(err)
	}

	if len(draft.Available) != 10 {
		t.Fatalf("Expected to find 10 players, got %d", len(draft.Available))
	}

	if !draft.IsInDraftList(playerToAdd.ID) {
		t.Fatalf("Expected to find new player in draftable list")
	}
}

func TestDraftReAddAvailablePlayers(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 10, 2)
	playerToAdd := draft.Available[0]

	err := draft.AssignDraftablePlayers(db, []UserId{playerToAdd})
	if err != nil {
		t.Fatal(err)
	}

	if len(draft.Available) != 10 {
		t.Fatalf("Expected to find 10 players, got %d", len(draft.Available))
	}

	if !draft.IsInDraftList(playerToAdd) {
		t.Fatalf("Expected to find new player in draftable list")
	}
}
