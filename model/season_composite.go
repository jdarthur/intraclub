package model

import (
	"fmt"
	"intraclub/common"
)

type SeasonComposite struct {
	Season          *Season `json:"season"`
	Draft           *Draft  `json:"draft"`
	IsSeasonCreated bool    `json:"is_season_created"`
}

func (s SeasonComposite) StaticallyValid() error {
	return nil
}

func GetMySeasons(db common.DatabaseProvider, token *common.AuthToken, asPlayer, asCommissioner bool) ([]*SeasonComposite, error) {
	if asPlayer {
		if asCommissioner {
			return GetMySeasonsAsPlayerOrCommissioner(db, token)
		}
		return GetMySeasonsAsPlayer(db, token)
	} else if asCommissioner {
		return GetMySeasonsAsCommissioner(db, token)
	}
	return nil, fmt.Errorf("must get seasons as either player or commissioner or both, not neither")
}

func GetMySeasonsAsCommissioner(db common.DatabaseProvider, token *common.AuthToken) ([]*SeasonComposite, error) {
	drafts, err := common.GetAllWhere(db, &Draft{}, func(c *Draft) bool {
		return isUserCommissioner(c, token.UserId)
	})
	return getSeasonComposites(db, drafts, err)
}

func GetMySeasonsAsPlayer(db common.DatabaseProvider, token *common.AuthToken) ([]*SeasonComposite, error) {
	drafts, err := common.GetAllWhere(db, &Draft{}, func(c *Draft) bool {
		return isUserPlayer(c, token.UserId, db)
	})
	return getSeasonComposites(db, drafts, err)
}

func GetMySeasonsAsPlayerOrCommissioner(db common.DatabaseProvider, token *common.AuthToken) ([]*SeasonComposite, error) {
	drafts, err := common.GetAllWhere(db, &Draft{}, func(c *Draft) bool {
		return isUserPlayerOrCommissioner(c, token.UserId, db)
	})
	return getSeasonComposites(db, drafts, err)
}

// isUserPlayer checks if this user is a player in a particular season,
// by checking if they were selected in the season's draft.
func isUserPlayer(c *Draft, userId common.RecordId, db common.DatabaseProvider) bool {
	picks, err := c.GetPicks(db)
	if err != nil {
		return false
	}
	for _, pick := range picks {
		if pick.UserId.RecordId() == userId {
			return true
		}
	}
	return false
}

// isUserCommissioner checks if the user is the commissioner of the provided draft
func isUserCommissioner(c *Draft, userId common.RecordId) bool {
	return c.Owner.RecordId() == userId
}

// isUserPlayerOrCommissioner checks if the user is either a player or
// the commissioner of the provided Draft
func isUserPlayerOrCommissioner(c *Draft, userId common.RecordId, db common.DatabaseProvider) bool {
	return isUserCommissioner(c, userId) || isUserPlayer(c, userId, db)
}

// getSeasonComposite takes a list of Draft records and returns a list of
// SeasonComposite records. This will pull all the associated Season records
// for a Draft if a season has been created.
func getSeasonComposites(db common.DatabaseProvider, drafts []*Draft, err error) ([]*SeasonComposite, error) {
	// The error from the "Get all drafts" query is passed into this function for a bit of
	// code reuse. Just return the error to the end-caller if it is non-nil
	if err != nil {
		return nil, err
	}

	seasons := make([]*SeasonComposite, 0, len(drafts))
	for _, draft := range drafts {

		// get the associated season for this draft if it exists
		season, err := draft.GetSeason(db)
		if err != nil {
			// if we got an error, return it
			return nil, err
		}

		composite := &SeasonComposite{
			Draft: draft,
		}

		// check if a season has been created for this draft.
		// if so, we will add it to the composite list
		if season != nil {
			composite.Season = season
			composite.IsSeasonCreated = true
		}

		// add at least this particular draft to the list
		seasons = append(seasons, composite)
	}
	return seasons, nil
}
