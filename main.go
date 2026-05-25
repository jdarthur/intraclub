package main

import (
	"flag"
	"fmt"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"
	"intraclub/route"
	"intraclub/route/user"

	"github.com/gin-gonic/gin"
)

func main() {
	database.SysAdminCheck = model.IsUserSystemAdministrator
	api.UserType = &model.User{}

	// set up the default database provider
	db := database.NewUnitTestDBProvider()

	// parse command-line flags
	parseFlags()

	// seed data for development mode
	if model.UseDevTokenMode {
		model.SeedDevData(db)
	}

	// generate or load JWT key pair
	err := api.GenerateJwtKeyPairIfNotExists()
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	rg := r.Group("/api")

	// noAuth for self-register and verify-email
	verifyEmail := api.RouteFamily[*user.VerifyEmailBody]{NoAuth: true, DatabaseProvider: db}
	verifyEmail.Handle(rg, user.VerifyEmail{})

	createUser := api.RouteFamily[*model.User]{NoAuth: true, DatabaseProvider: db}
	createUser.Handle(rg, user.SelfRegister{})

	whoAmI := api.RouteFamily[*model.User]{DatabaseProvider: db}
	whoAmI.Handle(rg, user.WhoAmI{})

	importHandler := &route.CsvImportHandler{DatabaseProvider: db}
	rg.Handle(api.HttpMethodPost.String(), "/import_users_from_csv", importHandler.HandleCsvImport)

	startTokenMgr := &model.StartLoginTokenManager{DatabaseProvider: db}
	rg.POST("/one_time_password", startTokenMgr.OneTimePassword)
	rg.POST("/token", startTokenMgr.CreateJwtFromOneTimePassword)

	// no auth for get user by ID / get all users functions

	getUsers := api.NewCrudCommon(model.NewUser, true, db)
	getUsers.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)

	// use auth for user deletion / update endpoints
	updateOrDeleteUsers := api.NewCrudCommon(model.NewUser, false, db)
	updateOrDeleteUsers.HandleRouteTypes(rg, api.CrudWrapperFunctionDelete, api.CrudWrapperFunctionUpdate)

	facilities := api.NewCrudCommon(model.NewFacility, false, db)
	facilities.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	rg.GET("/score_counting_types", model.GetScoreCountingTypes)
	scoringStructures := api.NewCrudCommon(model.NewScoringStructure, false, db)
	scoringStructures.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	ratings := api.NewCrudCommon(model.NewRating, false, db)
	ratings.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	formats := api.NewCrudCommon(model.NewFormat, false, db)
	formats.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	rg.GET("/draft_order_patterns", model.GetDraftOrderPatterns)

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
