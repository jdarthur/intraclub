package middleware

import "fmt"

type UserBasedRecord interface {
	GetUserId() string
}

func ValidateRecordIsOwnedByUser(r UserBasedRecord, userIdInToken string) error {
	if r.GetUserId() != userIdInToken {
		return fmt.Errorf("user %s cannot modify record owned by user %s", userIdInToken, r.GetUserId())
	}

	return nil
}
