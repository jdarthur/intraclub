package test

import (
	"intraclub/common"
	"intraclub/model"
	"time"
)

func createWeek() *model.Week {
	week := &model.Week{
		Date:         time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		OriginalDate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	v, err := common.Create(common.GlobalDbProvider, week)
	if err != nil {
		panic(err)
	}

	return v.(*model.Week)
}
