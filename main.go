package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"
	"intraclub/route"
	"intraclub/route/availability"
	"intraclub/route/blurb"
	"intraclub/route/comment"
	"intraclub/route/draft"
	"intraclub/route/format"
	"intraclub/route/lateaddition"
	"intraclub/route/lineup"
	"intraclub/route/match"
	"intraclub/route/organization"
	"intraclub/route/proposal"
	"intraclub/route/ruleset"
	"intraclub/route/schedule"
	"intraclub/route/scoringstructure"
	"intraclub/route/seasoncommissioner"
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

	// slow mode injects artificial latency into every request to simulate a
	// slow / high-RTT connection between the client and the API (see
	// resolveSlowMode / resolveSlowModeLatency for flag + env configuration).
	if cfg.slowMode {
		log.Printf("slow mode enabled: injecting %s of artificial latency into every API request", cfg.slowModeLatency)
		r.Use(slowModeMiddleware(cfg.slowModeLatency))
	}

	rg := r.Group("/api")

	// noAuth for self-register and verify-email
	verifyEmail := api.RouteFamily[*user.VerifyEmailBody]{NoAuth: true, DatabaseProvider: db}
	verifyEmail.Handle(rg, user.VerifyEmail{})

	createUser := api.RouteFamily[*model.User]{NoAuth: true, DatabaseProvider: db}
	createUser.Handle(rg, user.SelfRegister{})

	whoAmI := api.RouteFamily[*model.User]{DatabaseProvider: db}
	whoAmI.Handle(rg, user.WhoAmI{})

	// whoami/roles lists the current user's role names; the UI uses it to gate
	// sysadmin-only controls (e.g. season late additions).
	roles := api.RouteFamily[*model.User]{DatabaseProvider: db}
	roles.Handle(rg, user.Roles{})

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

	// ScoringStructureSecondary records are the join table that links a
	// (composite) ScoringStructure to its secondary (tie-breaker) scoring
	// structures, ordered by SecondaryIndex. The raw collection is exposed via
	// generic CRUD for read/admin access, but the ordered list is managed from
	// the scoring structure detail page through the get/set routes below, which
	// enforce the primary structure's edit authorization (owner / sysadmin) and
	// replace the full ordered list atomically (preserving SecondaryIndex
	// ordering).
	scoringStructureSecondaries := api.NewCrudCommon(func() *model.ScoringStructureSecondary { return &model.ScoringStructureSecondary{} }, false, db)
	scoringStructureSecondaries.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	scoringStructureSecondaryRoutes := api.RouteFamily[*scoringstructure.SetSecondaryScoringStructuresBody]{DatabaseProvider: db}
	scoringStructureSecondaryRoutes.Handle(rg, scoringstructure.GetSecondaryScoringStructures{}, scoringstructure.SetSecondaryScoringStructures{})

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

	// RuleSections are the individual rule blocks (title + markdown) that make
	// up a ruleset; RulesetSections are the join-table records that order them
	// within a ruleset via SectionIndex. Both are exposed via generic CRUD for
	// read/admin access, but section edits from the UI go through the
	// amend_sections route below (which applies a RuleAmendment through the
	// Ruleset.Amend flow, producing new revisions and preserving ordering).
	ruleSections := api.NewCrudCommon(func() *model.RuleSection { return &model.RuleSection{} }, false, db)
	ruleSections.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	rulesetSections := api.NewCrudCommon(func() *model.RulesetSection { return &model.RulesetSection{} }, false, db)
	rulesetSections.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	amendRulesetSections := api.RouteFamily[*model.RuleAmendment]{DatabaseProvider: db}
	amendRulesetSections.Handle(rg, ruleset.AmendSections{})

	// Commissioner proposals are the "manage club rules" feature: a season
	// commissioner proposes a rule change / administrative action and the
	// season's participants (commissioners + team captains) ratify it by
	// majority or unanimous consent. Generic CRUD covers create / list / read /
	// update / delete of proposals; writes to votes go exclusively through the
	// custom cast-vote endpoint in route/proposal (which validates the voter is
	// a season participant), so the vote records are exposed read-only here.
	proposals := api.NewCrudCommon(model.NewCommissionerProposal, false, db)
	proposals.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)
	proposalVotes := api.NewCrudCommon(func() *model.CommissionerProposalVote { return &model.CommissionerProposalVote{} }, false, db)
	proposalVotes.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)
	proposal.RegisterRoutes(rg, db)

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

	// SeasonLateAdditions link a Season to Users added after the draft was
	// completed. The generic surface is read-only (the season page shows them);
	// writes go exclusively through the custom routes in route/lateaddition,
	// which enforce the model's sysadmin-only EditableBy constraint.
	seasonLateAdditions := api.NewCrudCommon(func() *model.SeasonLateAddition { return &model.SeasonLateAddition{} }, false, db)
	seasonLateAdditions.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)
	lateAdditions := api.RouteFamily[*lateaddition.AddLateAdditionBody]{DatabaseProvider: db}
	lateAdditions.Handle(rg, lateaddition.AddLateAddition{}, lateaddition.RemoveLateAddition{})

	// SeasonCommissioners link a Season to its commissioner users
	// (co-commissioners). The generic surface is read-only (the season page
	// shows them); writes go exclusively through the custom routes in
	// route/seasoncommissioner, which enforce the model's sysadmin-only
	// EditableBy constraint.
	seasonCommissioners := api.NewCrudCommon(func() *model.SeasonCommissioner { return &model.SeasonCommissioner{} }, false, db)
	seasonCommissioners.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)
	commissioners := api.RouteFamily[*seasoncommissioner.AddCommissionerBody]{DatabaseProvider: db}
	commissioners.Handle(rg, seasoncommissioner.AddCommissioner{}, seasoncommissioner.RemoveCommissioner{})

	// Teams are largely immutable after the draft finalizes: they are exposed
	// only through the constrained read + role-assignment surface in
	// route/team (no generic create/update/delete on the raw records). Roster
	// reads are restricted to team members, sysadmins, and season
	// commissioners; only a team's captain / co-captains can assign roles.
	team.RegisterRoutes(rg, db)

	week.RegisterRoutes(rg, db)

	availability.RegisterRoutes(rg, db)

	schedule.RegisterRoutes(rg, db)

	lineup.RegisterRoutes(rg, db)

	// Match scoring is driven through the custom route/match surface (generate,
	// score, complete, week score sheet, standings); the underlying records are
	// also exposed read-only here so the season page can render a score sheet
	// without the generic write surface (writes are gated by each model's
	// EditableBy — team matches are sysadmin-only, individual matches are
	// editable by their match_editor rows).
	match.RegisterRoutes(rg, db)

	individualMatches := api.NewCrudCommon(model.NewMatch, false, db)
	individualMatches.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)

	teamMatches := api.NewCrudCommon(func() *model.TeamMatch { return &model.TeamMatch{} }, false, db)
	teamMatches.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)

	teamMatchIndividualMatches := api.NewCrudCommon(func() *model.TeamMatchIndividualMatch { return &model.TeamMatchIndividualMatch{} }, false, db)
	teamMatchIndividualMatches.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)

	matchEditors := api.NewCrudCommon(func() *model.MatchEditor { return &model.MatchEditor{} }, false, db)
	matchEditors.HandleRouteTypes(rg, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)

	playoffStructures := api.NewCrudCommon(model.NewPlayoffStructure, false, db)
	playoffStructures.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	photos := api.NewCrudCommon(model.NewPhoto, false, db)
	photos.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	// Blurbs are season-scoped posts (news/announcements) that users can
	// comment on and react to. Generic CRUD covers create / list / read /
	// update / delete (an owner creates and edits their own blurb); reactions
	// and photo attachment go through the custom routes in route/blurb, which
	// enforce season-participation and ownership rules.
	blurbs := api.NewCrudCommon(model.NewBlurb, false, db)
	blurbs.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	// Comments attach to a blurb and support replies; reactions attach to
	// comments. Generic CRUD covers create / list / read / update / delete
	// (edits are gated by the model's EditableBy: owner / blurb owner /
	// season commissioners / sysadmin); reactions go through the custom routes
	// in route/comment.
	comments := api.NewCrudCommon(model.NewComment, false, db)
	comments.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	// BlurbPhoto / BlurbReaction / CommentReaction are child-table join rows.
	// They are exposed via generic CRUD for read/admin access; the interactive
	// writes go through the custom routes in route/blurb and route/comment
	// (which enforce ownership, dedup, and season-participation).
	blurbPhotos := api.NewCrudCommon(func() *model.BlurbPhoto { return &model.BlurbPhoto{} }, false, db)
	blurbPhotos.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	blurbReactions := api.NewCrudCommon(func() *model.BlurbReaction { return &model.BlurbReaction{} }, false, db)
	blurbReactions.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	commentReactions := api.NewCrudCommon(func() *model.CommentReaction { return &model.CommentReaction{} }, false, db)
	commentReactions.HandleRouteTypes(rg, api.CrudWrapperFunctionAll...)

	blurbReactionRoutes := api.RouteFamily[*blurb.ReactionBody]{DatabaseProvider: db}
	blurbReactionRoutes.Handle(rg, blurb.React{}, blurb.Unreact{})

	blurbPhotoRoutes := api.RouteFamily[*blurb.PhotoBody]{DatabaseProvider: db}
	blurbPhotoRoutes.Handle(rg, blurb.AddPhoto{}, blurb.RemovePhoto{})

	commentReactionRoutes := api.RouteFamily[*comment.ReactionBody]{DatabaseProvider: db}
	commentReactionRoutes.Handle(rg, comment.React{}, comment.Unreact{})

	err = r.Run(cfg.addr)
	if err != nil {
		panic(err)
	}
}

