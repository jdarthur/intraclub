package model

import (
	"errors"
	"fmt"
	"intraclub/common"
)

type TeamCaptainAssigment struct {
	TeamId    TeamId
	CaptainId UserId
}

func (t TeamCaptainAssigment) StaticallyValid() error {
	return nil
}

func (t TeamCaptainAssigment) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &Team{}, t.TeamId.RecordId())
	if err != nil {
		return err
	}

	return common.ExistsById(db, &User{}, t.CaptainId.RecordId())
}

type DraftSelection struct {
	Round int
	Pick  int
	User  *User
}

type DraftId common.RecordId

func (id DraftId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id DraftId) String() string {
	return id.RecordId().String()
}

type Draft struct {
	ID         DraftId
	Owner      UserId
	Captains   []TeamCaptainAssigment
	Available  []UserId
	Selections []UserId
	Format     FormatId
	Started    bool
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

func (d *Draft) DynamicallyValid(db common.DatabaseProvider, existing common.DatabaseValidatable) error {

	err := common.ExistsById(db, &User{}, d.Owner.RecordId())
	if err != nil {
		return err
	}

	err = common.ExistsById(db, &Format{}, d.Format.RecordId())
	if err != nil {
		return err
	}

	for _, captainAssignment := range d.Captains {
		err = captainAssignment.DynamicallyValid(db)
		if err != nil {
			return err
		}
		if !d.IsInDraftList(captainAssignment.CaptainId) {
			return fmt.Errorf("captain %s is not in draft list", captainAssignment.CaptainId)
		}
	}

	for _, a := range d.Available {
		err = common.ExistsById(db, &User{}, a.RecordId())
		if err != nil {
			return err
		}
	}

	for _, s := range d.Selections {
		err = common.ExistsById(db, &User{}, s.RecordId())
		if err != nil {
			return err
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

func (d *Draft) getUser(db common.DatabaseProvider, userId UserId) (*User, error) {
	user, exists, err := common.GetOneById(db, &User{}, userId.RecordId())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("user with id %s not found", userId)
	}
	return user, nil
}

func (d *Draft) IsInDraftList(userId UserId) bool {
	for _, a := range d.Available {
		if a == userId {
			return true
		}
	}
	return false
}

func (d *Draft) IsSelected(userId UserId) bool {
	for _, a := range d.Selections {
		if a == userId {
			return true
		}
	}
	return false
}

func (d *Draft) IsAvailableToSelect(userId UserId) bool {
	if !d.IsInDraftList(userId) {
		return false
	}
	return !d.IsSelected(userId)
}

func (d *Draft) GetAllAvailableToSelect(captainId UserId) []UserId {
	output := make([]UserId, 0)
	for _, v := range d.Available {
		err := d.IsADifferentCaptainId(v, captainId)
		if err == nil && !d.IsSelected(v) {
			output = append(output, v)
		}

	}
	return output
}

func (d *Draft) GetCaptainOnTheClock() (UserId, error) {
	if len(d.Captains) == 0 {
		return 0, fmt.Errorf("no captains set for draft")
	}

	round, pick := d.GetRoundAndPick(len(d.Selections))
	if round%2 == 0 {
		return d.Captains[len(d.Captains)-pick].CaptainId, nil
	}

	return d.Captains[pick-1].CaptainId, nil
}

func (d *Draft) IsADifferentCaptainId(player, captain UserId) error {
	for _, otherCaptain := range d.Captains {
		if otherCaptain.CaptainId != captain && player == otherCaptain.CaptainId {
			return fmt.Errorf("captain %s cannot select another captain %s", captain, otherCaptain.CaptainId)
		}
	}
	return nil
}

func (d *Draft) SelectByCaptain(db common.DatabaseProvider, player, captain UserId) error {
	onTheClock, err := d.GetCaptainOnTheClock()
	if err != nil {
		return err
	}
	if onTheClock != captain {
		return fmt.Errorf("captain with id %s is not on-the-clock", player)
	}

	err = d.IsADifferentCaptainId(player, captain)
	if err != nil {
		return err
	}

	return d.Select(db, player)
}

func (d *Draft) Select(db common.DatabaseProvider, u UserId) error {
	if !d.IsAvailableToSelect(u) {
		return fmt.Errorf("user with id %s is not in the available list", u)
	}
	_, err := d.getUser(db, u)
	if err != nil {
		return err
	}
	d.Selections = append(d.Selections, u)
	return nil
}

func (d *Draft) GetTeamIndexByTeam(teamId TeamId) (int, error) {
	for i, assignment := range d.Captains {
		if assignment.TeamId == teamId {
			return i, nil
		}
	}
	return -1, fmt.Errorf("team id %s not found", teamId)
}

func (d *Draft) GetTeamIndexByCaptain(captainId UserId) (int, error) {
	for i, assignment := range d.Captains {
		if assignment.CaptainId == captainId {
			return i, nil
		}
	}
	return -1, fmt.Errorf("captain id %s not found", captainId)
}

func (d *Draft) GetRoundAndPick(selectionIndex int) (round int, pick int) {
	return (selectionIndex / len(d.Captains)) + 1, (selectionIndex % len(d.Captains)) + 1
}

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

func (d *Draft) getDraftSelectionsByTeamIndex(db common.DatabaseProvider, teamIndex int) ([]DraftSelection, error) {
	selections := make([]DraftSelection, 0)
	for i, userId := range d.Selections {
		round, pick := d.GetRoundAndPick(i)

		// this userId belongs to this team if it matches the team's
		// selection index for the given round, e.g. the 1.1 for the
		// team which received the first overall pick or the 2.3 for
		// the team which received the second overall pick
		appliesToTeam := false
		if round%2 == 0 && pick == (teamIndex+1) {
			// 1->2->3->4 round, e.g. first round, third round, etc.
			appliesToTeam = true
		} else if pick == (len(d.Captains) - teamIndex) {
			// 4->3->2->1 round, e.g. second round, fourth round, etc.
			appliesToTeam = true
		}

		// if this player applies to the team, retrieve the user for
		// this particular selection and add it to the list
		if appliesToTeam {
			user, err := d.getUser(db, userId)
			if err != nil {
				return nil, err
			}
			selections = append(selections, DraftSelection{Round: round, Pick: pick, User: user})
		}
	}
	return selections, nil
}

func (d *Draft) GetAvailableRatings(db common.DatabaseProvider) ([]RatingId, error) {
	ratings := make([]RatingId, 0)
	format, exists, err := common.GetOneById(db, &Format{}, d.Format.RecordId())
	if err != nil {
		return nil, err
	}
	if !exists {
		return ratings, fmt.Errorf("format with ID %s does not exist", d.Format)
	}
	return format.GetAvailableRatings(), nil
}

func ratingInList(rating RatingId, list []RatingId) bool {
	for _, v := range list {
		if v == rating {
			return true
		}
	}
	return false
}
