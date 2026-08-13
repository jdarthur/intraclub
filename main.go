package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"
	"intraclub/route"
	"intraclub/route/draft"
	"intraclub/route/format"
	"intraclub/route/organization"
	"intraclub/route/ruleset"
	"intraclub/route/schedule"
	"intraclub/route/team"
	"intraclub/route/user"
	"intraclub/route/week"

	"github.com/gin-gonic/gin"
)

func main() {
	database.SysAdminCheck = model.IsUserSystemAdministrator
	api.UserType = &model.User{}

	// parse command-line flags
	cfg := parseFlags()

	// construct, open, and migrate the requested database provider
	db, err := database.NewProvider(context.Background(), cfg.providerConfig())
	if err != nil {
		log.Fatalf("failed to initialize database provider: %v", err)
	}

	// seed data for development mode
	if model.UseDevTokenMode {
		model.SeedDevData(db)
	}

	// generate or load JWT key pair
	err = api.GenerateJwtKeyPairIfNotExists()
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

	organizations := api.NewCrudCommon(model.NewOrganization, false, db)
	organizations.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)
	organization.RegisterRoutes(rg, db)

	rg.GET("/score_counting_types", model.GetScoreCountingTypes)
	scoringStructures := api.NewCrudCommon(model.NewScoringStructure, false, db)
	scoringStructures.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	ratings := api.NewCrudCommon(model.NewRating, false, db)
	ratings.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	formats := api.NewCrudCommon(model.NewFormat, false, db)
	formats.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	formatRatings := api.RouteFamily[*format.SetPossibleRatingsBody]{DatabaseProvider: db}
	formatRatings.Handle(rg, format.GetPossibleRatings{}, format.SetPossibleRatings{})

	rulesets := api.NewCrudCommon(model.NewRuleset, false, db)
	rulesets.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	amendRuleset := api.RouteFamily[*ruleset.AmendNameBody]{DatabaseProvider: db}
	amendRuleset.Handle(rg, ruleset.AmendRulesetName{})

	rg.GET("/draft_order_patterns", model.GetDraftOrderPatterns)

	draft.RegisterRoutes(rg, db)

	// Read-only REST surface for Seasons and their drafted rosters (used by
	// /seasons/[id] and the draft finalize flow). Rosters are reconstructed from
	// the public draft/season join tables (season_team, team_rating) plus the
	// draft's captains, since Team/TeamAssignment records are restricted to team
	// members only.
	seasons := api.NewCrudCommon(model.NewSeason, false, db)
	seasons.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)
	seasonTeams := api.NewCrudCommon(func() *model.SeasonTeam { return &model.SeasonTeam{} }, false, db)
	seasonTeams.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)
	teamRatings := api.NewCrudCommon(func() *model.TeamRating { return &model.TeamRating{} }, false, db)
	teamRatings.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)

	// Teams are largely immutable after the draft finalizes: they are exposed
	// only through the constrained read + role-assignment surface in
	// route/team (no generic create/update/delete on the raw records). Roster
	// reads are restricted to team members, sysadmins, and season
	// commissioners; only a team's captain / co-captains can assign roles.
	team.RegisterRoutes(rg, db)

	week.RegisterRoutes(rg, db)

	schedule.RegisterRoutes(rg, db)

	playoffStructures := api.NewCrudCommon(model.NewPlayoffStructure, false, db)
	playoffStructures.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	photos := api.NewCrudCommon(model.NewPhoto, false, db)
	photos.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	err = r.Run(cfg.addr)
	if err != nil {
		panic(err)
	}
}

// serverConfig holds the settings parsed from command-line flags and
// environment variables that configure the running server.
type serverConfig struct {
	addr   string
	dbKind database.ProviderKind
	dbPath string
}

// providerConfig converts the server's parsed database settings into the
// database.ProviderConfig consumed by database.NewProvider.
func (c serverConfig) providerConfig() database.ProviderConfig {
	return database.ProviderConfig{Kind: c.dbKind, Path: c.dbPath}
}

func parseFlags() serverConfig {
	useDevTokenMode := flag.Bool("dev-token", false, "Use development token mode")
	addr := flag.String("addr", "127.0.0.1:8080", "Address to listen on, e.g. 127.0.0.1:8080")
	jwtLifetimeFlag := flag.String("jwt-lifetime", "", "JWT token lifetime (e.g. 2h, 90m, 5s); falls back to INTRACLUB_JWT_LIFETIME env, then 2h")

	// defaultDBKind is the provider used at startup. SQLite is the default
	// (file-backed, single-file .db); pass --db memory for an ephemeral run.
	defaultDBKind := database.ProviderSqlite
	dbKind := flag.String("db", string(defaultDBKind), "Database provider (memory | sqlite)")
	dbPath := flag.String("db-path", "", "Path to the SQLite database file; falls back to INTRACLUB_DB_PATH env")
	flag.Parse()

	jwtLifetime, err := resolveJwtLifetime(*jwtLifetimeFlag)
	if err != nil {
		log.Fatalf("invalid JWT lifetime: %v", err)
	}
	api.JwtLifetime = jwtLifetime

	if useDevTokenMode != nil && *useDevTokenMode == true {
		model.UseDevTokenMode = true
		fmt.Println("Using development token mode")
	}

	if model.UseDevTokenMode && !isLoopbackAddress(*addr) {
		log.Fatalf("--dev-token mode is DEV MODE ONLY and bypasses authentication; it may only be used when the server is bound to a loopback address (127.0.0.1 / localhost), but got %q", *addr)
	}

	// resolve the database path from the flag or the environment variable
	return serverConfig{
		addr:   *addr,
		dbKind: database.ProviderKind(*dbKind),
		dbPath: resolveDBPath(*dbPath),
	}
}

// resolveDBPath returns the SQLite database file path, preferring the
// explicit --db-path flag and falling back to the INTRACLUB_DB_PATH env var.
func resolveDBPath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}
	return os.Getenv("INTRACLUB_DB_PATH")
}

// resolveJwtLifetime returns the JWT token lifetime, preferring the explicit
// --jwt-lifetime flag and falling back to the INTRACLUB_JWT_LIFETIME env var,
// then to the package default of api.JwtLifetime.
func resolveJwtLifetime(flagVal string) (time.Duration, error) {
	raw := flagVal
	if raw == "" {
		raw = os.Getenv("INTRACLUB_JWT_LIFETIME")
	}
	if raw == "" {
		return api.JwtLifetime, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid JWT lifetime %q: %w", raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("JWT lifetime must be positive, got %q", raw)
	}
	return d, nil
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
