package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

type Team struct {
	ID         primitive.ObjectID `json:"team_id" bson:"_id"` // unique ID for this team
	LeagueId   string             `json:"league_id" bson:"league_id"`
	Year       int                `json:"-" bson:"year"`                  // year for this team, e.g. 2024
	Name       string             `json:"name" bson:"name"`               // custom team name
	Color      TeamColor          `json:"color" bson:"color"`             // red, blue, green, white
	CaptainId  string             `json:"captain" bson:"captain_id"`      // user ID of captain
	CoCaptains []string           `json:"co_captains" bson:"co_captains"` // user ID(s) of any co-captains
	Players    []string           `json:"players" bson:"players"`         // list of Player s on team
	Active     bool               `json:"active,omitempty" bson:"-"`
}

func (t *Team) RecordType() string {
	return "team"
}

func (t *Team) OneRecord() common.CrudRecord {
	return new(Team)
}

type listOfTeams []*Team

func (l listOfTeams) Get(index int) common.CrudRecord {
	return l[index]
}

func (l listOfTeams) Length() int {
	return len(l)
}

func (t *Team) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfTeams, 0)
}

func (t *Team) SetId(id primitive.ObjectID) {
	t.ID = id
}

func (t *Team) GetId() primitive.ObjectID {
	return t.ID
}

func (t *Team) ValidateStatic() error {

	err := t.Color.ValidateStatic()
	if err != nil {
		return fmt.Errorf("invalid team color: %s", err)
	}

	return nil
}

func (t *Team) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {

	err := common.CheckExistenceOrErrorByStringId(db, &User{}, t.CaptainId)
	if err != nil {
		return err
	}

	captainInPlayers, err := t.userIdIsInPlayerList(db, t.CaptainId)
	if err != nil {
		return err
	}

	if !captainInPlayers {
		return fmt.Errorf("captain %s is not in player list", t.CaptainId)
	}

	err = common.CheckExistenceOrErrorByStringId(db, &League{}, t.LeagueId)
	if err != nil {
		return err
	}

	for _, coCaptain := range t.CoCaptains {
		err = common.CheckExistenceOrErrorByStringId(db, &User{}, coCaptain)
		if err != nil {
			return fmt.Errorf("error with co-captain ID %s: %s", coCaptain, err.Error())
		}

		coCaptainInPlayers, err := t.userIdIsInPlayerList(db, coCaptain)
		if err != nil {
			return err
		}

		if !coCaptainInPlayers {
			return fmt.Errorf("co-captain %s is not in player list", coCaptain)
		}
	}

	for _, player := range t.Players {
		err = common.CheckExistenceOrErrorByStringId(db, &Player{}, player)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Team) userIdIsInPlayerList(db common.DbProvider, userId string) (bool, error) {

	players, err := t.GetPlayers(db)
	if err != nil {
		return false, err
	}

	for _, player := range players {

		if player.UserId == userId {
			return true, nil
		}
	}
	return false, nil
}

func (t *Team) playerIdIsInPlayerList(playerId string) bool {
	for _, p := range t.Players {
		if p == playerId {
			return true
		}
	}
	return false
}

func (t *Team) GetPlayers(db common.DbProvider) ([]*Player, error) {

	players := make([]*Player, 0)
	for _, playerId := range t.Players {

		player, err := common.GetOneByStringId(db, &Player{}, playerId)
		if err != nil {
			return nil, err
		}

		players = append(players, player.(*Player))
	}

	return players, nil
}

func (t *Team) PlayerIdListIsEqual(newPlayers []string) bool {
	if len(t.Players) != len(newPlayers) {
		return false
	}

	for _, player := range newPlayers {
		if !t.playerIdIsInPlayerList(player) {
			return false
		}
	}

	return true
}

func (t *Team) ValidatePlayerUpdate(db common.DbProvider, apiUserId primitive.ObjectID, newPlayers []string) error {

	if !t.PlayerIdListIsEqual(newPlayers) {
		return nil
	}

	search := &User{ID: apiUserId}

	user, exists, err := common.GetOne(db, search)
	if err != nil {
		return err
	}

	if !exists {
		return common.RecordDoesNotExist(search)
	}

	leagueId, err := primitive.ObjectIDFromHex(t.LeagueId)
	if err != nil {
		return err
	}

	search2 := &League{ID: leagueId}

	league, exists, err := common.GetOne(db, search2)
	if err != nil {
		return err
	}
	if !exists {
		return common.RecordDoesNotExist(search2)
	}

	if user.GetId().String() != league.(*League).Commissioner {
		return fmt.Errorf("team's players may only be updated by the league commissioner")
	}

	return nil

}

func (t *Team) AddPlayerByUserId(db common.DbProvider, userId primitive.ObjectID, line int) error {

	player := &Player{
		UserId: userId.Hex(),
		Line:   line,
	}

	created, err := common.Create(db, player)
	if err != nil {
		return err
	}

	t.Players = append(t.Players, created.GetId().Hex())

	return nil
}

func NewTeam(color TeamColor, captain string) Team {
	return Team{
		Color:     color,
		CaptainId: captain,
	}
}

func (t *Team) IsActive(db common.DbProvider) (bool, error) {
	v, err := common.GetOneByStringId(db, &League{}, t.LeagueId)
	if err != nil {
		return false, err
	}

	league := v.(*League)

	return league.IsActive(db)
}
