package schedule

import (
	"context"
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base path for the Schedule REST surface. It matches the
// singular route convention used by the generic CRUD routes (derived from
// Schedule.Type()).
const BaseRoute = "/schedule"

// isSeasonCommissioner reports whether the requesting user is one of the
// season's commissioners (the only role allowed to create / modify the
// schedule and its weekly matchups). Everyone else is view-only.
func isSeasonCommissioner[T database.Validatable](req api.Request[T], seasonId model.SeasonId) bool {
	if req.Token == nil {
		return false
	}
	for _, uid := range model.EditableBySeason(req.Context, req.DatabaseProvider, seasonId) {
		if uid == req.Token.UserId {
			return true
		}
	}
	return false
}

// CreateScheduleBody is the request body for CreateSchedule.
type CreateScheduleBody struct {
	// SeasonId is the season this schedule belongs to. Only one schedule may
	// exist per season (enforced by Schedule.UniquenessEquivalent).
	SeasonId model.SeasonId `json:"season_id"`
}

// StaticallyValid ensures a season is specified before any handler logic runs.
func (b *CreateScheduleBody) StaticallyValid() error {
	if b.SeasonId.RecordId() == database.InvalidRecordId {
		return errors.New("season_id must be set")
	}
	return nil
}

// CreateSchedule creates a Schedule for a Season. Only a season commissioner
// may create the schedule; other participants may only view it.
type CreateSchedule struct{}

func (c CreateSchedule) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute
}

func (c CreateSchedule) RequestBody() (*CreateScheduleBody, bool) {
	return &CreateScheduleBody{}, true
}

func (c CreateSchedule) Handler(req api.Request[*CreateScheduleBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	// The season must already exist before a schedule can reference it.
	if _, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Season{}, req.Body.SeasonId.RecordId()); err != nil {
		return nil, http.StatusBadRequest, err
	}

	if !isSeasonCommissioner(req, req.Body.SeasonId) {
		return nil, http.StatusForbidden, errors.New("only a season commissioner may create a schedule")
	}

	schedule := model.NewSchedule()
	schedule.SeasonId = req.Body.SeasonId
	// One schedule per season is enforced by Schedule.UniquenessEquivalent,
	// which CreateOne checks against existing schedules.
	created, err := database.CreateOne(req.Context, req.DatabaseProvider, schedule)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: created}, http.StatusOK, nil
}

// WeeklyMatchupDTO is the wire representation of a weekly matchup plus its
// assigned team matchups (home/away/bye entries) for a schedule.
type WeeklyMatchupDTO struct {
	ID       model.WeeklyMatchupId `json:"id"`
	WeekId   model.WeekId          `json:"week_id"`
	SeasonId model.SeasonId        `json:"season_id"`
	Matchups []*model.TeamMatchup  `json:"matchups"`
}

// ScheduleDetail is the wire representation of a Schedule with its ordered
// weekly matchups and the season's commissioners (so the UI can determine
// whether the current user may edit).
type ScheduleDetail struct {
	Schedule      *model.Schedule      `json:"schedule"`
	Commissioners []database.UserId    `json:"commissioners"`
	WeeklyMatchups []*WeeklyMatchupDTO `json:"weekly_matchups"`
}

// buildWeeklyMatchupDTO converts a stored WeeklyMatchup into its wire form.
func buildWeeklyMatchupDTO(ctx context.Context, db database.Provider, wm *model.WeeklyMatchup) (*WeeklyMatchupDTO, error) {
	matchups, err := wm.GetMatchups(ctx, db)
	if err != nil {
		return nil, err
	}
	return &WeeklyMatchupDTO{
		ID:       wm.ID,
		WeekId:   wm.WeekId,
		SeasonId: wm.SeasonId,
		Matchups: matchups,
	}, nil
}

