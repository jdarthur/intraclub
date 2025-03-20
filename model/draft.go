package model

import (
	"errors"
	"fmt"
	"intraclub/common"
)

// TeamCaptainAssignment is a struct which binds a TeamId to its captain
// This is used in the draft logic to keep track of rules around selection,
// e.g. who is on the clock, captains picking themselves, etc. without having
// to re-query the team from the database on each call
type TeamCaptainAssignment struct {
	TeamId    TeamId
	CaptainId UserId
}

func (t TeamCaptainAssignment) StaticallyValid() error {
	return nil
}

// DynamicallyValid validates that the TeamCaptainAssignment has a real TeamId and a
// captain corresponding to a valid UserId, and that the UserId is actually the captain on
// the Team record corresponding to the TeamId
func (t TeamCaptainAssignment) DynamicallyValid(db common.DatabaseProvider) error {
	team, err := common.GetExistingRecordById(db, &Team{}, t.TeamId.RecordId())
	if err != nil {
		return err
	}

	if team.Captain != t.CaptainId {
		return fmt.Errorf("team/captain assignment (%s/%s) in draft does not match captain set on team record (%s)", t.TeamId, t.CaptainId, team.Captain)
	}

	return common.ExistsById(db, &User{}, t.CaptainId.RecordId())
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
	ID            DraftId
	Owner         UserId                   // This will be the commissioner of the league
	Captains      []*TeamCaptainAssignment // This stores a cache of team and captain ID for validation purposes
	Available     []UserId                 // All user IDs who are available to be drafted
	Selections    []UserId                 // All user IDs who have been drafted, in order
	Format        FormatId                 // Format in which the Season associated with this draft will be played
	RatingCutoffs map[RatingId]int         // map from rating ID to the last selection index matching that ID
}

func (d *Draft) SetOwner(recordId common.RecordId) {
	d.Owner = UserId(recordId)
}

func NewDraft() *Draft {
	return &Draft{}
}

func (d *Draft) EditableBy(common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{d.Owner.RecordId()}
}

func (d *Draft) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (d *Draft) StaticallyValid() error {
	if len(d.Available) == 0 {
		return errors.New("no available players to draft")
	}
	return nil
}

