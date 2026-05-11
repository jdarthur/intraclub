package model

import (
	"context"
	"fmt"

	"intraclub/api"
	"intraclub/database"
)

type SeasonComposite struct {
	Season          *Season `json:"season"`
	Draft           *Draft  `json:"draft"`
	IsSeasonCreated bool    `json:"is_season_created"`
}

func (s SeasonComposite) StaticallyValid() error {
	return nil
}

func GetMySeasons(ctx context.Context, db database.DatabaseProvider, token *api.AuthToken, asPlayer, asCommissioner bool) ([]*SeasonComposite, error) {
	if asPlayer {
		if asCommissioner {
			return GetMySeasonsAsPlayerOrCommissioner(ctx, db, token)
		}
		return GetMySeasonsAsPlayer(ctx, db, token)
	} else if asCommissioner {
		return GetMySeasonsAsCommissioner(ctx, db, token)
	}
	return nil, fmt.Errorf("must get seasons as either player or commissioner or both, not neither")
}

func GetMySeasonsAsCommissioner(ctx context.Context, db database.DatabaseProvider, token *api.AuthToken) ([]*SeasonComposite, error) {
	drafts, err := database.GetAllWhere[*Draft](ctx, db, func(_ context.Context, c *Draft) bool {
		return isUserCommissioner(c, token.UserId)
	})
	return getSeasonComposites(ctx, db, drafts, err)
}

func GetMySeasonsAsPlayer(ctx context.Context, db database.DatabaseProvider, token *api.AuthToken) ([]*SeasonComposite, error) {
	drafts, err := database.GetAllWhere[*Draft](ctx, db, func(_ context.Context, c *Draft) bool {
		return isUserPlayer(ctx, c, token.UserId, db)
	})
	return getSeasonComposites(ctx, db, drafts, err)
}

func GetMySeasonsAsPlayerOrCommissioner(ctx context.Context, db database.DatabaseProvider, token *api.AuthToken) ([]*SeasonComposite, error) {
	drafts, err := database.GetAllWhere[*Draft](ctx, db, func(_ context.Context, c *Draft) bool {
		return isUserPlayerOrCommissioner(ctx, c, token.UserId, db)
	})
	return getSeasonComposites(ctx, db, drafts, err)
}

// isUserPlayer checks if this user is a player in a particular season,
// by checking if they were selected in the season's draft.
func isUserPlayer(ctx context.Context, c *Draft, userId database.UserId, db database.DatabaseProvider) bool {
	picks, err := c.GetPicks(ctx, db)
	if err != nil {
		return false
	}
	for _, pick := range picks {
		if pick.UserId == userId {
			return true
		}
	}
	return false
}

// isUserCommissioner checks if the user is the commissioner of the provided draft
func isUserCommissioner(c *Draft, userId database.UserId) bool {
	return c.Owner == userId
}

// isUserPlayerOrCommissioner checks if the user is either a player or
// the commissioner of the provided Draft
func isUserPlayerOrCommissioner(ctx context.Context, c *Draft, userId database.UserId, db database.DatabaseProvider) bool {
	return isUserCommissioner(c, userId) || isUserPlayer(ctx, c, userId, db)
}

// getSeasonComposite takes a list of Draft records and returns a list of
// SeasonComposite records. This will pull all the associated Season records
// for a Draft if a season has been created.
func getSeasonComposites(ctx context.Context, db database.DatabaseProvider, drafts []*Draft, err error) ([]*SeasonComposite, error) {
	// The error from the "Get all drafts" query is passed into this function for a bit of
	// code reuse. Just return the error to the end-caller if it is non-nil
	if err != nil {
		return nil, err
	}

	seasons := make([]*SeasonComposite, 0, len(drafts))
	for _, draft := range drafts {

		// get the associated season for this draft if it exists
		season, err := draft.GetSeason(ctx, db)
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
