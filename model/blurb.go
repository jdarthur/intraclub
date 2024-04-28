package model

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"strings"
	"time"
)

// Blurb is an object that is used in the main feed for a given league / season.
// It allows the commissioner or designated reporters to create posts related to h
// that league. These are either informational posts or weekly updates for example.
//
// The blurb creator must add a free text value and can optionally add a list of
// Image s to attach to the bottom of the blurb.
//
// The Comments and Reactions are added by User s in order to either hit an emoji
// Reaction to the blurb or add a comment to it.
type Blurb struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"` // the main object ID for this blurb
	Author    string             `json:"author"`        // id of the User that created this Blurb
	LeagueId  string             `json:"league_id"`     // ID of the League that this applies to (so we can get all Blurb s for a League)
	Content   string             `json:"content"`       // free text where the commish / reporter can add a weekly summary for the main content of this Blurb
	Gallery   []string           `json:"gallery"`       // list of IDs that correspond to an Image
	CreatedAt time.Time          `json:"created_at"`    // Date / time when this blurb was posted
	Comments  []string           `json:"comments"`      // list of Comment IDs referencing this blurb
	Reactions []Reaction         `json:"reactions"`     // list of Reaction s to this blurb
}

func (b *Blurb) RecordType() string {
	return "blurb"
}

func (b *Blurb) OneRecord() common.CrudRecord {
	return new(Blurb)
}

type listOfBlurbs []*Blurb

func (l listOfBlurbs) Length() int {
	return len(l)
}

func (l listOfBlurbs) Get(index int) common.CrudRecord {
	return l[index]
}

func (b *Blurb) ListOfRecords() common.ListOfCrudRecords {
	return listOfBlurbs{}
}

func (b *Blurb) SetId(id primitive.ObjectID) {
	b.ID = id
}

func (b *Blurb) GetId() primitive.ObjectID {
	return b.ID
}

func (b *Blurb) ValidateStatic() error {

	b.Content = strings.Trim(b.Content, " \t\r\n")

	if b.Content == "" {
		return errors.New("content must not be empty")
	}

	return nil
}

func (b *Blurb) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {

	err := common.CheckExistenceOrErrorByStringId(db, &User{}, b.Author)
	if err != nil {
		return err
	}

	// validate all of the comment IDs and the comments themselves for correctness
	for _, commentId := range b.Comments {
		err = common.GetOneByIdAndValidate(db, &Comment{}, commentId)
		if err != nil {
			return err
		}
	}

	// validate all of the Reactions in the list
	for _, reaction := range b.Reactions {
		err = common.ValidateStaticAndDynamic(db, reaction)
	}

	for _, imageId := range b.Gallery {
		err = common.GetOneByIdAndValidate(db, &Image{}, imageId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Blurb) AddComment(db common.DbProvider, commentId string) error {
	comment, err := GetComment(db, commentId)
	if err != nil {
		return err
	}

	err = common.ValidateStaticAndDynamic(db, comment)
	if err != nil {
		return err
	}

	if comment.References != b.ID.Hex() {
		return fmt.Errorf("comment '%s' does not reference blurb '%s'", commentId, b.ID.Hex())
	}

	canComment, err := b.UserCanComment(db, comment.UserID)
	if !canComment {
		return err
	}

	b.Comments = append(b.Comments, commentId)
	return nil
}

func (b *Blurb) UserCanComment(db common.DbProvider, userId string) (bool, error) {

	if userId == "" {
		return false, errors.New("user ID is empty")
	}

	league, err := b.GetLeague(db)
	if err != nil {
		return false, err
	}

	teams, err := GetTeams(db, league)
	if err != nil {
		return false, err
	}
	for _, team := range teams {

		players, err := team.GetPlayers(db)
		if err != nil {
			return false, err
		}

		for _, player := range players {
			if player.UserId == userId {
				return true, nil
			}
		}
	}

	return false, errors.New("user is not a member of league")
}

func GetTeams(db common.DbProvider, league *League) ([]*Team, error) {
	return make([]*Team, 0), nil
}

func (b *Blurb) GetLeague(db common.DbProvider) (*League, error) {
	id, err := primitive.ObjectIDFromHex(b.LeagueId)
	if err != nil {
		return nil, err
	}

	league, exists, err := common.GetOne(db, &League{ID: id})
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, common.RecordDoesNotExist(&League{ID: id})
	}

	return league.(*League), nil
}

func GetComment(db common.DbProvider, commentId string) (*Comment, error) {
	comment, err := common.GetOneByStringId(db, &Comment{}, commentId)
	if err != nil {
		return nil, err
	}
	return comment.(*Comment), nil
}
