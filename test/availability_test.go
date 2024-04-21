package test

import (
	"intraclub/common"
	"intraclub/model"
	"testing"
)

func init() {
	ResetDatabase()
}

func TestInvalidAvailabilityOption(t *testing.T) {
	a := &model.Availability{Available: 999}

	_, err := common.Create(common.GlobalDbProvider, a)
	if err == nil {
		t.Fatalf("expected error on invalid availability")
	} else {
		ValidateErrorContains(t, err, "unexpected availability option")
	}
}

func TestEmptyWeekId(t *testing.T) {
	a := &model.Availability{Available: model.Available}

	_, err := common.Create(common.GlobalDbProvider, a)
	if err == nil {
		t.Fatalf("expected error on invalid week ID")
	} else {
		ValidateErrorContains(t, err, "invalid object id")
	}
}

func TestInvalidWeekId(t *testing.T) {
	a := &model.Availability{Available: model.Available, WeekId: "test123456789"}

	_, err := common.Create(common.GlobalDbProvider, a)
	if err == nil {
		t.Fatalf("expected error on invalid week ID")
	} else {
		ValidateErrorContains(t, err, "invalid object id")
	}
}

func TestInvalidUserId(t *testing.T) {

	week := createWeek()

	a := &model.Availability{
		Available: model.Available,
		WeekId:    week.ID.Hex(),
	}

	_, err := common.Create(common.GlobalDbProvider, a)
	if err == nil {
		t.Fatalf("expected error on invalid user ID")
	} else {
		ValidateErrorContains(t, err, "invalid object id")
	}
}

func TestValidAvailability(t *testing.T) {

	week := createWeek()
	user := createUser()

	a := &model.Availability{
		Available: model.Available,
		WeekId:    week.ID.Hex(),
		UserId:    user.ID.Hex(),
	}

	_, err := common.Create(common.GlobalDbProvider, a)
	if err != nil {
		t.Fatal(err)
	}
}
