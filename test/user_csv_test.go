package test

import (
	"fmt"
	"intraclub/model"
	"strings"
	"testing"
)

func TestExpectedHeaders(t *testing.T) {

	actualHeaders := []string{"header1", "header2", "header3"}
	expectedHeaders := []string{"header2", "header3", "header4"}

	err := model.ValidateHeaders(actualHeaders, expectedHeaders)
	if err == nil {
		t.Errorf("Expected non-matching headers to return an error")
	}

	ValidateErrorContains(t, err, "was not in the expected headers list")
}

func TestExpectedHeadersWrongLength(t *testing.T) {

	actualHeaders := []string{"header2", "header3"}
	expectedHeaders := []string{"header2", "header3", "header4"}

	err := model.ValidateHeaders(actualHeaders, expectedHeaders)
	if err == nil {
		t.Errorf("Expected non-matching headers to return an error")
	}

	ValidateErrorContains(t, err, "expected 3 headers")
}

func TestEmptyValue(t *testing.T) {

	v := "First Name, Last Name, Email\n"
	v += "Jim, Bibby, jim@bibby.com\n"
	v += ", Lastname, jim2@bibby.com\n"

	_, err := model.ParseUserCsvFromReader(strings.NewReader(v))
	if err == nil {
		t.Errorf("expected empty value in record to return an error")
	} else {
		ValidateErrorContains(t, err, "empty value for header 'First Name'")
	}
}

func TestValidList(t *testing.T) {

	v := "First Name, Last Name, Email\n"
	v += "Jim, Bibby, jim@bibby.com\n"
	v += "John, Smith, john@smith.com\n"
	v += "Norbert, Totinos, norbert@totinos.com\n"

	expectedUsers := []model.User{
		{FirstName: "Jim", LastName: "Bibby", Email: "jim@bibby.com"},
		{FirstName: "John", LastName: "Smith", Email: "john@smith.com"},
		{FirstName: "Norbert", LastName: "Totinos", Email: "norbert@totinos.com"},
	}

	actualUsers, err := model.ParseUserCsvFromString(v)
	if err != nil {
		t.Fatalf("Got error parsing CSV values: %s", err)
	}

	err = userListIsEquivalent(actualUsers, expectedUsers)
	if err != nil {
		t.Fatalf("Got error comparing user list equivalence: %s", err)
	}
}

func userListIsEquivalent(actual, expected []model.User) error {

	if len(actual) != len(expected) {
		return fmt.Errorf("actual length %d does not match expected length %d", len(actual), len(expected))
	}

	for _, user := range actual {
		found := false
		for _, expectedUser := range expected {
			if user.Email == expectedUser.Email {
				found = true
				if !userIsEquivalent(user, expectedUser) {
					return fmt.Errorf("user \n%+v\n did not match expected user \n%+v", user, expectedUser)
				}
			}
		}
		if !found {
			return fmt.Errorf("user with email %s was not found in expected user list", user.Email)
		}
	}

	return nil
}

func userIsEquivalent(actual, expected model.User) bool {
	return actual.Email == expected.Email &&
		actual.FirstName == expected.FirstName &&
		actual.LastName == expected.LastName
}
