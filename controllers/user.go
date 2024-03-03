package controllers

import (
	"intraclub/model"
)

func UserExists(userId string) (model.User, bool) {
	return model.User{}, true
}
