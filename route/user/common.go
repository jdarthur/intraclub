package user

import (
	"context"

	"intraclub/database"
	"intraclub/model"
)

var BaseRoute = "/user"

func DoesMatchingUserExist(ctx context.Context, db database.Provider, user *model.User) (bool, *model.User, error) {
	users, err := database.GetAll[*model.User](ctx, db)
	if err != nil {
		return false, nil, err
	}
	for _, u := range users {
		if u.UniquenessEquivalent(user) == nil {
			return true, u, nil
		}
	}
	return false, nil, nil
}
