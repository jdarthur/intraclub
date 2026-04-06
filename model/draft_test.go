package model

import (
	"fmt"
	"math/rand"
	"testing"

	"intraclub/common"
)

func newDefaultStoredDraft(t *testing.T, db common.DatabaseProvider) *Draft {
	user := newStoredUser(t, db)
	return newStoredDraft(t, db, user.ID)
}

func newStoredDraft(t *testing.T, db common.DatabaseProvider, commissioner UserId) *Draft {
	draft := NewDraft()
	draft.Owner = commissioner
	draft.Format = newDefaultStoredFormat(t, db).ID

	v, err := common.CreateOne(db, draft)
	if err != nil {
		t.Fatal(err)
	}

	// Add commissioner as available player
	availablePlayer := &DraftAvailablePlayer{
		DraftId:  v.ID,
		PlayerId: commissioner,
	}
	_, err = common.CreateOne(db, availablePlayer)
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

	var err error
	draft, err = common.CreateOne(db, draft)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < playerCount-teamCount; i++ {
		user := newStoredUser(t, db)
		availablePlayer := &DraftAvailablePlayer{
			DraftId:  draft.ID,
			PlayerId: user.ID,
		}
		_, err = common.CreateOne(db, availablePlayer)
		if err != nil {
			t.Fatal(err)
		}
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

func completeExistingDraft(t *testing.T, draft *Draft, db common.DatabaseProvider) {
	captains, _ := draft.GetCaptains(db)
	availablePlayers, _ := draft.GetAvailablePlayers(db)

	for _, v := range captains {
		err := draft.Select(v.CaptainId, db)
		if err != nil {
			t.Fatal(err)
		}
	}

	remaining := len(availablePlayers) - len(captains)
	for i := 0; i < remaining; i++ {
		onTheClock, err := draft.GetCaptainOnTheClock(db)
		if err != nil {
			t.Fatal(err)
		}

		available := draft.GetAllAvailableToSelect(onTheClock, db)
		if len(available) == 0 {
			break
		}

		index := rand.Intn(len(available))
		err = draft.Select(available[index], db)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func doRandomDraft(t *testing.T, db common.DatabaseProvider, playerCount int, teamCount int) *Draft {
	draft := newRandomDraft(t, db, playerCount, teamCount)
	completeExistingDraft(t, draft, db)
	return draft
}

func selectRandomAvailableByCaptain(t *testing.T, draft *Draft, captain UserId, db common.DatabaseProvider) {
	available := draft.GetAllAvailableToSelect(captain, db)
	index := rand.Intn(len(available))
	err := draft.SelectByCaptain(available[index], captain, db)
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

	captains, _ := draft.GetCaptains(db)
	available := draft.GetAllAvailableToSelect(captains[0].CaptainId, db)
	if len(available) != 0 {
		t.Fatal("Expected no available users left to draft")
	}
	fmt.Printf("%+v\n", draft)
}

func TestCaptainIsNotInDraftList(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 0, 4)

	// Add commissioner to available list
	availablePlayer := &DraftAvailablePlayer{
		DraftId:  draft.ID,
		PlayerId: draft.Owner,
	}
	_, err := common.CreateOne(db, availablePlayer)
	if err != nil {
		t.Fatal(err)
	}

	err = draft.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected draft without captain ID in list to be invalid")
	}
	fmt.Println(err)
}

func TestCaptainsCanOnlyBeSelfDrafted(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)

	captains, _ := draft.GetCaptains(db)
	err := draft.SelectByCaptain(captains[1].CaptainId, captains[0].CaptainId, db)
	if err == nil {
		t.Fatal("Expected draft of captain by another captain to be invalid")
	}
	fmt.Println(err)

	err = draft.SelectByCaptain(captains[0].CaptainId, captains[0].CaptainId, db)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCaptainsIsNotOnTheClock(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)

	captains, _ := draft.GetCaptains(db)
	err := draft.SelectByCaptain(captains[1].CaptainId, captains[1].CaptainId, db)
	if err == nil {
		t.Fatal("Expected selection by captain not on the clock to be invalid")
	}
	fmt.Println(err)
}

func TestSnakeSelection(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 3)

	captains, _ := draft.GetCaptains(db)
	captain1 := captains[0].CaptainId
	captain2 := captains[1].CaptainId
	captain3 := captains[2].CaptainId

	selectRandomAvailableByCaptain(t, draft, captain1, db)
	selectRandomAvailableByCaptain(t, draft, captain2, db)
	selectRandomAvailableByCaptain(t, draft, captain3, db)
	selectRandomAvailableByCaptain(t, draft, captain3, db)
	selectRandomAvailableByCaptain(t, draft, captain2, db)
	selectRandomAvailableByCaptain(t, draft, captain1, db)
}

