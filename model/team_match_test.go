package model

import (
	"context"
	"testing"
	"time"

	"intraclub/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newStoredTeamMatch builds the full chain required to persist a TeamMatch:
// format -> draft -> two teams -> week -> lineup -> team match.
func newStoredTeamMatch(t *testing.T, db database.Provider) *TeamMatch {
	format := newDefaultStoredFormat(t, db)
	commissioner := newStoredUser(t, db)

	draft := NewDraft()
	draft.Owner = commissioner.ID
	draft.Format = format.ID
	draftV, err := database.CreateOne(context.Background(), db, draft)
	require.NoError(t, err)

	team1 := newStoredTeam(t, db, newStoredUser(t, db).ID)
	team2 := newStoredTeam(t, db, newStoredUser(t, db).ID)

	week := NewWeek()
	week.DraftId = draftV.ID
	week.Date = time.Date(2020, 1, 1, 8, 0, 0, 0, time.UTC)
	weekV, err := database.CreateOne(context.Background(), db, week)
	require.NoError(t, err)

	lineup := &Lineup{TeamId: team1.ID, WeekId: weekV.ID}
	lineupV, err := database.CreateOne(context.Background(), db, lineup)
	require.NoError(t, err)

	tm := &TeamMatch{WeekId: weekV.ID, HomeTeam: team1.ID, AwayTeam: team2.ID, Lineup: lineupV.ID}
	tmV, err := database.CreateOne(context.Background(), db, tm)
	require.NoError(t, err)
	return tmV
}

// newStoredLineupPairing builds a LineupPairing belonging to the given lineup,
// with two freshly-created users added as members of the lineup's team. The two
// players are assigned ratings matching the format's line 0 so the pairing
// passes DynamicallyValid's rating check.
func newStoredLineupPairing(t *testing.T, db database.Provider, lineup *Lineup, team *Team) *LineupPairing {
	player1 := newStoredUser(t, db)
	player2 := newStoredUser(t, db)

	// Resolve the format line 0's required ratings so the players carry them.
	week, err := database.GetExistingRecordById(context.Background(), db, &Week{}, lineup.WeekId.RecordId())
	require.NoError(t, err)
	draft, err := database.GetExistingRecordById(context.Background(), db, &Draft{}, week.DraftId.RecordId())
	require.NoError(t, err)
	format, err := database.GetExistingRecordById(context.Background(), db, &Format{}, draft.Format.RecordId())
	require.NoError(t, err)
	lines, err := format.GetLines(context.Background(), db)
	require.NoError(t, err)
	line := lines[0]

	for _, player := range []database.UserId{player1.ID, player2.ID} {
		assignment := &TeamAssignment{
			TeamId: team.ID,
			UserId: player,
			Role:   TeamRoleMember,
		}
		_, err := database.CreateOne(context.Background(), db, assignment)
		require.NoError(t, err)
	}

	ratings := []database.UserId{player1.ID, player2.ID}
	required := []RatingId{line.Player1Rating, line.Player2Rating}
	for i, player := range ratings {
		teamRating := &TeamRating{
			TeamId:   team.ID,
			UserId:   player,
			RatingId: required[i],
		}
		_, err := database.CreateOne(context.Background(), db, teamRating)
		require.NoError(t, err)
	}

	lp := &LineupPairing{
		LineupId:        lineup.ID,
		TeamId:          team.ID,
		Player1:         player1.ID,
		Player2:         player2.ID,
		FormatLineIndex: 0,
	}
	v, err := database.CreateOne(context.Background(), db, lp)
	require.NoError(t, err)
	return v
}

// newStoredIndividualMatch builds a single, unpaired IndividualMatch.
func newStoredIndividualMatch(t *testing.T, db database.Provider) *IndividualMatch {
	scoring := newDefaultStoredScoringStructure(t, db)
	match := NewMatch()
	match.Structure = scoring.ID
	v, err := database.CreateOne(context.Background(), db, match)
	require.NoError(t, err)
	if _, err := v.AssignEditor(context.Background(), db, scoring.Owner); err != nil {
		require.NoError(t, err)
	}
	return v
}

func TestTeamMatchIndividualMatchRoundTrip(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	teamMatch := newStoredTeamMatch(t, db)
	lineup, err := database.GetExistingRecordById(context.Background(), db, &Lineup{}, teamMatch.Lineup.RecordId())
	require.NoError(t, err)
	homeTeam, err := database.GetExistingRecordById(context.Background(), db, &Team{}, teamMatch.HomeTeam.RecordId())
	require.NoError(t, err)
	pairing := newStoredLineupPairing(t, db, lineup, homeTeam)
	match := newStoredIndividualMatch(t, db)

	row := &TeamMatchIndividualMatch{
		TeamMatchId:       teamMatch.ID,
		LineupPairingId:   pairing.ID,
		IndividualMatchId: match.ID,
	}
	created, err := database.CreateOne(context.Background(), db, row)
	require.NoError(t, err)

	got, exists, err := database.GetOneById(context.Background(), db, &TeamMatchIndividualMatch{}, created.GetId())
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, teamMatch.ID, got.TeamMatchId)
	assert.Equal(t, pairing.ID, got.LineupPairingId)
	assert.Equal(t, match.ID, got.IndividualMatchId)
}

