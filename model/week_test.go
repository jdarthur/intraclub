package model

import (
	"intraclub/common"
	"testing"
	"time"
)

func newStoredWeek(t *testing.T, db common.DatabaseProvider, seasonId SeasonId) *Week {
	week := NewWeek()
	week.SeasonId = seasonId
	week.Date = time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC)
	v, err := common.CreateOne(db, week)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
