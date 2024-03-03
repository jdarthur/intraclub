package test

import (
	"intraclub/model"
	"testing"
)

func TestBasicLineResult(t *testing.T) {
	lineResult := model.LineResult{
		Team1:      matchup(Team1Id, true),
		Team2:      matchup(Team2Id, false),
		SetResults: threeSetResults(),
	}

	err := lineResult.ValidateStatic()
	if err != nil {
		t.Error(err)
	}
}

func matchup(teamId string, isTeam1 bool) model.Matchup {
	player1 := Tom
	if !isTeam1 {
		player1 = Ethan
	}

	player2 := Andy
	if !isTeam1 {
		player2 = JD
	}

	return model.Matchup{
		Team:           teamId,
		Line1:          player1.Line,
		Player1:        player1.UserId,
		Player1Penalty: false,
		Line2:          player2.Line,
		Player2:        player2.UserId,
		Player2Penalty: false,
	}
}

func threeSetResults() []model.SetResult {
	return []model.SetResult{
		{
			Team1: model.TeamSetResult{
				ID:       Team1Id,
				GamesWon: 7,
			},
			Team2: model.TeamSetResult{
				ID:       Team2Id,
				GamesWon: 5,
			},
		},
		{
			Team1: model.TeamSetResult{
				ID:       Team1Id,
				GamesWon: 4,
			},
			Team2: model.TeamSetResult{
				ID:       Team2Id,
				GamesWon: 6,
			},
		},
		{
			Team1: model.TeamSetResult{
				ID:       Team1Id,
				GamesWon: 6,
			},
			Team2: model.TeamSetResult{
				ID:       Team2Id,
				GamesWon: 3,
			},
		},
	}
}
