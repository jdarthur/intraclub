package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newValidComment(u UserId, blurb BlurbId) *Comment {
	c := NewComment()
	c.Owner = u
	c.Content = "content"
	c.Blurb = blurb
	return c
}

func getAnyTeamCaptain(t *testing.T, db common.DatabaseProvider, season *Season) UserId {
	teams, err := season.GetTeams(db)
	if err != nil {
		t.Fatal(err)
	}
	return teams[0].Captain
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
		Reactions: c.Reactions,
	}
}

func newStoredComment(t *testing.T, db common.DatabaseProvider, user UserId, blurb *Blurb) *Comment {
	c := newValidComment(user, blurb.ID)

	v, err := common.CreateOne(db, c)
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
	c.ID = CommentId(common.NewRecordId())
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
	c.ID = CommentId(common.NewRecordId())
	c.Content = "content"
	err := c.StaticallyValid()
	if err == nil {
		t.Error("Empty create date should produce error")
	}
	fmt.Println(err)
}

func TestCommentUserIdIsInvalid(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	c := newValidComment(UserId(common.InvalidRecordId), blurb.ID)
	err := c.DynamicallyValid(db)
	if err == nil {
		t.Error("Invalid user id should produce error")
	}
	fmt.Println(err)
}

func TestEditBySysAdmin(t *testing.T) {
	common.SysAdminCheck = IsUserSystemAdministrator

	db := common.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	c := newStoredComment(t, db, blurb.Owner, blurb)

	sysAdmin := newStoredUser(t, db)
	err := sysAdmin.AssignRole(db, SystemAdministrator)
	if err != nil {
		t.Fatal(err)
	}

	copied := copyComment(c)
	copied.Content = "new content"

	wac := common.WithAccessControl[*Comment]{Database: db, AccessControlUser: sysAdmin.ID.RecordId()}
	err = wac.UpdateOneById(copied)
	if err == nil {
		t.Error("Edit by privileged non-owner should produce error")
	}

	fmt.Println(err)
}

func TestEditByCommissioner(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	blurb, season := newDefaultBlurb(t, db)
	teamCaptain := getAnyTeamCaptain(t, db, season)
	c := newStoredComment(t, db, teamCaptain, blurb)

	commissioner := season.Commissioners[0]

	copied := copyComment(c)
	copied.Content = "new content"

	wac := common.WithAccessControl[*Comment]{Database: db, AccessControlUser: commissioner.RecordId()}
	err := wac.UpdateOneById(copied)
	if err == nil {
		t.Error("Edit by commissioner should produce error")
	}
	fmt.Println(err)
}

func TestEditByOwner(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	c := newStoredComment(t, db, blurb.Owner, blurb)

	originalCreateDate := c.CreatedAt

	copied := copyComment(c)
	copied.Content = "new content"

	wac := common.WithAccessControl[*Comment]{Database: db, AccessControlUser: c.Owner.RecordId()}
	err := wac.UpdateOneById(copied)
	if err != nil {
		t.Error("Edit by owner should not produce error")
	}

	v, err := common.GetExistingRecordById(db, &Comment{}, c.ID.RecordId())
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
	common.SysAdminCheck = IsUserSystemAdministrator

	db := common.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)

	c := newStoredComment(t, db, blurb.Owner, blurb)

	sysAdmin := newStoredUser(t, db)
	err := sysAdmin.AssignRole(db, SystemAdministrator)
	if err != nil {
		t.Fatal(err)
	}

	wac := common.WithAccessControl[*Comment]{Database: db, AccessControlUser: sysAdmin.ID.RecordId()}
	_, _, err = wac.DeleteOneById(c, c.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteByCommissioner(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	blurb, season := newDefaultBlurb(t, db)
	teamCaptain := getAnyTeamCaptain(t, db, season)
	c := newStoredComment(t, db, teamCaptain, blurb)

	commissioner := season.Commissioners[0]

	copied := copyComment(c)
	copied.Content = "new content"

	wac := common.WithAccessControl[*Comment]{Database: db, AccessControlUser: commissioner.RecordId()}
	err := wac.UpdateOneById(copied)
	if err == nil {
		t.Error("Edit by commissioner should produce error")
	}
	fmt.Println(err)
}

func TestDeleteByOwner(t *testing.T) {
	common.SysAdminCheck = IsUserSystemAdministrator
	db := common.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	c := newStoredComment(t, db, blurb.Owner, blurb)

	wac := common.WithAccessControl[*Comment]{Database: db, AccessControlUser: c.Owner.RecordId()}
	_, _, err := wac.DeleteOneById(c, c.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}
}

func TestCommentByNonSeasonParticipant(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	blurb, _ := newDefaultBlurb(t, db)
	otherUser := newStoredUser(t, db)

	comment := newValidComment(otherUser.ID, blurb.ID)
	err := comment.DynamicallyValid(db)
	if err == nil {
		t.Error("Comment by non-season participant should produce error")
	}
	fmt.Println(err)
}
