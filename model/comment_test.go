package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newValidComment(u database.UserId, blurb BlurbId) *Comment {
	c := NewComment()
	c.Owner = u
	c.Content = "content"
	c.Blurb = blurb
	return c
}

func getAnyTeamCaptain(t *testing.T, db database.Provider, season *Season) database.UserId {
	teams, err := season.GetTeams(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	captain, err := teams[0].GetCaptain(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	return captain
}

func copyComment(c *Comment) *Comment {
	return &Comment{
		ID:        c.ID,
		Blurb:     c.Blurb,
		ReplyTo:   c.ReplyTo,
		Owner:     c.Owner,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
		EditedAt:  c.EditedAt,
	}
}

func newStoredComment(t *testing.T, db database.Provider, user database.UserId, blurb *Blurb) *Comment {
	c := newValidComment(user, blurb.ID)

	v, err := database.CreateOne(context.Background(), db, c)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestCommentContentIsNotEmpty(t *testing.T) {
	c := NewComment()
	c.Content = ""
	err := c.StaticallyValid()
	if err == nil {
		t.Error("Empty comment should produce error")
	}
	fmt.Println(err)
}

func TestCommentContentIsWhitespace(t *testing.T) {
	c := NewComment()
	c.Content = "    "
	err := c.StaticallyValid()
	if err == nil {
		t.Error("Whitespace comment should produce error")
	}
	fmt.Println(err)
}

func TestCommentReferencesSelf(t *testing.T) {
	c := NewComment()
	c.ID = CommentId(database.NewRecordId())
	c.ReplyTo = c.ID
	c.Content = "test"
	err := c.StaticallyValid()
	if err == nil {
		t.Error("Comment in reply to itself should produce error")
	}
	fmt.Println(err)
}

func TestCommentCreateDateIsEmpty(t *testing.T) {
	c := NewComment()
	c.ID = CommentId(database.NewRecordId())
	c.Content = "content"
	err := c.StaticallyValid()
	if err == nil {
		t.Error("Empty create date should produce error")
	}
	fmt.Println(err)
}

func TestCommentUserIdIsInvalid(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	c := newValidComment(database.InvalidUserId, blurb.ID)
	err := c.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Error("Invalid user id should produce error")
	}
	fmt.Println(err)
}

func TestEditBySysAdmin(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	c := newStoredComment(t, db, blurb.Owner, blurb)

	sysAdmin := newSysAdmin(t, db)

	copied := copyComment(c)
	copied.Content = "new content"

	wac := database.WithAccessControl[*Comment]{Database: db, AccessControlUser: sysAdmin.ID}
	err := wac.UpdateOneById(context.Background(), copied)
	if err == nil {
		t.Error("Edit by privileged non-owner should produce error")
	}

	fmt.Println(err)
}

func TestEditByCommissioner(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	blurb, season := newDefaultBlurb(t, db)
	teamCaptain := getAnyTeamCaptain(t, db, season)
	c := newStoredComment(t, db, teamCaptain, blurb)

	commissioners, _ := season.GetCommissioners(context.Background(), db)
	commissioner := commissioners[0]

	copied := copyComment(c)
	copied.Content = "new content"

	wac := database.WithAccessControl[*Comment]{Database: db, AccessControlUser: commissioner}
	err := wac.UpdateOneById(context.Background(), copied)
	if err == nil {
		t.Error("Edit by commissioner should produce error")
	}
	fmt.Println(err)
}

func TestEditByOwner(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	c := newStoredComment(t, db, blurb.Owner, blurb)

	originalCreateDate := c.CreatedAt

	copied := copyComment(c)
	copied.Content = "new content"

	wac := database.WithAccessControl[*Comment]{Database: db, AccessControlUser: c.Owner}
	err := wac.UpdateOneById(context.Background(), copied)
	if err != nil {
		t.Error("Edit by owner should not produce error")
	}

	v, err := database.GetExistingRecordById(context.Background(), db, &Comment{}, c.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}
	if v.Content != copied.Content {
		t.Error("Edit by owner should update database value")
	}
	if v.CreatedAt != originalCreateDate {
		t.Error("Edit by owner should not update create date")
	}
	if v.EditedAt.IsZero() {
		t.Error("Edit by owner should update edited date")
	}
}

func TestDeleteBySysAdmin(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)

	c := newStoredComment(t, db, blurb.Owner, blurb)
	sysAdmin := newSysAdmin(t, db)

	wac := database.WithAccessControl[*Comment]{Database: db, AccessControlUser: sysAdmin.ID}
	_, _, err := wac.DeleteOneById(context.Background(), c, c.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteByCommissioner(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	blurb, season := newDefaultBlurb(t, db)
	teamCaptain := getAnyTeamCaptain(t, db, season)
	c := newStoredComment(t, db, teamCaptain, blurb)

	commissioners, _ := season.GetCommissioners(context.Background(), db)
	commissioner := commissioners[0]

	copied := copyComment(c)
	copied.Content = "new content"

	wac := database.WithAccessControl[*Comment]{Database: db, AccessControlUser: commissioner}
	err := wac.UpdateOneById(context.Background(), copied)
	if err == nil {
		t.Error("Edit by commissioner should produce error")
	}
	fmt.Println(err)
}

func TestDeleteByOwner(t *testing.T) {
	database.SysAdminCheck = IsUserSystemAdministrator
	db := database.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	c := newStoredComment(t, db, blurb.Owner, blurb)

	wac := database.WithAccessControl[*Comment]{Database: db, AccessControlUser: c.Owner}
	_, _, err := wac.DeleteOneById(context.Background(), c, c.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}
}

func TestCommentByNonSeasonParticipant(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	otherUser := newStoredUser(t, db)

	comment := newValidComment(otherUser.ID, blurb.ID)
	err := comment.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Error("Comment by non-season participant should produce error")
	}
	fmt.Println(err)
}

func TestCommentPostDeleteCascadesReactions(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	blurb, season := newDefaultBlurb(t, db)
	owner := getAnyTeamCaptain(t, db, season)
	comment := newStoredComment(t, db, owner, blurb)

	err := comment.React(context.Background(), db, owner, ThumbsUp)
	if err != nil {
		t.Fatal(err)
	}

	count := func() int {
		rows, err := database.GetAllWhere[*CommentReaction](context.Background(), db, func(_ context.Context, r *CommentReaction) bool {
			return r.CommentId == comment.ID
		})
		if err != nil {
			t.Fatal(err)
		}
		return len(rows)
	}
	if count() == 0 {
		t.Fatal("expected comment to have reactions")
	}

	_, _, err = database.DeleteOneById(context.Background(), db, &Comment{}, comment.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}
	if count() != 0 {
		t.Fatal("expected 0 reactions after delete")
	}
}
