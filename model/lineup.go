package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

type LineupId database.RecordId

func (id LineupId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id LineupId) String() string {
	return id.RecordId().String()
}

func (id LineupId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id *LineupId) UnmarshalJSON(bytes []byte) error {
	rid := database.RecordId(0)
	if err := (*database.RecordId)(&rid).UnmarshalJSON(bytes); err != nil {
		return err
	}
	*id = LineupId(rid)
	return nil
}

type Lineup struct {
	ID        LineupId `json:"id"`
	TeamId    TeamId   `json:"team_id"` // TeamId for this particular Lineup
	WeekId    WeekId   `json:"week_id"` // Week that this Lineup applies to
	Confirmed bool     `json:"confirmed"` // set by the captain/co-captain once the lineup is final
	Official  bool     `json:"official"`  // set by the season commissioner, requires Confirmed
}

func (l *Lineup) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (l *Lineup) UniquenessEquivalent(other *Lineup) error {
	if l.WeekId == other.WeekId && l.TeamId == other.TeamId {
		return fmt.Errorf("duplicate record for team ID & week ID")
	}
	return nil
}

func (l *Lineup) SetOwner(userId database.UserId) {
	// don't need to do anything here as ownership is enforced by
	// team captain or co-captain status
}

func (l *Lineup) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return EditableByTeamCaptainOrCoCaptains(ctx, db, l.TeamId)
}

func (l *Lineup) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return AccessibleByTeamMembers(ctx, db, l.TeamId)
}

func (l *Lineup) Type() string {
	return "lineup"
}

func (l *Lineup) GetId() database.RecordId {
	return l.ID.RecordId()
}

func (l *Lineup) SetId(id database.RecordId) {
	l.ID = LineupId(id)
}

func (l *Lineup) StaticallyValid() error {
	return nil
}

func (l *Lineup) DynamicallyValid(ctx context.Context, db database.Provider) error {
	err := database.ExistsById(ctx, db, &Team{}, l.TeamId.RecordId())
	if err != nil {
		return err
	}
	err = database.ExistsById(ctx, db, &Week{}, l.WeekId.RecordId())
	if err != nil {
		return err
	}
	return nil
}

func (l *Lineup) GetFormat(ctx context.Context, db database.Provider) (*Format, error) {
	week, err := database.GetExistingRecordById(ctx, db, &Week{}, l.WeekId.RecordId())
	if err != nil {
		return nil, err
	}
	draft, err := database.GetExistingRecordById(ctx, db, &Draft{}, week.DraftId.RecordId())
	if err != nil {
		return nil, err
	}
	return database.GetExistingRecordById(ctx, db, &Format{}, draft.Format.RecordId())
}

// PostDelete cascades deletion to this lineup's lineup_pairing child rows.
// Without this, deleting a lineup would orphan those rows (see #97).
func (l *Lineup) PostDelete(ctx context.Context, db database.Provider) error {
	pairings, err := database.GetAllWhere[*LineupPairing](ctx, db, func(_ context.Context, p *LineupPairing) bool {
		return p.LineupId == l.ID
	})
	if err != nil {
		return err
	}
	for _, p := range pairings {
		if _, _, err := database.DeleteOneById(ctx, db, p, p.ID.RecordId()); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lineup) NewRecord() database.CrudRecord {
	return new(Lineup)
}

// NewLineup returns a new Lineup record. It is the conventional constructor
// used by the generic CRUD route registration.
func NewLineup() *Lineup {
	return &Lineup{}
}