func TestTeamMatchIndividualMatchUniquenessConstraint(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	teamMatch := newStoredTeamMatch(t, db)
	lineup, err := database.GetExistingRecordById(context.Background(), db, &Lineup{}, teamMatch.Lineup.RecordId())
	require.NoError(t, err)
	homeTeam, err := database.GetExistingRecordById(context.Background(), db, &Team{}, teamMatch.HomeTeam.RecordId())
	require.NoError(t, err)
	pairing := newStoredLineupPairing(t, db, lineup, homeTeam)
	match1 := newStoredIndividualMatch(t, db)
	match2 := newStoredIndividualMatch(t, db)

	row1 := &TeamMatchIndividualMatch{
		TeamMatchId:       teamMatch.ID,
		LineupPairingId:   pairing.ID,
		IndividualMatchId: match1.ID,
	}
	_, err = database.CreateOne(context.Background(), db, row1)
	require.NoError(t, err)

	// a second assignment for the same (teamMatch, lineupPairing) must be rejected
	row2 := &TeamMatchIndividualMatch{
		TeamMatchId:       teamMatch.ID,
		LineupPairingId:   pairing.ID,
		IndividualMatchId: match2.ID,
	}
	_, err = database.CreateOne(context.Background(), db, row2)
	assert.Error(t, err)
}

func TestTeamMatchIndividualMatchDynamicallyValid(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	teamMatch := newStoredTeamMatch(t, db)
	lineup, err := database.GetExistingRecordById(context.Background(), db, &Lineup{}, teamMatch.Lineup.RecordId())
	require.NoError(t, err)
	homeTeam, err := database.GetExistingRecordById(context.Background(), db, &Team{}, teamMatch.HomeTeam.RecordId())
	require.NoError(t, err)
	pairing := newStoredLineupPairing(t, db, lineup, homeTeam)
	match := newStoredIndividualMatch(t, db)

	// invalid team match ID
	badTeamMatch := &TeamMatchIndividualMatch{
		TeamMatchId:       TeamMatchId(database.InvalidRecordId),
		LineupPairingId:   pairing.ID,
		IndividualMatchId: match.ID,
	}
	assert.Error(t, badTeamMatch.DynamicallyValid(context.Background(), db))

	// invalid lineup pairing ID
	badPairing := &TeamMatchIndividualMatch{
		TeamMatchId:       teamMatch.ID,
		LineupPairingId:   LineupPairingId(database.InvalidRecordId),
		IndividualMatchId: match.ID,
	}
	assert.Error(t, badPairing.DynamicallyValid(context.Background(), db))

	// invalid individual match ID
	badMatch := &TeamMatchIndividualMatch{
		TeamMatchId:       teamMatch.ID,
		LineupPairingId:   pairing.ID,
		IndividualMatchId: IndividualMatchId(database.InvalidRecordId),
	}
	assert.Error(t, badMatch.DynamicallyValid(context.Background(), db))
}

func TestTeamMatchAssignAndGetIndividualMatches(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	teamMatch := newStoredTeamMatch(t, db)
	lineup, err := database.GetExistingRecordById(context.Background(), db, &Lineup{}, teamMatch.Lineup.RecordId())
	require.NoError(t, err)
	homeTeam, err := database.GetExistingRecordById(context.Background(), db, &Team{}, teamMatch.HomeTeam.RecordId())
	require.NoError(t, err)
	pairing1 := newStoredLineupPairing(t, db, lineup, homeTeam)
	pairing2 := newStoredLineupPairing(t, db, lineup, homeTeam)
	match1 := newStoredIndividualMatch(t, db)
	match2 := newStoredIndividualMatch(t, db)

	_, err = teamMatch.AssignIndividualMatch(context.Background(), db, pairing1.ID, match1.ID)
	require.NoError(t, err)
	_, err = teamMatch.AssignIndividualMatch(context.Background(), db, pairing2.ID, match2.ID)
	require.NoError(t, err)

	matches, err := teamMatch.GetIndividualMatches(context.Background(), db)
	require.NoError(t, err)
	assert.Len(t, matches, 2)
	assert.Equal(t, match1.ID, matches[pairing1.ID])
	assert.Equal(t, match2.ID, matches[pairing2.ID])
}