// scheduleDetailForSeason builds the ScheduleDetail for a season. If the
// season has no schedule yet, Schedule is nil (so the UI can still learn the
// commissioners and offer to create one).
func scheduleDetailForSeason(ctx context.Context, db database.Provider, season *model.Season) (*ScheduleDetail, error) {
	commissioners, err := season.GetCommissioners(ctx, db)
	if err != nil {
		return nil, err
	}

	detail := &ScheduleDetail{Commissioners: commissioners, WeeklyMatchups: []*WeeklyMatchupDTO{}}

	if season.ScheduleID.RecordId() == database.InvalidRecordId {
		return detail, nil
	}

	schedule, err := database.GetExistingRecordById(ctx, db, &model.Schedule{}, season.ScheduleID.RecordId())
	if err != nil {
		return nil, err
	}
	detail.Schedule = schedule

	matchups, err := schedule.GetMatchups(ctx, db)
	if err != nil {
		return nil, err
	}
	for _, wm := range matchups {
		// Weekly matchups may exist even if their team entries haven't been
		// filled in yet; GetMatchups on the schedule resolves them lazily.
		dto, err := buildWeeklyMatchupDTO(ctx, db, wm)
		if err != nil {
			return nil, err
		}
		detail.WeeklyMatchups = append(detail.WeeklyMatchups, dto)
	}
	return detail, nil
}

// ScheduleQuery holds the optional query parameters for ListSchedules.
type ScheduleQuery struct {
	// SeasonId, when set, restricts the response to a single ScheduleDetail
	// for that season (Schedule may be null if no schedule exists yet).
	SeasonId model.SeasonId `json:"season_id"`
}

// StaticallyValid has no static constraints.
func (q *ScheduleQuery) StaticallyValid() error { return nil }

// ListSchedules returns all schedules, or a single ScheduleDetail for a season
// when the `season_id` query parameter is provided. Schedules are viewable by
// everyone.
type ListSchedules struct{}

func (c ListSchedules) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, BaseRoute
}

func (c ListSchedules) RequestBody() (*ScheduleQuery, bool) {
	return &ScheduleQuery{}, false
}

func (c ListSchedules) Handler(req api.Request[*ScheduleQuery]) (any, int, error) {
	query := req.HTTPRequest().URL.Query()

	if seasonStr := query.Get("season_id"); seasonStr != "" {
		seasonId, err := database.RecordIdFromString(seasonStr)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		season, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Season{}, seasonId)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		detail, err := scheduleDetailForSeason(req.Context, req.DatabaseProvider, season)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		return gin.H{api.ResourceKey: detail}, http.StatusOK, nil
	}

	schedules, err := database.GetAll[*model.Schedule](req.Context, req.DatabaseProvider)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	details := make([]*ScheduleDetail, 0, len(schedules))
	for _, sched := range schedules {
		season, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Season{}, sched.SeasonId.RecordId())
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		detail, err := scheduleDetailForSeason(req.Context, req.DatabaseProvider, season)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		details = append(details, detail)
	}
	return gin.H{api.ResourceKey: details}, http.StatusOK, nil
}

// GetScheduleDetail returns the ScheduleDetail for a specific schedule. It is
// viewable by everyone.
type GetScheduleDetail struct{}

func (c GetScheduleDetail) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, api.AppendPathId(BaseRoute)
}

func (c GetScheduleDetail) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c GetScheduleDetail) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	schedule, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Schedule{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	season, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Season{}, schedule.SeasonId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	detail, err := scheduleDetailForSeason(req.Context, req.DatabaseProvider, season)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: detail}, http.StatusOK, nil
}

// AssignWeeklyMatchupBody is the request body for AssignWeeklyMatchup.
type AssignWeeklyMatchupBody struct {
	// WeekId is the week of the season this weekly matchup corresponds to.
	WeekId model.WeekId `json:"week_id"`
	// Matchups are the home/away/bye entries for the week. Together they must
	// cover every team in the season exactly once (validated by the model).
	Matchups []*model.TeamMatchup `json:"matchups"`
}

