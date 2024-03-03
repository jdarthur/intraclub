package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Team struct {
	ID         string    `json:"team_id"`     // unique ID for this team
	Year       int       `json:"-"`           // year for this team, e.g. 2024
	Name       string    `json:"name"`        // custom team name
	Color      TeamColor `json:"color"`       // red, blue, green, white
	CaptainId  string    `json:"captain"`     // user ID of captain
	CoCaptains []string  `json:"co_captains"` // user ID(s) of any co-captains
	Players    []Player  `json:"players"`     // list of Player s on team
}

func NewTeam(color TeamColor, captain string) Team {
	id := primitive.NewObjectID()

	return Team{
		ID:        id.String(),
		Color:     color,
		CaptainId: captain,
	}
}
