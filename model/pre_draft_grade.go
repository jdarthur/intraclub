package model

import (
	"fmt"
	"intraclub/common"
	"sort"
)

type PreDraftRatingModifier int

const (
	WeakModifier PreDraftRatingModifier = iota
	AverageModifier
	StrongModifier
)

func (p PreDraftRatingModifier) Int() int {
	return int(p)
}

type PreDraftGrade struct {
	ID       common.RecordId        // unique ID of this PreDraftGrade
	PlayerId UserId                 // ID of the User who is being graded
	DraftId  DraftId                // ID of the Draft that this PreDraftGrade pertains to
	GraderId UserId                 // ID of the User providing a Grade
	Modifier PreDraftRatingModifier // 0, 1, or 2 to indicate that this is a weak, average, or strong version of this Rating
	Rating   RatingId
}

func (p *PreDraftGrade) GetOwner() common.RecordId {
	return p.GraderId.RecordId()
}

func (p *PreDraftGrade) UniquenessEquivalent(other *PreDraftGrade) error {
	if p.PlayerId == other.PlayerId && p.GraderId == other.GraderId && p.DraftId == other.DraftId {
		return fmt.Errorf("duplicate record for player ID, grader ID and draft ID")
	}
	return nil
}

type PreDraftAggregate struct {
	PlayerId  UserId
	Aggregate float64 // average of the ratings for this player (higher is better)
}

func (p *PreDraftGrade) SetOwner(recordId common.RecordId) {
	p.GraderId = UserId(recordId)
}

func NewPreDraftGrade() *PreDraftGrade {
	return &PreDraftGrade{}
}

func (p *PreDraftGrade) Type() string {
	return "pre_draft_grade"
}

func (p *PreDraftGrade) GetId() common.RecordId {
	return p.ID
}

func (p *PreDraftGrade) SetId(id common.RecordId) {
	p.ID = id
}

func (p *PreDraftGrade) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{p.GraderId.RecordId()}
}

func (p *PreDraftGrade) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (p *PreDraftGrade) StaticallyValid() error {
	if p.Modifier < 0 || p.Modifier > 2 {
		return fmt.Errorf("modifier must be 0, 1, or 2")
	}
	return nil
}

func (p *PreDraftGrade) DynamicallyValid(db common.DatabaseProvider) error {
	// user ID being graded must be valid in DB
	err := common.ExistsById(db, &User{}, p.PlayerId.RecordId())
	if err != nil {
		return err
	}

	// grader must be a valid user in DB
	err = common.ExistsById(db, &User{}, p.GraderId.RecordId())
	if err != nil {
		return err
	}

	// draft must exist in db
	draft, err := common.GetExistingRecordById(db, &Draft{}, p.DraftId.RecordId())
	if err != nil {
		return err
	}

	if draft.IsDraftCompleted() {
		return fmt.Errorf("draft is already completed")
	}

	if !draft.IsInDraftList(p.PlayerId) {
		return fmt.Errorf("player ID '%s' is not in draft list", p.PlayerId)
	}

	format, err := common.GetExistingRecordById(db, &Format{}, draft.Format.RecordId())
	if err != nil {
		return err
	}

	if !format.IsRatingValidForFormat(p.Rating) {
		return fmt.Errorf("rating %s is not a valid rating for draft's format %s", p.Rating, draft.Format.RecordId())
	}
	return nil
}

