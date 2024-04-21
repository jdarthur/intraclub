package model

import (
	"errors"
	"fmt"
	"intraclub/common"
)

type SetResult struct {
	ID    string        `json:"id"`
	Team1 TeamSetResult `json:"team1"`
	Team2 TeamSetResult `json:"team2"`
}

type TeamSetResult struct {
	TeamId      string `json:"team_id"`
	GamesWon    int    `json:"games_won"`
	TiebreakSet bool   `json:"tiebreak_set"`
}

func (tsr TeamSetResult) ValidateStatic(otherTeam TeamSetResult) error {
	if tsr.TiebreakSet {
		// 3rd-set tiebreaks are scored 1-0

		if tsr.GamesWon > 1 {
			return errors.New("set was marked as a third-set tiebreak with > 1 games won")
		}

		if tsr.GamesWon == 0 && otherTeam.GamesWon == 0 {
			return errors.New("set was marked as a third-set tiebreak with both teams winning 0 games")
		}

		if tsr.GamesWon == 1 && otherTeam.GamesWon == 1 {
			return errors.New("set was marked as a third-set tiebreak with both teams winning 1 game")
		}

	} else {
		// normal sets are either first to 6, 7-5, or 7-6 if ended in a tiebreak

		if tsr.GamesWon < 0 {
			return errors.New("games won total less than 0")
		}

		if tsr.GamesWon > 7 {
			return errors.New("games won total was greater than 7")
		}

		if tsr.GamesWon == 7 && otherTeam.GamesWon < 5 {
			return errors.New("won 7 games when opponent won four or less games")
		}

		if tsr.GamesWon == 7 && otherTeam.GamesWon == 7 {
			return errors.New("team and opponent both won 7 games")
		}

		if tsr.GamesWon == 6 && otherTeam.GamesWon == 6 {
			return errors.New("team and opponent both won 6 games")
		}

		if tsr.GamesWon == 6 && otherTeam.GamesWon == 5 {
			return errors.New("team marked as winning 6-5")
		}

		if tsr.GamesWon < 6 && otherTeam.GamesWon < 6 {
			return errors.New("neither team won at least 6 games")
		}
	}

	return nil
}

// Tiebreak returns true if one team won the set in a tiebreak
func (s SetResult) Tiebreak() bool {
	return (s.Team1.GamesWon == 7 && s.Team2.GamesWon == 6) || (s.Team2.GamesWon == 7 && s.Team1.GamesWon == 6)
}

// TiebreakWinner returns the team ID of the team that won the tiebreak
func (s SetResult) TiebreakWinner() string {
	if s.Tiebreak() {
		if s.Team1.GamesWon == 7 {
			return s.Team1.TeamId
		} else {
			return s.Team2.TeamId
		}
	} else {
		return ""
	}
}

func (s SetResult) ValidateStatic() error {
	if s.Team1.TeamId == s.Team2.TeamId {
		return errors.New("teams 1 and 2 are identical")
	}

	err := s.Team1.ValidateStatic(s.Team2)
	if err != nil {
		msg := fmt.Sprintf("Error with team 1's set results: %s", err)
		return errors.New(msg)
	}

	err = s.Team2.ValidateStatic(s.Team1)
	if err != nil {
		msg := fmt.Sprintf("Error with team 2's set results: %s", err)
		return errors.New(msg)
	}

	return nil
}

func (s SetResult) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {
	err := common.CheckExistenceOrErrorByStringId(db, &Team{}, s.Team1.TeamId)
	if err != nil {
		return err
	}

	err = common.CheckExistenceOrErrorByStringId(db, &Team{}, s.Team2.TeamId)
	if err != nil {
		return err
	}

	return nil
}
