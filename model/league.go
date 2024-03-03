package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type League struct {
	ID           string      `json:"league_id"`
	Colors       []TeamColor `json:"colors"`
	Commissioner string      `json:"commissioner"`
}

func NewLeague(colors []TeamColor, commissioner string) League {
	leagueId := primitive.NewObjectID()
	return League{
		ID:           leagueId.String(),
		Colors:       colors,
		Commissioner: commissioner,
	}
}