func (d *Draft) DynamicallyValid(db common.DatabaseProvider) error {

	// validate that the owner exists
	err := common.ExistsById(db, &User{}, d.Owner.RecordId())
	if err != nil {
		return err
	}

	// validate that the format exists
	format, err := common.GetExistingRecordById(db, &Format{}, d.Format.RecordId())
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

	for _, captainAssignment := range d.Captains {
		// validate that the team-to-captain assignment is correct
		err = captainAssignment.DynamicallyValid(db)
		if err != nil {
			return err
		}

		// validate that the captain is in the available-to-draft list
		if !d.IsInDraftList(captainAssignment.CaptainId) {
			return fmt.Errorf("captain %s is not in draft list", captainAssignment.CaptainId)
		}
	}

	// validate that each user marked available-to-draft exists
	for _, a := range d.Available {
		err = common.ExistsById(db, &User{}, a.RecordId())
		if err != nil {
			return err
		}
	}

	for i, s := range d.Selections {
		// validate that each selection is in the available-to-draft list
		// we don't need to validate the user ID against the DB as we just
		// did that for each available-to-draft player
		if !d.IsInDraftList(s) {
			return fmt.Errorf("user %s selected at index %d is not in draft list", s, i)
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
func (d *Draft) IsInDraftList(userId UserId) bool {
	for _, a := range d.Available {
		if a == userId {
			return true
		}
	}
	return false
}

// IsSelected returns true if this particular User has been selected
func (d *Draft) IsSelected(userId UserId) bool {
	for _, a := range d.Selections {
		if a == userId {
			return true
		}
	}
	return false
}

// IsAvailableToSelect returns true if this UserId is present in the
// available-to-draft list and hasn't already been selected
func (d *Draft) IsAvailableToSelect(userId UserId) bool {
	if !d.IsInDraftList(userId) {
		return false
	}
	return !d.IsSelected(userId)
}

// GetAllAvailableToSelect returns a list of all UserId values which
// the provided captain is allowed to select.
func (d *Draft) GetAllAvailableToSelect(captainId UserId) []UserId {
	output := make([]UserId, 0)
	for _, v := range d.Available {
		// check if this user is a different captain
		err := d.IsADifferentCaptainId(v, captainId)

		// if the user is not a different captain, and they are
		// not already selected, add to the available list
		if err == nil && !d.IsSelected(v) {
			output = append(output, v)
		}
	}
	return output
}

// GetCaptainOnTheClock gets which captain is currently on the clock
// to select, based on the order in the TeamCaptainAssignment
func (d *Draft) GetCaptainOnTheClock() (UserId, error) {
	if len(d.Captains) == 0 {
		return 0, fmt.Errorf("no captains set for draft")
	}

	// get the round and pick of the next selection
	round, pick := d.GetRoundAndPick(len(d.Selections))

	// if this is an even round, we draft in reverse order (snake draft)
	if round%2 == 0 {
		return d.Captains[len(d.Captains)-pick].CaptainId, nil
	}

	// otherwise we draft in the order of the TeamCaptainAssignment
	return d.Captains[pick-1].CaptainId, nil
}

// IsADifferentCaptainId checks if this player ID belongs to a different captain
// assigned to this Draft. This is used to validate that a captain cannot select
// another captain in the Draft.
func (d *Draft) IsADifferentCaptainId(player, captain UserId) error {
	for _, otherCaptain := range d.Captains {
		// captain can draft themselves but not a different captain
		if otherCaptain.CaptainId != captain && player == otherCaptain.CaptainId {
			return fmt.Errorf("captain %s cannot select another captain %s", captain, otherCaptain.CaptainId)
		}
	}
	return nil
}

// SelectByCaptain selects a particular player by a particular captain. This validates that
// the given captain is currently on the clock and that the player is
func (d *Draft) SelectByCaptain(player, captain UserId) error {
	// get the captain who is currently on the clock. If the
	// captains have not yet been set, this will return an error
	onTheClock, err := d.GetCaptainOnTheClock()
	if err != nil {
		return err
	}

	// if this captain is not the one currently on the clock, return an error
	if onTheClock != captain {
		return fmt.Errorf("captain with id %s is not on-the-clock", player)
	}

	// return an error if the selected user ID belongs to a different captain
	err = d.IsADifferentCaptainId(player, captain)
	if err != nil {
		return err
	}

	// select the player, returning an error if they are not available to select
	return d.Select(player)
}

// Select validates that this particular user is available to select and
// adds them to the selections list if so
func (d *Draft) Select(u UserId) error {
	if !d.IsAvailableToSelect(u) {
		return fmt.Errorf("user with id %s is not in the available list", u)
	}
	d.Selections = append(d.Selections, u)
	return nil
}

// GetTeamIndexByTeam gets the index in the team captain assignment list for a particular TeamId
func (d *Draft) GetTeamIndexByTeam(teamId TeamId) (int, error) {
	for i, assignment := range d.Captains {
		if assignment.TeamId == teamId {
			return i, nil
		}
	}
	return -1, fmt.Errorf("team id %s not found", teamId)
}

// GetTeamIndexByCaptain gets the index in the team captain assignment list for a particular
// captain ID. This is used to get draft results for a particular captain
func (d *Draft) GetTeamIndexByCaptain(captainId UserId) (int, error) {
	for i, assignment := range d.Captains {
		if assignment.CaptainId == captainId {
			return i, nil
		}
	}
	return -1, fmt.Errorf("captain id %s not found", captainId)
}

// GetRoundAndPick gets the round and pick for a particular index in the selection list,
// e.g. selections[8] is round 3, pick 1 in a 4-team Draft
func (d *Draft) GetRoundAndPick(selectionIndex int) (round int, pick int) {
	return (selectionIndex / len(d.Captains)) + 1, (selectionIndex % len(d.Captains)) + 1
}

// GetDraftSelections gets the DraftSelection results for a particular team
func (d *Draft) GetDraftSelections(db common.DatabaseProvider, teamId TeamId) ([]DraftSelection, error) {
	teamIndex, err := d.GetTeamIndexByTeam(teamId)
	if err != nil {
		return nil, err
	}
	return d.getDraftSelectionsByTeamIndex(db, teamIndex)
}

func (d *Draft) GetDraftSelectionsByCaptainId(db common.DatabaseProvider, captainId UserId) ([]DraftSelection, error) {
	teamIndex, err := d.GetTeamIndexByCaptain(captainId)
	if err != nil {
		return nil, err
	}
	return d.getDraftSelectionsByTeamIndex(db, teamIndex)
}

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

func (d *Draft) getDraftSelectionsByTeamIndex(db common.DatabaseProvider, teamIndex int) ([]DraftSelection, error) {
	selections := make([]DraftSelection, 0)
	ratings, err := d.GetAvailableRatings(db)
	if err != nil {
		return nil, err
	}

	for i, userId := range d.Selections {
		round, pick := d.GetRoundAndPick(i)

		// this userId belongs to this team if it matches the team's
		// selection index for the given round, e.g. the 1.1 for the
		// team which received the first overall pick or the 2.3 for
		// the team which received the second overall pick
		appliesToTeam := false
		if round%2 != 0 {
			if pick == (teamIndex + 1) {
				// 1->2->3->4 round, e.g. first round, third round, etc.
				appliesToTeam = true
			}
		} else if pick == (len(d.Captains) - teamIndex) {
			// 4->3->2->1 round, e.g. second round, fourth round, etc.
			appliesToTeam = true
		}

		// if this player applies to the team, retrieve the user for
		// this particular selection and add it to the list
		if appliesToTeam {
			user, err := common.GetExistingRecordById(db, &User{}, userId.RecordId())
			if err != nil {
				return nil, err
			}

			rating := d.GetRatingForPick(ratings, i)
			selections = append(selections, DraftSelection{Round: round, Pick: pick, User: user, Rating: rating})
		}
	}
	return selections, nil
}

func (d *Draft) GetAvailableRatings(db common.DatabaseProvider) ([]RatingId, error) {
	format, err := common.GetExistingRecordById(db, &Format{}, d.Format.RecordId())
	if err != nil {
		return nil, err
	}
	return format.PossibleRatings, nil
}

func (d *Draft) Initialize(db common.DatabaseProvider, captains []UserId) error {
	if len(d.Captains) != 0 {
		return errors.New("team captains list is already initialized")
	}

	for i, captain := range captains {
		err := common.ExistsById(db, &User{}, captain.RecordId())
		if err != nil {
			return err
		}
		name := fmt.Sprintf("Team %d", i+1)
		team := NewDefaultTeam(captain, name)
		v, err := common.CreateOne(db, team)
		if err != nil {
			return err
		}

		d.Captains = append(d.Captains, &TeamCaptainAssignment{
			TeamId:    v.ID,
			CaptainId: captain,
		})
	}
	return common.UpdateOne(db, d)
}

func (d *Draft) AssignDraftedPlayersToTeams(db common.DatabaseProvider) error {
	for _, assignment := range d.Captains {
		// get the team record corresponding to the assignment
		team, err := common.GetExistingRecordById(db, &Team{}, assignment.TeamId.RecordId())
		if err != nil {
			return err
		}

		// get all the draft selections for this team
		results, err := d.GetDraftSelectionsByCaptainId(db, assignment.CaptainId)
		if err != nil {
			return err
		}

		for _, result := range results {
			// if drafted player is not already a member of the team, add them.
			// The only added player when this is called should be the captain
			if !team.IsTeamMember(result.User.ID) {
				team.Members = append(team.Members, result.User.ID)
			}
		}

		// update the team in the database
		err = common.UpdateOne(db, team)
		if err != nil {
			return err
		}
	}
	return nil
}
