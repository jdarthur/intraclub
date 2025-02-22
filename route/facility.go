package route

import (
	"intraclub/common"
	"intraclub/model"
)

var FacilityEndpoints = common.NewCrudCommon(model.NewFacility, true)
