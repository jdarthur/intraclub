package model

import (
	"encoding/csv"
	"fmt"
	"intraclub/common"
	"io"
	"os"
	"sort"
	"strings"
)

const (
	FirstNameCsvField = "First Name"
	LastNameCsvField  = "Last Name"
	EmailCsvField     = "Email"
)

var ExpectedHeaders = []string{
	FirstNameCsvField,
	LastNameCsvField,
	EmailCsvField,
}

var HeadersLine = fmt.Sprintf("%s, %s, %s\n", FirstNameCsvField, LastNameCsvField, EmailCsvField)

func ParseUserCsvFromFile(filename string) ([]*User, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return ParseUserCsvFromReader(file)
}

func ParseUserCsvFromString(s string) ([]*User, error) {
	return ParseUserCsvFromReader(strings.NewReader(s))
}

func ParseUserCsvFromReader(reader io.Reader) ([]*User, error) {
	r := csv.NewReader(reader)

	var headersInOrder []string

	users := make([]*User, 0)
	i := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			if i == 0 {
				return nil, fmt.Errorf("no headers found in CSV file")
			}
			break
		}

		if err != nil {
			return nil, err
		}

		if i == 0 {

			// parse the header line in the first iteration of the loop

			headers := TrimHeaders(record) // remove leading/trailing whitespace
			err := ValidateHeaders(headers, ExpectedHeaders)
			if err != nil {
				return nil, err
			}

			// save the headers in order so we can use them to parse record lines
			headersInOrder = headers
		} else {

			// parse the rest of the lines using the headersInOrder from above

			user, err := ParseCsvLine(record, headersInOrder)
			if err != nil {
				return nil, err
			}

			// save this user to the list
			users = append(users, user)
		}

		i += 1
	}

	return users, nil
}

func ParseCsvLine(line, headers []string) (*User, error) {
	user := &User{}
	for i, key := range headers {

		value := line[i]
		value = strings.TrimSpace(value)

		if value == "" {
			return user, fmt.Errorf("got empty value for header '%s' at line %d", key, i)
		}

		if key == FirstNameCsvField {
			user.FirstName = value
		} else if key == LastNameCsvField {
			user.LastName = value
		} else if key == EmailCsvField {
			user.Email = EmailAddress(value)
		} else {
			// shouldn't be able to get here since we validated the headers
			// in ValidateHeaders before we called ParseCsvLine
			return user, fmt.Errorf("invalid header '%s'", key)
		}
	}

	err := user.StaticallyValid()
	if err != nil {
		return user, err
	}

	return user, nil
}

func ParseUserList(db common.DatabaseProvider, input []*User) (newUsers, alreadyExistingUsers []*User, err error) {
	existingUsersInDatabase, err := common.GetAll(db, &User{})
	if err != nil {
		return nil, nil, err
	}

	newUsers = make([]*User, 0)
	alreadyExistingUsers = make([]*User, 0)
	for _, user := range input {
		found := false

		// look for this user in the database list
		for _, existing := range existingUsersInDatabase {
			if existing.Email == user.Email {
				// if already existing, add it to that list and break out of loop
				alreadyExistingUsers = append(alreadyExistingUsers, existing)
				found = true
				break
			}
		}

		// if this user did not exist in the database, add it to the new list
		if !found {
			newUsers = append(newUsers, user)
		}
	}
	return newUsers, alreadyExistingUsers, nil
}

func ParseAndCreateCsvUsers(db common.DatabaseProvider, csvUserList []*User) (createdUsers, existingUsers []*User, err error) {

	newUsers, alreadyExistingUsers, err := ParseUserList(db, csvUserList)
	if err != nil {
		return nil, nil, err
	}

	created := make([]*User, 0)
	for _, user := range newUsers {
		v, err := common.CreateOne(db, user)
		if err != nil {
			return nil, nil, err
		}
		created = append(created, v)
	}

	sort.Slice(created, func(i, j int) bool {
		return created[i].LastName < created[j].LastName
	})

	sort.Slice(alreadyExistingUsers, func(i, j int) bool {
		return alreadyExistingUsers[i].LastName < alreadyExistingUsers[j].LastName
	})

	return created, alreadyExistingUsers, nil
}

func TrimHeaders(headers []string) []string {
	output := make([]string, 0)
	for _, header := range headers {
		output = append(output, strings.TrimSpace(header))
	}
	return output
}

func ValidateHeaders(actualHeaders, expectedHeaders []string) error {

	if len(actualHeaders) != len(expectedHeaders) {
		return fmt.Errorf("expected %d headers (%v), got %d (%v)", len(expectedHeaders), expectedHeaders, len(actualHeaders), actualHeaders)
	}

	for _, actualHeader := range actualHeaders {
		found := false
		for _, expectedHeader := range expectedHeaders {
			if actualHeader == expectedHeader {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("header '%s' was not in the expected headers list (expected: %+v)", actualHeader, expectedHeaders)
		}
	}

	return nil
}