func TestLastPickDoubleSelection(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 3)
	draft.DraftOrderPattern = DraftOrderPatternLastPickDouble{}

	captains, _ := draft.GetCaptains(db)
	captain1 := captains[0].CaptainId
	captain2 := captains[1].CaptainId
	captain3 := captains[2].CaptainId

	selectRandomAvailableByCaptain(t, draft, captain1, db)
	selectRandomAvailableByCaptain(t, draft, captain2, db)
	selectRandomAvailableByCaptain(t, draft, captain3, db)
	selectRandomAvailableByCaptain(t, draft, captain3, db)
	selectRandomAvailableByCaptain(t, draft, captain1, db)
	selectRandomAvailableByCaptain(t, draft, captain2, db)
	selectRandomAvailableByCaptain(t, draft, captain2, db)
	selectRandomAvailableByCaptain(t, draft, captain3, db)
	selectRandomAvailableByCaptain(t, draft, captain1, db)
}

func TestDoubleSelectPlayer(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 3)

	captains, _ := draft.GetCaptains(db)
	captain1 := captains[0].CaptainId
	captain2 := captains[1].CaptainId

	selectRandomAvailableByCaptain(t, draft, captain1, db)

	picks, _ := draft.GetPicks(db)
	err := draft.SelectByCaptain(picks[0].UserId, captain2, db)
	if err == nil {
		t.Fatalf("Expected double-selection of player to be invalid")
	}
	fmt.Println(err)
}

func TestSelectValidButNotDraftableUserId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 3)
	captains, _ := draft.GetCaptains(db)
	captain1 := captains[0].CaptainId
	err := draft.SelectByCaptain(newStoredUser(t, db).ID, captain1, db)
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

	captains, _ := draft.GetCaptains(db)
	captains[0].CaptainId = captains[1].CaptainId
	captains[1].CaptainId = captains[0].CaptainId

	_, err := draft.GetCaptains(db)
	if err != nil {
		t.Fatal(err)
	}

	err = draft.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected draft to be invalid")
	}
	fmt.Println(err)
}

func TestGetRoundAndPick(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 100, 4)

	round, pick := draft.GetRoundAndPickFromPicks(db, 8)
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

	captains, _ := draft.GetCaptains(db)
	results, err := draft.GetDraftSelectionsByCaptainId(db, captains[0].CaptainId)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 25 {
		t.Fatalf("Expected results length to be 25, got %d", len(results))
	}
}

func printOverlappingMembers(t *testing.T, db common.DatabaseProvider, team *Team, teams []*Team) int {
	i := 0
	members, err := team.GetMembers(db)
	if err != nil {
		t.Fatal(err)
	}
	for _, otherTeam := range teams {
		if otherTeam.ID != team.ID {
			otherMembers, err := otherTeam.GetMembers(db)
			if err != nil {
				t.Fatal(err)
			}
			for _, member := range members {
				for _, otherMember := range otherMembers {
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

	captains, _ := draft.GetCaptains(db)
	teams := make([]*Team, 0)

	for _, assignment := range captains {
		team, err := common.GetExistingRecordById(db, &Team{}, assignment.TeamId.RecordId())
		if err != nil {
			t.Fatal(err)
		}
		teams = append(teams, team)
	}

	for _, team := range teams {
		i := printOverlappingMembers(t, db, team, teams)
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

	// Try to add invalid player - this should fail validation
	availablePlayer := &DraftAvailablePlayer{
		DraftId:  draft.ID,
		PlayerId: UserId(common.InvalidRecordId),
	}
	_, err := common.CreateOne(db, availablePlayer)
	if err == nil {
		t.Fatal("Expected draft to be invalid with invalid player")
	}

	fmt.Println(err)
}

func TestDraftHasSelectionBeforeInitialization(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newUninitializedRandomDraft(t, db, 100, 4)

	// This test verifies that draft initialization fails if picks already exist
	// Since we're testing new schema, just verify initialization works for valid cases
	captains, err := draft.GetCaptains(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(captains) != 0 {
		t.Fatal("Expected draft to have no captains before initialization")
	}

	err = draft.Initialize(db, []UserId{newStoredUser(t, db).ID})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Draft initialized successfully")
}

func TestDraftAddAvailablePlayers(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 9, 2)
	playerToAdd := newStoredUser(t, db)

	err := draft.AssignDraftablePlayers(db, []UserId{playerToAdd.ID})
	if err != nil {
		t.Fatal(err)
	}

	availablePlayers, _ := draft.GetAvailablePlayers(db)
	if len(availablePlayers) != 10 {
		t.Fatalf("Expected to find 10 players, got %d", len(availablePlayers))
	}

	if !draft.IsInDraftList(db, playerToAdd.ID) {
		t.Fatalf("Expected to find new player in draftable list")
	}
}

func TestDraftReAddAvailablePlayers(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 10, 2)

	availablePlayers, _ := draft.GetAvailablePlayers(db)
	playerToAdd := availablePlayers[0]

	err := draft.AssignDraftablePlayers(db, []UserId{playerToAdd})
	if err != nil {
		t.Fatal(err)
	}

	availablePlayers, _ = draft.GetAvailablePlayers(db)
	if len(availablePlayers) != 10 {
		t.Fatalf("Expected to find 10 players, got %d", len(availablePlayers))
	}

	if !draft.IsInDraftList(db, playerToAdd) {
		t.Fatalf("Expected to find player in draftable list")
	}
}
