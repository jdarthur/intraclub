package test

import (
	"intraclub/models"
	"testing"
)

func TestBasicLineResult(t *testing.T) {
	lineResult := models.LineResult{
		Team1:      matchup(Team1Id, true),
		Team2:      matchup(Team2Id, false),
		SetResults: threeSetResults(),
	}

	err := lineResult.ValidateStatic()
	if err != nil {
		t.Error(err)
	}
}

func matchup(teamId string, isTeam1 bool) models.Matchup {
	player1 := TomEasum
	if !isTeam1 {
		player1 = EthanMoland
	}

	player2 := AndyLascik
	if !isTeam1 {
		player2 = JdArthur
	}

	return models.Matchup{
		Team:           teamId,
		Line1:          player1.Line,
		Player1:        player1.PlayerId,
		Player1Penalty: false,
		Line2:          player2.Line,
		Player2:        player2.PlayerId,
		Player2Penalty: false,
	}
}

func threeSetResults() []models.SetResult {
	return []models.SetResult{
		{
			Team1: models.TeamSetResult{
				ID:       Team1Id,
				GamesWon: 7,
			},
			Team2: models.TeamSetResult{
				ID:       Team2Id,
				GamesWon: 5,
			},
		},
		{
			Team1: models.TeamSetResult{
				ID:       Team1Id,
				GamesWon: 4,
			},
			Team2: models.TeamSetResult{
				ID:       Team2Id,
				GamesWon: 6,
			},
		},
		{
			Team1: models.TeamSetResult{
				ID:       Team1Id,
				GamesWon: 6,
			},
			Team2: models.TeamSetResult{
				ID:       Team2Id,
				GamesWon: 3,
			},
		},
	}
}
