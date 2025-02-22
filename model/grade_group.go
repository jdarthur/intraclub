package model

import "intraclub/common"

type GradeGroup struct {
	ID     common.RecordId   // unique ID for this GradeGroup
	Grades []common.RecordId // list of Grade IDs
}
