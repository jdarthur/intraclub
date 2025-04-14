package model

import (
	"fmt"
	"intraclub/common"
)

type WeeklyMatchupId common.RecordId

func (id WeeklyMatchupId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id WeeklyMatchupId) String() string {
	return id.RecordId().String()
}

type TeamMatchup struct {
	HomeTeam TeamId
	AwayTeam TeamId
	Bye      bool
}

func (t *TeamMatchup) Validate(db common.DatabaseProvider, season *Season) error {
	if t.Bye && t.AwayTeam != TeamId(common.InvalidRecordId) {
		return fmt.Errorf("away team ID must not be set during a bye")
	}

	err := common.ExistsById(db, &Team{}, t.HomeTeam.RecordId())
	if err != nil {
		return fmt.Errorf("home team error: %s", err)
	}

	if !season.IsTeamAssignedToSeason(t.HomeTeam) {
		return fmt.Errorf("home team %s is not assigned to season %s", t.HomeTeam, season.ID)
	}

	if !t.Bye {
		err = common.ExistsById(db, &Team{}, t.AwayTeam.RecordId())
		if err != nil {
			return fmt.Errorf("away team error: %s", err)
		}

		if !season.IsTeamAssignedToSeason(t.AwayTeam) {
			return fmt.Errorf("away team %s is not assigned to season %s", t.AwayTeam, season.ID)
		}
	}
	return nil
}

// WeeklyMatchup is an instance of one or more TeamMatchup s for a given Week
// during a Season's Schedule. It
type WeeklyMatchup struct {
	ID       WeeklyMatchupId
	WeekId   WeekId         // Week that this WeeklyMatchup corresponds to, i.e. a particular date
	SeasonId SeasonId       // Season that this WeeklyMatchup corresponds to
	Matchups []*TeamMatchup // List of TeamMatchup s for this WeeklyMatchup, e.g. team 1 playing team 2, team 3 on bye, etc.
}

func (w *WeeklyMatchup) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

func (w *WeeklyMatchup) UniquenessEquivalent(other *WeeklyMatchup) error {
	if w.SeasonId == other.SeasonId && w.WeekId == other.WeekId {
		return fmt.Errorf("duplicate record for season ID and week ID")
	}
	return nil
}

func NewWeeklyMatchup() *WeeklyMatchup {
	return &WeeklyMatchup{}
}

func (w *WeeklyMatchup) Type() string {
	return "weekly_matchup"
}

func (w *WeeklyMatchup) GetId() common.RecordId {
	return w.ID.RecordId()
}

func (w *WeeklyMatchup) SetId(id common.RecordId) {
	w.ID = WeeklyMatchupId(id)
}

func (w *WeeklyMatchup) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return EditableBySeason(db, w.SeasonId)
}

func (w *WeeklyMatchup) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (w *WeeklyMatchup) SetOwner(recordId common.RecordId) {
	// don't need to do anything here as the ownership of the
	// WeeklyMatchup record type is automatically inferred &
	// enforced by the associated Season assigned to it
}

func (w *WeeklyMatchup) StaticallyValid() error {
	// loop through all matchups but the last to validate that
	// team X or Y is not double-booked into 2 matchups in the
	// same week of play.
	for i, match := range w.Matchups[:len(w.Matchups)-1] {

		// loop through all matchups after this one in the list
		for j, match2 := range w.Matchups[i+1:] {
			if match.HomeTeam == match2.HomeTeam || match.HomeTeam == match2.AwayTeam {
				return fmt.Errorf("home team %s from matchup %d is also playing in matchup %d", match.HomeTeam, i, j+i+1)
			}
			if !match.Bye {
				if match.AwayTeam == match2.AwayTeam || match.AwayTeam == match2.HomeTeam {
					return fmt.Errorf("away team %s from matchup %d is also playing in matchup %d", match.AwayTeam, i, j+i+1)
				}
			}
		}
	}

	return nil
}

func (w *WeeklyMatchup) DynamicallyValid(db common.DatabaseProvider) error {
	week, err := common.GetExistingRecordById(db, &Week{}, w.WeekId.RecordId())
	if err != nil {
		return err
	}

	// validate that the season in question exists.
	season, err := common.GetExistingRecordById(db, &Season{}, w.SeasonId.RecordId())
	if err != nil {
		return err
	}

	// validate that the weekly matchup
	if week.DraftId != season.DraftId {
		return fmt.Errorf("draft %s assigned to season %s does not match draft %s assigned to week", week.Date, w.SeasonId, week.DraftId)
	}

	// validate that each individual matchup is
	for _, matchup := range w.Matchups {
		err = matchup.Validate(db, season)
		if err != nil {
			return err
		}
	}

	return w.ValidateThatEachTeamHasOneMatchup(season)
}

func (w *WeeklyMatchup) ValidateThatEachTeamHasOneMatchup(season *Season) error {
	m := make(map[TeamId]bool)
	for _, matchup := range w.Matchups {
		m[matchup.HomeTeam] = true
		if !matchup.Bye {
			m[matchup.AwayTeam] = true
		}
	}

	if len(m) != len(season.Teams) {
		for _, team := range season.Teams {
			_, ok := m[team]
			if !ok {
				return fmt.Errorf("team %s does not have a matchup or bye", team)
			}
		}
	}
	return nil
}
