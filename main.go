package main

import (
	"flag"
	"fmt"
	"log"
	"net"

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
	addr := parseFlags()

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

	err = r.Run(addr)
	if err != nil {
		panic(err)
	}
}

func parseFlags() string {
	useDevTokenMode := flag.Bool("dev-token", false, "Use development token mode")
	addr := flag.String("addr", "127.0.0.1:8080", "Address to listen on, e.g. 127.0.0.1:8080")
	flag.Parse()

	if useDevTokenMode != nil && *useDevTokenMode == true {
		model.UseDevTokenMode = true
		fmt.Println("Using development token mode")
	}

	if model.UseDevTokenMode && !isLoopbackAddress(*addr) {
		log.Fatalf("--dev-token mode is DEV MODE ONLY and bypasses authentication; it may only be used when the server is bound to a loopback address (127.0.0.1 / localhost), but got %q", *addr)
	}

	return *addr
}

// isLoopbackAddress reports whether addr's host is a loopback interface
// (127.0.0.1 / localhost). This guards dev mode, which bypasses auth gating.
func isLoopbackAddress(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// e.g. "localhost" or "127.0.0.1" with no port
		host = addr
	}

	// an empty host binds all interfaces (e.g. ":8080"), which is not loopback
	if host == "" {
		return false
	}

	return host == "localhost" || net.ParseIP(host).IsLoopback()
}