// NumericRating calculates a numeric rating for this PreDraftGrade based
// on the provided Format
//
// rating base value is calculated as
//   - no rating: 0
//   - lowest rating: 1 + modifier
//   - second-lowest rating: 3 + 1 + modifier
//   - ...
//   - highest rating: number of possible ratings * 3 + 1 + modifier
//
// so that in a 1/2/3 rating system:
//   - an unrated player would be given a 0 numeric rating
//   - a WeakModifier 3 would get a 1 numeric rating
//   - an AverageModifier 3 would get a 2 numeric rating
//   - a StrongModifier 3 would get a 3 numeric rating
//   - a WeakModifier 2 would get a 4 numeric rating
//   - an AverageModifier 2 would get a 5 numeric rating
//   - a StrongModifier 2 would get a 6 numeric rating
//   - a WeakModifier 1 would get a 7 numeric rating
//   - an AverageModifier 1 would get an 8 numeric rating
//   - a StrongModifier 1 would get a 9 numeric rating
func (p *PreDraftGrade) NumericRating(format *Format) float64 {

	ratingIndex := -1
	for i, rating := range format.PossibleRatings {
		if p.Rating == rating {
			ratingIndex = i
			break
		}
	}
	if ratingIndex == -1 {
		return -1
	}

	// rating base value is calculated as
	// 3 * (len - 1 - ratingIndex) <--- so that a weak 2 is 1 higher than a strong 3
	// + 1                         <--- so that an unrated player is lower than the weakest 3
	// + modifier                  <--- so that a strong 3 is higher than a weak or average 3
	ratingBaseValue := (len(format.PossibleRatings)-ratingIndex-1)*3 + 1
	return float64(ratingBaseValue + p.Modifier.Int())
}

func GetPreDraftGradesByGraderId(db common.DatabaseProvider, graderId UserId) ([]*PreDraftGrade, error) {
	return common.GetAllWhere(db, &PreDraftGrade{}, func(c *PreDraftGrade) bool {
		return c.GraderId == graderId
	})
}

func GetPreDraftGradesByPlayerId(db common.DatabaseProvider, playerId UserId) ([]*PreDraftGrade, error) {
	return common.GetAllWhere(db, &PreDraftGrade{}, func(c *PreDraftGrade) bool {
		return c.PlayerId == playerId
	})
}

func GetPreDraftGradesByDraftId(db common.DatabaseProvider, draftId DraftId) ([]*PreDraftGrade, error) {
	return common.GetAllWhere(db, &PreDraftGrade{}, func(c *PreDraftGrade) bool {
		return c.DraftId == draftId
	})
}

func GetDraftAggregateForPlayer(allGrades []*PreDraftGrade, format *Format, id UserId) PreDraftAggregate {

	// filter down to only the grades for this particular player ID
	gradesForThisPlayer := make([]*PreDraftGrade, 0)
	for _, grade := range allGrades {
		if grade.PlayerId == id {
			gradesForThisPlayer = append(gradesForThisPlayer, grade)
		}
	}

	// get the numeric value of each grade
	numeric := 0.0
	for _, grade := range gradesForThisPlayer {
		numeric += grade.NumericRating(format)
	}
	// get the average of all numeric ratings
	if len(gradesForThisPlayer) > 0 {
		// don't divide by zero if we don't have any grades
		numeric /= float64(len(gradesForThisPlayer))
	}

	return PreDraftAggregate{
		PlayerId:  id,
		Aggregate: numeric,
	}
}

func GetSortedListOfAllPreDraftGradesDescending(db common.DatabaseProvider, draft *Draft) ([]PreDraftAggregate, error) {

	// Get all pre draft grades for this draft
	allGrades, err := GetPreDraftGradesByDraftId(db, draft.ID)
	if err != nil {
		return nil, err
	}

	// get the format for the draft (to calculate the numeric value of each PreDraftGrade)
	format, err := common.GetExistingRecordById(db, &Format{}, draft.Format.RecordId())
	if err != nil {
		return nil, err
	}

	// for each player in the available-to-draft list, get their pre-draft aggregate
	aggregates := make([]PreDraftAggregate, 0)
	for _, player := range draft.Available {
		// get pre-draft aggregate for each player in the list
		a := GetDraftAggregateForPlayer(allGrades, format, player)
		aggregates = append(aggregates, a)
	}

	// sort the aggregate grades from highest grade to lowest grade
	sort.Slice(aggregates, func(i, j int) bool {
		return aggregates[i].Aggregate > aggregates[j].Aggregate
	})
	return aggregates, nil
}
