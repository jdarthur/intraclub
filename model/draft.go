package model

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"intraclub/common"
)

// TeamCaptainAssignment is a struct which binds a TeamId to its captain
// This is used in the draft logic to keep track of rules around selection,
// e.g. who is on the clock, captains picking themselves, etc. without having
// to re-query the team from the database on each call
type TeamCaptainAssignment struct {
	TeamId    TeamId `json:"team_id"`
	CaptainId UserId `json:"captain_idt"`
}

func getDraftContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = cancel
	return ctx
}

func (t TeamCaptainAssignment) StaticallyValid() error {
	return nil
}

// DynamicallyValid validates that the TeamCaptainAssignment has a real TeamId and a
// captain corresponding to a valid UserId, and that the UserId is actually the captain on
// the Team record corresponding to the TeamId
func (t TeamCaptainAssignment) DynamicallyValid(ctx context.Context, db common.DatabaseProvider) error {
	team, err := common.GetExistingRecordById(ctx, db, &Team{}, t.TeamId.RecordId())
	if err != nil {
		return err
	}

	captain, err := team.GetCaptain(ctx, db)
	if err != nil {
		return fmt.Errorf("team has no captain: %w", err)
	}

	if captain != t.CaptainId {
		return fmt.Errorf("team/captain assignment (%s/%s) in draft does not match captain set on team record (%s)", t.TeamId, t.CaptainId, captain)
	}

	return common.ExistsById(ctx, db, &User{}, t.CaptainId.RecordId())
}

// DraftSelection is a struct which stores the results of the Draft.
// It indicated which User was taken in which round & pick, and what
// RatingId that the user will consequently have based on the rating
// cutoff values assigned for the Draft
type DraftSelection struct {
	Round  int
	Pick   int
	User   *User
	Rating RatingId
}

type DraftId common.RecordId

func (id DraftId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id DraftId) String() string {
	return id.RecordId().String()
}

type Draft struct {
	ID                DraftId           `json:"id"`
	Name              string            `json:"name"`                // descriptive name for the draft, e.g. "2025 Men's Intraclub". this will become the default name of the Season post-draft
	Owner             UserId            `json:"owner"`               // This will be the commissioner of the league
	Format            FormatId          `json:"format"`              // Format in which the Season associated with this draft will be played
	RatingCutoffs     map[RatingId]int  `json:"rating_cutoffs"`      // map from rating ID to the last selection index matching that ID
	CompletedAt       time.Time         `json:"completed_at"`        // timestamp when the draft was completed
	DraftOrderPattern DraftOrderPattern `json:"draft_order_pattern"` // draft order pattern, e.g. snake, straight-up, etc.
}

func (d *Draft) SetOwner(recordId common.RecordId) {
	d.Owner = UserId(recordId)
}

func (d *Draft) GetOwner() common.RecordId {
	return d.Owner.RecordId()
}

func NewDraft() *Draft {
	return &Draft{
		DraftOrderPattern: DraftOrderPatternSnake{},
	}
}

func (d *Draft) EditableBy(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{d.Owner.RecordId()}
}

