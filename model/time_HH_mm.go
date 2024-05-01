package model

import (
	"fmt"
	"strings"
	"time"
)

type hhMmTime struct {
	time.Time
}

const hhMmLayout = "15:04"

func (h *hhMmTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		h.Time = time.Time{}
		return
	}
	h.Time, err = time.Parse(hhMmLayout, s)
	return
}

func (h *hhMmTime) MarshalJSON() ([]byte, error) {
	f := fmt.Sprintf("\"%s\"", h.Time.Format(hhMmLayout))
	return []byte(f), nil
}
