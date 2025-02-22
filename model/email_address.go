package model

import (
	"errors"
	"fmt"
	"strings"
)

var invalidStartOrEndCharacters = []rune{'@', '.', '_', '-'}

type EmailAddress string

func (e EmailAddress) StaticallyValid() error {

	if !strings.Contains(string(e), "@") {
		return fmt.Errorf("email address must contain @")
	}

	if strings.Count(string(e), "@") != 1 {
		return fmt.Errorf("email address must contain one @")
	}

	err := e.InvalidPrefix()
	if err != nil {
		return err
	}

	err = e.InvalidSuffix()
	if err != nil {
		return err
	}

	foundAtSign := false
	for _, char := range e {
		if !isValidCharacterForEmail(char) {
			return fmt.Errorf("invalid email address character: %s", string(char))
		}
		if char == '@' {
			foundAtSign = true
		}

		if foundAtSign && char == '_' {
			return errors.New("invalid email address character: '_' after @-sign")
		}
	}

	return nil
}

func (e EmailAddress) InvalidPrefix() error {
	for _, char := range invalidStartOrEndCharacters {
		if strings.HasPrefix(string(e), string(char)) {
			return fmt.Errorf("email address cannot start with character '%s'", string(char))
		}
	}
	return nil
}

func (e EmailAddress) InvalidSuffix() error {
	for _, char := range invalidStartOrEndCharacters {
		if strings.HasSuffix(string(e), string(char)) {
			return fmt.Errorf("email address cannot end with character '%s'", string(char))
		}
	}
	return nil
}

func isValidCharacterForEmail(r rune) bool {

	if r >= '0' && r <= '9' {
		return true
	}
	if r >= 'a' && r <= 'z' {
		return true
	}

	if r == '.' || r == '_' || r == '-' || r == '@' {
		return true
	}

	return false
}
