package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"intraclub/database"
)

var RatingOne = "Well-developed overall game, strong fundamentals, and skilled against many types of opponent play styles"
var RatingTwo = "Moderate overall game, perhaps lacking in some fundamentals but makes up for weaknesses through strengths such as finesse, quickness, or strategy"
var RatingThree = "Lower-skilled player who might be prone to mistakes or beatable due to lack of quickness or weakness to particular shot styles"

type RatingId database.RecordId

func (id *RatingId) UnmarshalJSON(bytes []byte) error {
	rid := id.RecordId()
	if err := (*database.RecordId)(&rid).UnmarshalJSON(bytes); err != nil {
		return err
	}
	*id = RatingId(rid)
	return nil
}

func (id RatingId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id RatingId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id RatingId) String() string {
	return id.RecordId().String()
}

type RatingList []RatingId

func (r *RatingList) UnmarshalJSON(bytes []byte) error {
	idList, err := database.UnmarshalStringIdList(bytes)
	if err != nil {
		return err
	}

	for _, id := range idList {
		*r = append(*r, RatingId(id))
	}
	return nil
}

type Rating struct {
	ID          RatingId        `json:"id"`
	UserId      database.UserId `json:"user_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
}

func (r *Rating) UniquenessEquivalent(other *Rating) error {
	if r.Name == other.Name {
		return fmt.Errorf("a rating with name '%s' already exists", r.Name)
	}
	return nil
}

func (r *Rating) GetOwner() database.UserId {
	return r.UserId
}

func (r *Rating) PreDelete(ctx context.Context, db database.Provider) error {
	// ratingRef records how many rows in a referencing table still point at this rating.
	type ratingRef struct {
		table string
		count int
	}

	refs := make([]ratingRef, 0, 6)

	formatRatings, err := database.GetAllWhere[*FormatRating](ctx, db, func(_ context.Context, fr *FormatRating) bool {
		return fr.RatingId == r.ID
	})
	if err != nil {
		return err
	}
	if len(formatRatings) > 0 {
		refs = append(refs, ratingRef{"format_rating", len(formatRatings)})
	}

	teamRatings, err := database.GetAllWhere[*TeamRating](ctx, db, func(_ context.Context, tr *TeamRating) bool {
		return tr.RatingId == r.ID
	})
	if err != nil {
		return err
	}
	if len(teamRatings) > 0 {
		refs = append(refs, ratingRef{"team_rating", len(teamRatings)})
	}

	draftRatingCutoffs, err := database.GetAllWhere[*DraftRatingCutoff](ctx, db, func(_ context.Context, dc *DraftRatingCutoff) bool {
		return dc.RatingId == r.ID
	})
	if err != nil {
		return err
	}
	if len(draftRatingCutoffs) > 0 {
		refs = append(refs, ratingRef{"draft_rating_cutoff", len(draftRatingCutoffs)})
	}

	draftPicks, err := database.GetAllWhere[*DraftPick](ctx, db, func(_ context.Context, dp *DraftPick) bool {
		return dp.Rating == r.ID
	})
	if err != nil {
		return err
	}
	if len(draftPicks) > 0 {
		refs = append(refs, ratingRef{"draft_pick", len(draftPicks)})
	}

	preDraftGrades, err := database.GetAllWhere[*PreDraftGrade](ctx, db, func(_ context.Context, pg *PreDraftGrade) bool {
		return pg.Rating == r.ID
	})
	if err != nil {
		return err
	}
	if len(preDraftGrades) > 0 {
		refs = append(refs, ratingRef{"pre_draft_grade", len(preDraftGrades)})
	}

	formatLines, err := database.GetAllWhere[*FormatLine](ctx, db, func(_ context.Context, fl *FormatLine) bool {
		return fl.Player1Rating == r.ID || fl.Player2Rating == r.ID
	})
	if err != nil {
		return err
	}
	if len(formatLines) > 0 {
		refs = append(refs, ratingRef{"format_line", len(formatLines)})
	}

	if len(refs) == 0 {
		return nil
	}

	parts := make([]string, 0, len(refs))
	for _, ref := range refs {
		parts = append(parts, fmt.Sprintf("%d %s", ref.count, ref.table))
	}
	return fmt.Errorf("rating with ID %s is in-use by %s", r.ID, strings.Join(parts, ", "))
}

func (r *Rating) SetOwner(userId database.UserId) {
	r.UserId = userId
}

func NewRating() *Rating {
	return &Rating{}
}

func (r *Rating) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return database.SysAdminAndUsers(r.UserId)
}

func (r *Rating) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (r *Rating) Type() string {
	return "rating"
}

func (r *Rating) GetId() database.RecordId {
	return r.ID.RecordId()
}

func (r *Rating) SetId(id database.RecordId) {
	r.ID = RatingId(id)
}

func (r *Rating) StaticallyValid() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)

	if r.Name == "" {
		return errors.New("rating name is empty")
	}
	if r.Description == "" {
		return errors.New("rating description is empty")
	}
	return nil
}

func (r *Rating) DynamicallyValid(ctx context.Context, db database.Provider) error {
	return database.ExistsById(ctx, db, &User{}, r.UserId.RecordId())
}

func (r *Rating) NewRecord() database.CrudRecord {
	return new(Rating)
}