func (d *Draft) AccessibleTo(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (d *Draft) StaticallyValid() error {
	// don't need to validate anything here as everything should be dynamically validated
	return nil
}

func (d *Draft) DynamicallyValid(ctx context.Context, db common.DatabaseProvider) error {

	// validate that the owner exists
	err := common.ExistsById(ctx, db, &User{}, d.Owner.RecordId())
	if err != nil {
		return err
	}

	// validate that the format exists
	format, err := common.GetExistingRecordById(ctx, db, &Format{}, d.Format.RecordId())
	if err != nil {
		return err
	}

	if d.RatingCutoffs != nil {
		// if the ratings cutoff are set, validate them, i.e.
		// we have an increasing rating cutoff for each possible
		// rating in the format, and that the last rating does not
		// have a cutoff assigned to it
		err = d.ValidateRatingsCutoff(format.PossibleRatings)
		if err != nil {
			return err
		}
	}

	captains, err := d.GetCaptains(ctx, db)
	if err != nil {
		return err
	}

	for _, captain := range captains {
		err = captain.DynamicallyValid(ctx, db)
		if err != nil {
			return err
		}
	}

	availablePlayers, err := d.GetAvailablePlayers(ctx, db)
	if err != nil {
		return err
	}

	availablePlayerMap := make(map[UserId]bool)
	for _, ap := range availablePlayers {
		availablePlayerMap[ap] = true
	}

	for _, captain := range captains {
		if !availablePlayerMap[captain.CaptainId] {
			return fmt.Errorf("captain %s is not in the available players list", captain.CaptainId)
		}
	}

	return nil
}

func (d *Draft) Type() string {
	return "draft"
}

func (d *Draft) GetId() common.RecordId {
	return d.ID.RecordId()
}

func (d *Draft) SetId(id common.RecordId) {
	d.ID = DraftId(id)
}

// IsInDraftList returns true if a particular UserId is present in
// the available-to-draft list for this Draft
func (d *Draft) IsInDraftList(ctx context.Context, db common.DatabaseProvider, userId UserId) bool {
	availablePlayers, err := d.GetAvailablePlayers(ctx, db)
	if err != nil {
		return false
	}
	for _, a := range availablePlayers {
		if a == userId {
			return true
		}
	}
	return false
}

// IsSelected returns true if this particular User has been selected
func (d *Draft) IsSelected(ctx context.Context, db common.DatabaseProvider, userId UserId) bool {
	picks, err := d.GetPicks(ctx, db)
	if err != nil {
		return false
	}
	for _, p := range picks {
		if p.UserId == userId {
			return true
		}
	}
	return false
}

// GetAvailablePlayers returns all userIds who are available to be drafted
func (d *Draft) GetAvailablePlayers(ctx context.Context, db common.DatabaseProvider) ([]UserId, error) {
	availablePlayers, err := common.GetAllWhere[*DraftAvailablePlayer](ctx, db, func(_ context.Context, dap *DraftAvailablePlayer) bool {
		return dap.DraftId == d.ID
	})
	if err != nil {
		return nil, err
	}
	result := make([]UserId, 0, len(availablePlayers))
	for _, ap := range availablePlayers {
		result = append(result, ap.PlayerId)
	}
	return result, nil
}

// GetPicks returns all draft picks ordered by their creation
func (d *Draft) GetPicks(ctx context.Context, db common.DatabaseProvider) ([]*DraftPick, error) {
	return common.GetAllWhere[*DraftPick](ctx, db, func(_ context.Context, dp *DraftPick) bool {
		return dp.DraftId == d.ID
	})
}

// GetCaptains retrieves all team captain assignments for this draft from the DraftCaptain join table.
func (d *Draft) GetCaptains(ctx context.Context, db common.DatabaseProvider) ([]*DraftCaptain, error) {
	captains, err := common.GetAllWhere[*DraftCaptain](ctx, db, func(_ context.Context, dc *DraftCaptain) bool {
		return dc.DraftId == d.ID
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(captains, func(i, j int) bool {
		return captains[i].DraftOrder < captains[j].DraftOrder
	})

	return captains, nil
}

// GetAllAvailableToSelect returns a list of all UserId values which the provided captain
// is allowed to select. This excludes captains (except themselves) and already-selected players.
func (d *Draft) GetAllAvailableToSelect(captainId UserId, db common.DatabaseProvider) []UserId {
	ctx := getDraftContext()
	output := make([]UserId, 0)
	availablePlayers, _ := d.GetAvailablePlayers(ctx, db)
	for _, v := range availablePlayers {
		// check if this user is a different captain
		err := d.IsADifferentCaptainId(getDraftContext(), v, captainId, db)

		// if the user is not a different captain, and they are
		// not already selected, add to the available list
		if err == nil && !d.IsSelected(ctx, db, v) {
			output = append(output, v)
		}
	}
	return output
}

// GetCaptainOnTheClock determines which captain is currently on the clock to make a selection,
// based on the draft order pattern and the current round/pick count.
func (d *Draft) GetCaptainOnTheClock(ctx context.Context, db common.DatabaseProvider) (UserId, error) {
	captains, err := d.GetCaptains(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("no captains set for draft")
	}
	if len(captains) == 0 {
		return 0, fmt.Errorf("no captains set for draft")
	}

	// get the round and pick of the next selection
	round, pick := d.GetRoundAndPick(getDraftContext(), db)
	// fmt.Printf("%d.%02d\n", round, pick)

	// get the captain index based on our draft order
	captainIndex := d.DraftOrderPattern.GetCaptainOnTheClock(round, pick, len(captains))
	return captains[captainIndex].CaptainId, nil
}

// IsADifferentCaptainId checks if the given player ID belongs to a different captain
// assigned to this Draft. Returns an error if the player is a different captain,
// which prevents a captain from drafting another captain.
func (d *Draft) IsADifferentCaptainId(ctx context.Context, player, captain UserId, db common.DatabaseProvider) error {
	captains, err := d.GetCaptains(ctx, db)
	if err != nil {
		return nil
	}
	for _, otherCaptain := range captains {
		// captain can draft themselves but not a different captain
		if otherCaptain.CaptainId != captain && player == otherCaptain.CaptainId {
			return fmt.Errorf("captain %s cannot select another captain %s", captain, otherCaptain.CaptainId)
		}
	}
	return nil
}

// SelectByCaptain selects a player on behalf of a captain. Validates that:
// 1. The captain is currently on the clock to make a selection
// 2. The player is not another captain
// 3. The player is in the available-to-draft list
// 4. The player has not already been selected
func (d *Draft) SelectByCaptain(ctx context.Context, player, captain UserId, db common.DatabaseProvider) error {
	// get the captain who is currently on the clock. If the
	// captains have not yet been set, this will return an error
	onTheClock, err := d.GetCaptainOnTheClock(ctx, db)
	if err != nil {
		return err
	}

	// if this captain is not the one currently on the clock, return an error
	if onTheClock != captain {
		return fmt.Errorf("captain with id %s is not on-the-clock (expected %s)", captain, onTheClock)
	}

	// return an error if the selected user ID belongs to a different captain
	err = d.IsADifferentCaptainId(getDraftContext(), player, captain, db)
	if err != nil {
		return err
	}

	// select the player, returning an error if they are not available to select
	return d.Select(getDraftContext(), player, db)
}

// Select creates a new DraftPick record for the given player. Validates that:
// 1. The player is in the available-to-draft list
// 2. The player has not already been selected
// The round, pick number, and rating are automatically calculated based on the draft state.
func (d *Draft) Select(ctx context.Context, player UserId, db common.DatabaseProvider) error {
	if !d.IsInDraftList(getDraftContext(), db, player) {
		return fmt.Errorf("user with id %s is not in the available-to-draft list", player)
	}
	if d.IsSelected(getDraftContext(), db, player) {
		return fmt.Errorf("user with id %s has already been selected", player)
	}

	// Get the last pick index to determine round and pick
	picks, _ := d.GetPicks(getDraftContext(), db)
	round, pick := d.GetRoundAndPickFromPicks(getDraftContext(), db, len(picks))

	// Find the team for this pick
	captains, _ := d.GetCaptains(getDraftContext(), db)
	var teamId TeamId
	// Use the captain's draft order to determine the team for this pick
	if round%2 != 0 {
		teamId = captains[pick-1].TeamId
	} else {
		teamId = captains[len(captains)-pick].TeamId
	}

	// Get the rating for this pick
	format, _ := common.GetExistingRecordById(getDraftContext(), db, &Format{}, d.Format.RecordId())
	rating := d.GetRatingForPick(format.PossibleRatings, len(picks))

	// Create the pick record
	draftPick := &DraftPick{
		DraftId: d.ID,
		TeamId:  teamId,
		UserId:  player,
		Round:   round,
		Pick:    pick,
		Rating:  rating,
	}
	_, err := common.CreateOne(ctx, db, draftPick)
	return err
}

// GetTeamIndexByTeam returns the 0-based index for a team in the draft captain assignment order.
// This index is used to determine the team's draft position in each round.
func (d *Draft) GetTeamIndexByTeam(ctx context.Context, db common.DatabaseProvider, teamId TeamId) (int, error) {
	captains, err := d.GetCaptains(ctx, db)
	if err != nil {
		return -1, err
	}
	for i, assignment := range captains {
		if assignment.TeamId == teamId {
			return i, nil
		}
	}
	return -1, fmt.Errorf("team id %s not found", teamId)
}

// GetTeamIndexByCaptain returns the 0-based index for a team captain in the draft order.
// This is used to retrieve draft results for a particular captain.
func (d *Draft) GetTeamIndexByCaptain(ctx context.Context, db common.DatabaseProvider, captainId UserId) (int, error) {
	captains, err := d.GetCaptains(ctx, db)
	if err != nil {
		return -1, err
	}
	for i, assignment := range captains {
		if assignment.CaptainId == captainId {
			return i, nil
		}
	}
	return -1, fmt.Errorf("captain id %s not found", captainId)
}

// GetRoundAndPick returns the round and pick number for the next selection,
// based on the current number of picks already made.
func (d *Draft) GetRoundAndPick(ctx context.Context, db common.DatabaseProvider) (round int, pick int) {
	picks, _ := d.GetPicks(getDraftContext(), db)
	return d.GetRoundAndPickFromPicks(getDraftContext(), db, len(picks))
}

// GetRoundAndPickFromPicks calculates the round and pick number from a selection index.
// For example, index 8 with 4 teams returns round 3, pick 1.
func (d *Draft) GetRoundAndPickFromPicks(ctx context.Context, db common.DatabaseProvider, selectionIndex int) (round int, pick int) {
	captains, _ := d.GetCaptains(getDraftContext(), db)
	return (selectionIndex / len(captains)) + 1, (selectionIndex % len(captains)) + 1
}

// GetDraftSelections returns the list of DraftSelection results for a particular team,
// ordered by pick number within each round.
func (d *Draft) GetDraftSelections(db common.DatabaseProvider, teamId TeamId) ([]DraftSelection, error) {
	teamIndex, err := d.GetTeamIndexByTeam(getDraftContext(), db, teamId)
	if err != nil {
		return nil, err
	}
	return d.getDraftSelectionsByTeamIndex(getDraftContext(), db, teamIndex)
}

// GetDraftSelectionsByCaptainId returns the list of DraftSelection results for a particular captain.
// This is a convenience method that looks up the captain's team and retrieves their selections.
func (d *Draft) GetDraftSelectionsByCaptainId(ctx context.Context, db common.DatabaseProvider, captainId UserId) ([]DraftSelection, error) {
	teamIndex, err := d.GetTeamIndexByCaptain(getDraftContext(), db, captainId)
	if err != nil {
		return nil, err
	}
	return d.getDraftSelectionsByTeamIndex(getDraftContext(), db, teamIndex)
}

// GetRatingForPick returns the rating assigned to a pick based on the draft's rating cutoffs.
// The rating is determined by comparing the pick number against the stored cutoffs.
func (d *Draft) GetRatingForPick(ratings []RatingId, pick int) RatingId {
	// check if this pick is below one of the cutoffs
	for _, rating := range ratings[:len(ratings)-1] {
		cutoff := d.RatingCutoffs[rating]
		if pick <= cutoff {
			return rating
		}
	}
	// if not, this pick is assigned the lowest rating
	return ratings[len(ratings)-1]
}

func (d *Draft) ValidateRatingsCutoff(ratings []RatingId) error {
	allButOneRatings := ratings[:len(ratings)-1]
	lastRating := ratings[len(ratings)-1]

	cutoffBefore := -1
	for _, rating := range allButOneRatings {
		v, ok := d.RatingCutoffs[rating]
		if !ok {
			return fmt.Errorf("rating cutoff for rating %s not found", rating)
		}
		if v <= 0 {
			return fmt.Errorf("rating cutoff for rating %s must be greater than zero (got %d)", rating, v)
		}

		if v <= cutoffBefore {
			return fmt.Errorf("rating cutoff for rating %s (%d) is <= the one before (%d)", rating, v, cutoffBefore)
		}

		cutoffBefore = v
	}

	_, ok := d.RatingCutoffs[lastRating]
	if ok {
		return fmt.Errorf("lowest rating %s must not have a rating cutoff", lastRating)
	}
	return nil
}

func (d *Draft) getDraftSelectionsByTeamIndex(ctx context.Context, db common.DatabaseProvider, teamIndex int) ([]DraftSelection, error) {
	selections := make([]DraftSelection, 0)

	picks, err := d.GetPicks(getDraftContext(), db)
	if err != nil {
		return nil, err
	}

	captains, err := d.GetCaptains(getDraftContext(), db)
	if err != nil {
		return nil, err
	}

	targetTeamId := captains[teamIndex].TeamId

	for _, pick := range picks {
		if pick.TeamId == targetTeamId {
			user, err := common.GetExistingRecordById(ctx, db, &User{}, pick.UserId.RecordId())
			if err != nil {
				return nil, err
			}

			selections = append(selections, DraftSelection{Round: pick.Round, Pick: pick.Pick, User: user, Rating: pick.Rating})
		}
	}
	return selections, nil
}

func (d *Draft) GetAvailableRatings(ctx context.Context, db common.DatabaseProvider) ([]RatingId, error) {
	format, err := common.GetExistingRecordById(ctx, db, &Format{}, d.Format.RecordId())
	if err != nil {
		return nil, err
	}
	return format.PossibleRatings, nil
}

func (d *Draft) Initialize(ctx context.Context, db common.DatabaseProvider, captains []UserId) error {
	captainsList, err := d.GetCaptains(getDraftContext(), db)
	if err != nil || len(captainsList) != 0 {
		return errors.New("draft is already initialized")
	}

	picksList, err := d.GetPicks(getDraftContext(), db)
	if err != nil || len(picksList) != 0 {
		return errors.New("draft has selections assigned before initialization")
	}

	for i, captain := range captains {
		err = common.ExistsById(ctx, db, &User{}, captain.RecordId())
		if err != nil {
			return err
		}
		name := fmt.Sprintf("Team %d", i+1)
		team := NewDefaultTeam(captain, name)
		team, err = common.CreateOne(ctx, db, team)
		if err != nil {
			return err
		}

		// Create a draft captain record
		draftCaptain := &DraftCaptain{
			DraftId:    d.ID,
			TeamId:     team.ID,
			CaptainId:  captain,
			DraftOrder: i,
		}
		_, err = common.CreateOne(ctx, db, draftCaptain)
		if err != nil {
			return err
		}

		// add this captain to the available players list
		if !d.IsInDraftList(getDraftContext(), db, captain) {
			availablePlayer := &DraftAvailablePlayer{
				DraftId:  d.ID,
				PlayerId: captain,
			}
			_, err = common.CreateOne(ctx, db, availablePlayer)
			if err != nil {
				return err
			}
		}
	}
	return common.UpdateOne(ctx, db, d)
}

func (d *Draft) AssignDraftablePlayers(ctx context.Context, db common.DatabaseProvider, players []UserId) error {
	if d.IsDraftCompleted(ctx, db) {
		return errors.New("draft is already completed")
	}

	for _, player := range players {
		if !d.IsInDraftList(getDraftContext(), db, player) {
			availablePlayer := &DraftAvailablePlayer{
				DraftId:  d.ID,
				PlayerId: player,
			}
			_, err := common.CreateOne(getDraftContext(), db, availablePlayer)
			if err != nil {
				return err
			}
		}
	}
	return common.UpdateOne(getDraftContext(), db, d)
}

func (d *Draft) AssignDraftedPlayersToTeams(ctx context.Context, db common.DatabaseProvider) error {
	if !d.IsDraftCompleted(ctx, db) {
		return errors.New("draft is not yet completed")
	}

	captains, err := d.GetCaptains(getDraftContext(), db)
	if err != nil {
		return err
	}

	for _, assignment := range captains {
		// get the team record corresponding to the assignment
		team, err := common.GetExistingRecordById(getDraftContext(), db, &Team{}, assignment.TeamId.RecordId())
		if err != nil {
			return err
		}

		// get all the draft selections for this team
		results, err := d.GetDraftSelectionsByCaptainId(getDraftContext(), db, assignment.CaptainId)
		if err != nil {
			return err
		}

		// create team assignment for each drafted player
		for _, result := range results {
			// check if user is already assigned to this team
			isMember, err := team.IsTeamMember(ctx, db, result.User.ID)
			if err != nil {
				return err
			}
			if !isMember {
				// create team assignment for this player
				teamAssignment := &TeamAssignment{
					TeamId: team.ID,
					UserId: result.User.ID,
					Role:   TeamRoleMember,
				}
				_, err = common.CreateOne(ctx, db, teamAssignment)
				if err != nil {
					return err
				}
				// also add to ratings map
				team.RatingsMap[result.User.ID] = result.Rating
			}
		}

		// update the team in the database
		err = common.UpdateOne(ctx, db, team)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Draft) CreateSeason(ctx context.Context, db common.DatabaseProvider, name string, facility FacilityId, t StartTime) (*Season, error) {
	if !d.IsDraftCompleted(ctx, db) {
		return nil, errors.New("draft is not completed")
	}

	// update the completed-at timestamp for this draft
	d.CompletedAt = time.Now()
	err := common.UpdateOne(ctx, db, d)
	if err != nil {
		return nil, err
	}

	s := NewSeason()

	// fill fields from the function args
	s.Name = name
	s.Facility = facility
	s.StartTime = t
	s.DraftId = d.ID

	// create a new Season record first
	s, err = common.CreateOne(ctx, db, s)
	if err != nil {
		return nil, err
	}

	// autofill the remaining required fields from the draft data
	err = s.AddCommissioner(getDraftContext(), db, d.Owner)
	if err != nil {
		return nil, err
	}

	captains, err := d.GetCaptains(getDraftContext(), db)
	if err != nil {
		return nil, err
	}

	for _, tca := range captains {
		err = s.AddTeam(getDraftContext(), db, tca.TeamId)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (d *Draft) IsDraftCompleted(ctx context.Context, db common.DatabaseProvider) bool {
	availablePlayers, err := d.GetAvailablePlayers(getDraftContext(), db)
	if err != nil {
		return false
	}
	if len(availablePlayers) == 0 {
		return false
	}

	picks, err := d.GetPicks(getDraftContext(), db)
	if err != nil {
		return false
	}
	return len(availablePlayers) == len(picks)
}

func (d *Draft) GetSeason(ctx context.Context, db common.DatabaseProvider) (*Season, error) {
	seasons, err := common.GetAllWhere[*Season](ctx, db, func(_ context.Context, c *Season) bool {
		return c.DraftId == d.ID
	})
	if err != nil {
		return nil, err
	}
	if len(seasons) == 0 {
		return nil, nil
	}
	return seasons[0], nil
}

func (d *Draft) BlankRecord() common.CrudRecord {
	return new(Draft)
}