// serverConfig holds the settings parsed from command-line flags and
// environment variables that configure the running server.
type serverConfig struct {
	addr            string
	dbKind          database.ProviderKind
	dbPath          string
	slowMode        bool
	slowModeLatency time.Duration
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
	slowMode := flag.Bool("slow-mode", false, "Inject artificial latency into every API request to simulate a slow / high-RTT connection; falls back to INTRACLUB_SLOW_MODE env")
	slowModeLatency := flag.Duration("slow-mode-latency", defaultSlowModeLatency, "Per-request artificial latency to inject when slow mode is enabled; falls back to INTRACLUB_SLOW_MODE_LATENCY env")
	flag.Parse()

	jwtLifetime, err := resolveJwtLifetime(*jwtLifetimeFlag)
	if err != nil {
		log.Fatalf("invalid JWT lifetime: %v", err)
	}
	api.JwtLifetime = jwtLifetime

	slowModeEnabled, err := resolveSlowMode(flagWasSet("slow-mode"), *slowMode)
	if err != nil {
		log.Fatalf("invalid slow mode setting: %v", err)
	}
	latency, err := resolveSlowModeLatency(flagWasSet("slow-mode-latency"), *slowModeLatency)
	if err != nil {
		log.Fatalf("invalid slow mode latency: %v", err)
	}

	if useDevTokenMode != nil && *useDevTokenMode == true {
		model.UseDevTokenMode = true
		fmt.Println("Using development token mode")
	}

	if model.UseDevTokenMode && !isLoopbackAddress(*addr) {
		log.Fatalf("--dev-token mode is DEV MODE ONLY and bypasses authentication; it may only be used when the server is bound to a loopback address (127.0.0.1 / localhost), but got %q", *addr)
	}

	// resolve the database path from the flag or the environment variable
	return serverConfig{
		addr:            *addr,
		dbKind:          database.ProviderKind(*dbKind),
		dbPath:          resolveDBPath(*dbPath),
		slowMode:        slowModeEnabled,
		slowModeLatency: latency,
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

// defaultSlowModeLatency is the per-request artificial latency injected when
// slow mode is enabled and no latency is configured via flag or env var.
const defaultSlowModeLatency = 500 * time.Millisecond

// flagWasSet reports whether the named flag was explicitly provided on the
// command line (as opposed to being left at its default value). It must be
// called after flag.Parse.
func flagWasSet(name string) bool {
	set := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}

// resolveSlowMode returns whether artificial latency should be injected into
// API requests, preferring the explicit --slow-mode flag and falling back to
// the INTRACLUB_SLOW_MODE env var, then the default (disabled).
func resolveSlowMode(flagSet, flagVal bool) (bool, error) {
	if flagSet {
		return flagVal, nil
	}
	raw := os.Getenv("INTRACLUB_SLOW_MODE")
	if raw == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("invalid INTRACLUB_SLOW_MODE %q: %w", raw, err)
	}
	return v, nil
}

// resolveSlowModeLatency returns the per-request artificial latency, preferring
// the explicit --slow-mode-latency flag and falling back to the
// INTRACLUB_SLOW_MODE_LATENCY env var, then the default.
func resolveSlowModeLatency(flagSet bool, flagVal time.Duration) (time.Duration, error) {
	if flagSet {
		return flagVal, nil
	}
	raw := os.Getenv("INTRACLUB_SLOW_MODE_LATENCY")
	if raw == "" {
		return defaultSlowModeLatency, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid INTRACLUB_SLOW_MODE_LATENCY %q: %w", raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("slow mode latency must be positive, got %q", raw)
	}
	return d, nil
}

// slowModeMiddleware injects artificial latency into every request to simulate
// a slow / high-RTT connection between the client and the API. It sleeps before
// the handler chain runs, so the added delay is reflected in each request's
// total latency.
func slowModeMiddleware(delay time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if delay > 0 {
			time.Sleep(delay)
		}
		c.Next()
	}
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
