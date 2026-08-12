package organization

import (
	"intraclub/api"
	"intraclub/database"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up the owner-managed Organization membership REST
// surface. The Organization CRUD routes themselves are registered directly in
// main.go via api.NewCrudCommon(model.NewOrganization, false, db) (mirroring
// the facilities wiring); this function adds the custom membership endpoints.
//
// Each route is registered on its own RouteFamily because the routes use
// different request-body types (ListMembers/RemoveMember take no body while
// AddMember takes an AddMemberBody).
func RegisterRoutes(e *gin.RouterGroup, db database.Provider) {
	listFamily := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	listFamily.Handle(e, ListMembers{})

	addFamily := api.RouteFamily[*AddMemberBody]{DatabaseProvider: db}
	addFamily.Handle(e, AddMember{})

	removeFamily := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	removeFamily.Handle(e, RemoveMember{})
}
