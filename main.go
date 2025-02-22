package main

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"intraclub/route"
)

func main() {
	common.SysAdminCheck = model.IsUserSystemAdministrator

	r := gin.Default()

	// noAuth for self-register
	createUser := common.RouteFamily[*model.User]{}
	createUser.Handle(r, route.SelfRegister{})

	// no auth for get user by ID / get all users functions

	getUsers := common.NewCrudCommon(model.NewUser, false)
	getUsers.HandleRouteTypes(r, common.CrudWrapperFunctionGetOne, common.CrudWrapperFunctionGetMany)

	// use auth for user deletion / update endpoints
	updateOrDeleteUsers := common.NewCrudCommon(model.NewUser, true)
	updateOrDeleteUsers.HandleRouteTypes(r, common.CrudWrapperFunctionDelete, common.CrudWrapperFunctionUpdate)

	// model.Facility endpoints
	route.FacilityEndpoints.HandleRouteTypes(r, common.CrudWrapperFunctionAll...)

}
