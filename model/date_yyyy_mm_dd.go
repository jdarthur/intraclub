package model

import (
	"fmt"
	"strings"
	"time"
)

type YyyyMmDdDate struct {
	time.Time
}

const yyyyMmDdLayout = "2006-01-02"

func (y *YyyyMmDdDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		y.Time = time.Time{}
		return
	}
	y.Time, err = time.Parse(yyyyMmDdLayout, s)
	return
}

func (y *YyyyMmDdDate) MarshalJSON() ([]byte, error) {
	f := fmt.Sprintf("\"%s\"", y.Time.Format(yyyyMmDdLayout))
	return []byte(f), nil
}