func TestTeamMatchValidateMatchesVsLineup(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	teamMatch := newStoredTeamMatch(t, db)
	lineup, err := database.GetExistingRecordById(context.Background(), db, &Lineup{}, teamMatch.Lineup.RecordId())
	require.NoError(t, err)
	homeTeam, err := database.GetExistingRecordById(context.Background(), db, &Team{}, teamMatch.HomeTeam.RecordId())
	require.NoError(t, err)
	pairing := newStoredLineupPairing(t, db, lineup, homeTeam)
	match := newStoredIndividualMatch(t, db)

	// valid assignment within the team match's lineup passes
	_, err = teamMatch.AssignIndividualMatch(context.Background(), db, pairing.ID, match.ID)
	require.NoError(t, err)
	require.NoError(t, teamMatch.ValidateMatchesVsLineup(context.Background(), db))

	// a pairing belonging to a different lineup must fail
	otherTeam := newStoredTeam(t, db, newStoredUser(t, db).ID)
	otherWeek := NewWeek()
	otherWeek.DraftId = newStoredDraft(t, db, newStoredUser(t, db).ID).ID
	otherWeek.Date = time.Date(2020, 2, 1, 8, 0, 0, 0, time.UTC)
	otherWeekV, err := database.CreateOne(context.Background(), db, otherWeek)
	require.NoError(t, err)
	otherLineup := &Lineup{TeamId: otherTeam.ID, WeekId: otherWeekV.ID}
	otherLineupV, err := database.CreateOne(context.Background(), db, otherLineup)
	require.NoError(t, err)
	otherPairing := newStoredLineupPairing(t, db, otherLineupV, otherTeam)
	otherMatch := newStoredIndividualMatch(t, db)

	_, err = teamMatch.AssignIndividualMatch(context.Background(), db, otherPairing.ID, otherMatch.ID)
	require.NoError(t, err)
	assert.Error(t, teamMatch.ValidateMatchesVsLineup(context.Background(), db))
}

func TestTeamMatchPostDeleteCascadesIndividualMatches(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	teamMatch := newStoredTeamMatch(t, db)
	lineup, err := database.GetExistingRecordById(context.Background(), db, &Lineup{}, teamMatch.Lineup.RecordId())
	require.NoError(t, err)
	homeTeam, err := database.GetExistingRecordById(context.Background(), db, &Team{}, teamMatch.HomeTeam.RecordId())
	require.NoError(t, err)
	pairing := newStoredLineupPairing(t, db, lineup, homeTeam)
	match := newStoredIndividualMatch(t, db)

	_, err = teamMatch.AssignIndividualMatch(context.Background(), db, pairing.ID, match.ID)
	require.NoError(t, err)

	count := func() int {
		rows, err := database.GetAllWhere[*TeamMatchIndividualMatch](context.Background(), db, func(_ context.Context, r *TeamMatchIndividualMatch) bool {
			return r.TeamMatchId == teamMatch.ID
		})
		require.NoError(t, err)
		return len(rows)
	}
	require.Equal(t, 1, count())

	_, _, err = database.DeleteOneById(context.Background(), db, &TeamMatch{}, teamMatch.ID.RecordId())
	require.NoError(t, err)

	require.Equal(t, 0, count())
}

func TestLineupPostDeleteCascadesPairings(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	teamMatch := newStoredTeamMatch(t, db)
	lineup, err := database.GetExistingRecordById(context.Background(), db, &Lineup{}, teamMatch.Lineup.RecordId())
	require.NoError(t, err)
	homeTeam, err := database.GetExistingRecordById(context.Background(), db, &Team{}, teamMatch.HomeTeam.RecordId())
	require.NoError(t, err)
	newStoredLineupPairing(t, db, lineup, homeTeam)

	count := func() int {
		rows, err := database.GetAllWhere[*LineupPairing](context.Background(), db, func(_ context.Context, p *LineupPairing) bool {
			return p.LineupId == lineup.ID
		})
		require.NoError(t, err)
		return len(rows)
	}
	require.Equal(t, 1, count())

	_, _, err = database.DeleteOneById(context.Background(), db, &Lineup{}, lineup.ID.RecordId())
	require.NoError(t, err)

	require.Equal(t, 0, count())
}
