package database

import "context"

// Authorizable is embedded into every CrudRecord to define access and edit rights.
//
// Each record has an individual UserId as its owner, who has inherent ability to
// access, edit, and delete their own records. The Owner field is generally intended
// to be immutable, but the SysAdmin role may change it (e.g. reassigning ownership
// of a Season to a new Commissioner during season play). Even the Owner themselves
// should not usually be able to change this field.
//
// EditableBy and AccessibleTo implement the business logic for modification and
// view rights respectively, defined on a record-by-record basis. This allows clean
// specification of authorization rules with easy unit testing.
type Authorizable interface {
	// EditableBy returns the list of UserIds that can modify this record at API-time.
	// For example, this can be used in API middleware to prevent updates to a Season
	// by non-commissioners.
	EditableBy(ctx context.Context, db Provider) []UserId

	// AccessibleTo returns the list of UserIds that can view this record.
	// For example, some information like Availability is private to the Team that
	// it pertains to only.
	AccessibleTo(ctx context.Context, db Provider) []UserId

	// SetOwner sets the UserId of the user who created this record.
	SetOwner(userId UserId)

	// GetOwner returns the UserId of the user who created this record.
	GetOwner() UserId
}
