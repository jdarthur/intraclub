package model

import (
	"fmt"
	"strings"
)

type PhoneNumber string

func (p PhoneNumber) AddDashes() PhoneNumber {
	if strings.Count(string(p), "-") == 0 {
		pn := fmt.Sprintf("%s-%s-%s", string(p[0:3]), string(p[3:6]), string(p[6:]))
		return PhoneNumber(pn)
	}
	return p
}

func (p PhoneNumber) StaticallyValid() error {
	for i, char := range p {
		if (char < '0' || char > '9') && char != '-' {
			return fmt.Errorf("invalid character: '%s' (at index %d, correct format: %s)", string(char), i, p.ValidFormat())
		}
	}

	if strings.Count(string(p), "-") == 0 {
		return nil
	} else if strings.Count(string(p), "-") == 1 {
		return fmt.Errorf("invalid phone number '%s'  (1 dash, correct format: %s)", string(p), p.ValidFormat())
	} else if strings.Count(string(p), "-") > 2 {
		return fmt.Errorf("invalid phone number '%s' (>2 dashes, correct format: %s)", string(p), p.ValidFormat())
	}

	lengthWithoutDashes := len(p) - strings.Count(string(p), "-")
	if lengthWithoutDashes != 10 {
		return fmt.Errorf("invalid phone number '%s' (wrong length %d, correct format: %s)", string(p), lengthWithoutDashes, p.ValidFormat())
	}

	return p.dashesInCorrectLocation()
}

func (p PhoneNumber) dashesInCorrectLocation() error {
	if p[3] != '-' || p[7] != '-' {
		return fmt.Errorf("invalid phone number '%s' (dashes in wrong location, correct format: %s)", string(p), p.ValidFormat())
	}
	return nil
}
func (p PhoneNumber) ValidFormat() string {
	return "123-456-7890"
}
