package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"intraclub/route"
)

func main() {
	common.SysAdminCheck = model.IsUserSystemAdministrator
	common.UserType = &model.User{}

	// set up the default database provider
	db := common.NewUnitTestDBProvider()

	// parse command-line flags
	parseFlags()

	// seed data for development mode
	if model.UseDevTokenMode {
		model.SeedDevData(db)
	}

	// generate or load JWT key pair
	err := common.GenerateJwtKeyPairIfNotExists()
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	api := r.Group("/api")

	// noAuth for self-register
	createUser := common.RouteFamily[*model.User]{DatabaseProvider: db}
	createUser.Handle(api, route.SelfRegister{})

	whoAmI := common.RouteFamily[*model.User]{UseAuth: true, DatabaseProvider: db}
	whoAmI.Handle(api, route.WhoAmI{})

	importHandler := &route.CsvImportHandler{DatabaseProvider: db}
	api.Handle(common.HttpMethodPost.String(), "/import_users_from_csv", importHandler.HandleCsvImport)

	startTokenMgr := &model.StartLoginTokenManager{DatabaseProvider: db}
	api.POST("/one_time_password", startTokenMgr.OneTimePassword)
	api.POST("/token", startTokenMgr.CreateJwtFromOneTimePassword)

	// no auth for get user by ID / get all users functions

	getUsers := common.NewCrudCommon(model.NewUser, false, db)
	getUsers.HandleRouteTypes(api, common.CrudWrapperFunctionGetOne, common.CrudWrapperFunctionGetMany)

	// use auth for user deletion / update endpoints
	updateOrDeleteUsers := common.NewCrudCommon(model.NewUser, true, db)
	updateOrDeleteUsers.HandleRouteTypes(api, common.CrudWrapperFunctionDelete, common.CrudWrapperFunctionUpdate)

	facilities := common.NewCrudCommon(model.NewFacility, true, db)
	facilities.HandleRouteTypes(api, common.CrudWrapperFunctionAll...)

	api.GET("/score_counting_types", model.GetScoreCountingTypes)
	scoringStructures := common.NewCrudCommon(model.NewScoringStructure, true, db)
	scoringStructures.HandleRouteTypes(api, common.CrudWrapperFunctionAll...)

	ratings := common.NewCrudCommon(model.NewRating, true, db)
	ratings.HandleRouteTypes(api, common.CrudWrapperFunctionAll...)

	formats := common.NewCrudCommon(model.NewFormat, true, db)
	formats.HandleRouteTypes(api, common.CrudWrapperFunctionAll...)

	api.GET("/draft_order_patterns", model.GetDraftOrderPatterns)

	seasonComposite := common.RouteFamily[*model.SeasonComposite]{UseAuth: true, DatabaseProvider: db}
	seasonComposite.Handle(api, route.GetMySeasons{})

	err = r.Run("127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
}

func parseFlags() {
	useDevTokenMode := flag.Bool("dev-token", false, "Use development token mode")
	flag.Parse()

	if useDevTokenMode != nil && *useDevTokenMode == true {
		model.UseDevTokenMode = true
		fmt.Println("Using development token mode")
	}
}