// StaticallyValid ensures a week is specified.
func (b *AssignWeeklyMatchupBody) StaticallyValid() error {
	if b.WeekId.RecordId() == database.InvalidRecordId {
		return errors.New("week_id must be set")
	}
	return nil
}

// AssignWeeklyMatchup creates or replaces the weekly matchup for a single week
// of a season's schedule (home/away/bye entries), and ensures it is assigned to
// the schedule. Only a season commissioner may modify the schedule.
type AssignWeeklyMatchup struct{}

func (c AssignWeeklyMatchup) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/weekly_matchup"
}

func (c AssignWeeklyMatchup) RequestBody() (*AssignWeeklyMatchupBody, bool) {
	return &AssignWeeklyMatchupBody{}, true
}

func (c AssignWeeklyMatchup) Handler(req api.Request[*AssignWeeklyMatchupBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	schedule, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Schedule{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if !isSeasonCommissioner(req, schedule.SeasonId) {
		return nil, http.StatusForbidden, errors.New("only a season commissioner may modify the schedule")
	}

	season, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Season{}, schedule.SeasonId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	week, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Week{}, req.Body.WeekId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	// The week must belong to this season's draft.
	if week.DraftId != season.DraftId {
		return nil, http.StatusBadRequest, errors.New("week does not belong to this season")
	}

	// Find (or create) the WeeklyMatchup for this season + week.
	existing, err := database.GetAllWhere[*model.WeeklyMatchup](req.Context, req.DatabaseProvider, func(_ context.Context, wm *model.WeeklyMatchup) bool {
		return wm.SeasonId == season.ID && wm.WeekId == week.ID
	})
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	var wm *model.WeeklyMatchup
	if len(existing) > 0 {
		wm = existing[0]
	} else {
		wm = model.NewWeeklyMatchup()
		wm.WeekId = week.ID
		wm.SeasonId = season.ID
		wm, err = database.CreateOne(req.Context, req.DatabaseProvider, wm)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
	}

	// Capture the previous entries so we can restore them if the new ones fail
	// validation (the model enforces no double-booking and that every team has
	// exactly one matchup or bye).
	previous, err := wm.GetMatchups(req.Context, req.DatabaseProvider)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err := wm.SetMatchups(req.Context, req.DatabaseProvider, req.Body.Matchups); err != nil {
		return nil, http.StatusBadRequest, err
	}
	if err := wm.DynamicallyValid(req.Context, req.DatabaseProvider); err != nil {
		_ = wm.SetMatchups(req.Context, req.DatabaseProvider, previous)
		return nil, http.StatusBadRequest, err
	}

	// Ensure this weekly matchup is assigned to the schedule (position = append).
	if err := ensureScheduleMatchup(req.Context, req.DatabaseProvider, schedule, wm.ID); err != nil {
		return nil, http.StatusBadRequest, err
	}

	dto, err := buildWeeklyMatchupDTO(req.Context, req.DatabaseProvider, wm)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: dto}, http.StatusOK, nil
}

// ensureScheduleMatchup adds a ScheduleMatchup join (schedule <-> weekly
// matchup) if one doesn't already exist, appending at the end of the schedule.
func ensureScheduleMatchup(ctx context.Context, db database.Provider, schedule *model.Schedule, wmId model.WeeklyMatchupId) error {
	existing, err := database.GetAllWhere[*model.ScheduleMatchup](ctx, db, func(_ context.Context, sm *model.ScheduleMatchup) bool {
		return sm.ScheduleId == schedule.ID
	})
	if err != nil {
		return err
	}
	for _, sm := range existing {
		if sm.WeeklyMatchupId == wmId {
			return nil
		}
	}
	position := len(existing)
	sm := model.NewScheduleMatchup(schedule.ID, wmId, position)
	_, err = database.CreateOne(ctx, db, sm)
	return err
}

// EmptyBody is a placeholder request body for routes that take no body.
type EmptyBody struct{}

// StaticallyValid has no static constraints.
func (b *EmptyBody) StaticallyValid() error { return nil }
