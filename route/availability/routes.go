package availability

import (
	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up the Availability REST surface.
//
// The generic CRUD routes cover single-record reads, create, update, and delete
// (get-many is deliberately not registered because GET /availability is the
// per-user query). Custom routes add the participant-facing "set my
// availability for a week" upsert (POST /availability/set) and the per-user /
// per-team availability queries. Access control comes from the Availability
// model: a record is only editable by its owner (the participant who set it)
// and accessible to the owner's team members.
func RegisterRoutes(e *gin.RouterGroup, db database.Provider) {
	// Generic create/update/delete/get-one on the Availability records. The
	// create route auto-assigns ownership to the requesting user; the model's
	// per-user+week uniqueness prevents duplicate records.
	crud := api.NewCrudCommon(model.NewAvailability, false, db)
	crud.HandleRouteTypes(e,
		api.CrudWrapperFunctionGetOne,
		api.CrudWrapperFunctionCreate,
		api.CrudWrapperFunctionUpdate,
		api.CrudWrapperFunctionDelete,
	)

	setFamily := api.RouteFamily[*SetAvailabilityBody]{DatabaseProvider: db}
	setFamily.Handle(e, SetAvailability{})

	queryFamily := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	queryFamily.Handle(e, GetForUser{}, GetForTeam{})
}
