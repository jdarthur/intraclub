package model

import (
	"fmt"
	"strings"
)

var DefaultAddress = "rcintra.club"
var DefaultSender = EmailAddress(fmt.Sprintf("noreply@%s", DefaultAddress))
var MaxSubjectLength = 255
var MaxBodyLength = 256 * 1024

type Email struct {
	From    EmailAddress   `json:"from"`
	To      []EmailAddress `json:"to"`
	Subject string         `json:"subject"`
	Body    string         `json:"body"`
}

func NewDefaultEmail() *Email {
	return &Email{
		From: DefaultSender,
	}
}

func (e *Email) StaticallyValid() error {

	err := e.From.StaticallyValid()
	if err != nil {
		return fmt.Errorf("from address (%s) is invalid: %s", e.From, err)

	}

	for i, to := range e.To {
		err = to.StaticallyValid()
		if err != nil {
			return fmt.Errorf("to address %d (%s) is invalid: %s", i+1, to, err)
		}
	}

	e.Subject = strings.TrimSpace(e.Subject)
	e.Body = strings.TrimSpace(e.Body)

	if e.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if e.Body == "" {
		return fmt.Errorf("body is required")
	}

	if len(e.Subject) > MaxSubjectLength {
		return fmt.Errorf("subject is too long (%d > %d)", len(e.Subject), MaxSubjectLength)
	}
	if len(e.Body) > MaxBodyLength {
		return fmt.Errorf("body is too long (%d > %d)", len(e.Body), MaxBodyLength)
	}
	return nil
}

func (e *Email) Send() error {
	return nil
}
