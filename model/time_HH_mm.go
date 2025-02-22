package model

import (
	"fmt"
	"strings"
	"time"
)

type HhMmTime struct {
	time.Time
}

const hhMmLayout = "15:04"

func (h *HhMmTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		h.Time = time.Time{}
		return
	}
	h.Time, err = time.Parse(hhMmLayout, s)
	return
}

func (h *HhMmTime) MarshalJSON() ([]byte, error) {
	f := fmt.Sprintf("\"%s\"", h.Time.Format(hhMmLayout))
	return []byte(f), nil
}
