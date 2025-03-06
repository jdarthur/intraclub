package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"math/rand"
	"strings"
	"testing"
)

var FirstNames = []string{
	"Jim", "John", "Jacob", "Joe", "Joel",
	"Jack", "Jeremy", "Jeff", "Jared", "Jackson",
	"Julius", "Jayden", "Jordan", "Josh", "Jesse",
	"Jonah", "Jose", "Jasper", "Josiah", "Jason",
}

var LastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones",
	"Garcia", "Miller", "Davis", "Rodriguez",
	"Martinez", "Hernandez", "Lopez", "Wilson",
}

type Name struct {
	FirstName string
	LastName  string
}

func (n Name) Equals(other Name) bool {
	return n.FirstName == other.FirstName && n.LastName == other.LastName
}

func CreateRandomName() Name {
	firstNameIndex := rand.Intn(len(FirstNames))
	lastNameIndex := rand.Intn(len(LastNames))
	return Name{
		FirstNames[firstNameIndex],
		LastNames[lastNameIndex],
	}
}

var MaxNameListLength = (len(FirstNames) * len(LastNames)) / 2

func GenerateListOfUniqueNames(length int) []Name {
	if length > MaxNameListLength {
		fmt.Printf("too many names (max %d, got %d)", MaxNameListLength, length)
		return nil
	}
	output := make([]Name, 0, length)
	for len(output) < length {
		name := CreateRandomName()
		if !isNameAlreadyInList(name, output) {
			output = append(output, name)
		}
	}
	return output
}

func isNameAlreadyInList(name Name, list []Name) bool {
	for _, listItem := range list {
		if name.Equals(listItem) {
			return true
		}
	}
	return false
}

var CsvSize = 40

func TestListCreateSomeNames(t *testing.T) {
	nameList := GenerateListOfUniqueNames(CsvSize)
	if len(nameList) != CsvSize {
		t.Fatalf("length of nameList is not %d (got %d)", CsvSize, len(nameList))
	}
	for _, name := range nameList[:5] {
		fmt.Printf("  %+v\n", name)
	}
	fmt.Printf("nameList: %v ...\n", nameList[:5])
}

func TestListCreateTooManyNames(t *testing.T) {
	nameList := GenerateListOfUniqueNames(1_000_000)
	if len(nameList) != 0 {
		t.Fatalf("should get empty list when name count is too high")
	}
}

func GenerateUserCsv() (string, error) {
	output := HeadersLine
	names := GenerateListOfUniqueNames(CsvSize)
	if len(names) == 0 {
		return "", errors.New("got empty names list")
	}

	for _, name := range names {
		output += fmt.Sprintf("%s, %s, %s.%s@email.com\n", name.FirstName, name.LastName, name.FirstName, name.LastName)
	}
	return output, nil
}

func TestUserCsv(t *testing.T) {
	csv, err := GenerateUserCsv()
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(csv, "\n")[:5] {
		fmt.Println(line)
	}
	fmt.Println("...")
}

func generateCsvAndParse(t *testing.T) []*User {
	csv, err := GenerateUserCsv()
	if err != nil {
		t.Fatal(err)
	}

	users, err := ParseUserCsvFromString(csv)
	if err != nil {
		t.Fatal(err)
	}
	return users
}

func TestImportCsvToUserList(t *testing.T) {
	users := generateCsvAndParse(t)
	if len(users) != CsvSize {
		t.Fatalf("length of user is not %d (got %d)", CsvSize, len(users))
	}
}

func TestImportCsvToFreshDatabase(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	users := generateCsvAndParse(t)
	newUsers, existingUsers, err := ParseAndCreateCsvUsers(db, users)
	if err != nil {
		t.Fatal(err)
	}

	if len(existingUsers) != 0 {
		t.Fatalf("existing users should be empty")
	}
	if len(newUsers) != len(users) {
		t.Fatalf("new users should be the same length as import (%d, got %d)", len(users), len(newUsers))
	}
	fmt.Printf("newUsers count: %v\n", len(newUsers))
}

func TestParseNewAndExistingUsers(t *testing.T) {

	db := common.NewUnitTestDBProvider()
	users := generateCsvAndParse(t)
	_, _, err := ParseAndCreateCsvUsers(db, users)
	if err != nil {
		t.Fatal(err)
	}

	users2 := generateCsvAndParse(t)
	newUsers, existing, err := ParseUserList(db, users2)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("newUsers: %d\n", len(newUsers))
	fmt.Printf("existingUsers: %d\n", len(existing))
}

func TestPartiallyImportCsvToDatabase(t *testing.T) {

	db := common.NewUnitTestDBProvider()
	users := generateCsvAndParse(t)
	_, _, err := ParseAndCreateCsvUsers(db, users)
	if err != nil {
		t.Fatal(err)
	}

	users2 := generateCsvAndParse(t)
	newUsers, existingUsers, err := ParseAndCreateCsvUsers(db, users2)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("newUsers count: %d\n", len(newUsers))
	fmt.Printf("existingUsers count: %d\n", len(existingUsers))

	allUsers, err := common.GetAll(db, &User{})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("allUsers: %d\n", len(allUsers))
}
