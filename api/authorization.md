# Authorization model

This document describes the authorization rules for user access and edit rights
on object in the provisioning database. As a summary:
- Each `CrudRecord` has an individual `UserId` as an owner and the _owner_ of a 
  record has inherent ability to access, edit and delete their own records
- Some higher-level grouping constructs such as `Team`s and `Season`s define the 
  rules around access to particular record types such as `Availability` or `Lineup`
- Some special roles such as `Commissioner`, `Captain` and `SysAdmin` allow an 
  individual `User` elevated rights to perform specific modification actions

## Database-level
 - [x] The `Authorizable` interface is embedded into the `CrudRecord` type
 - [x] This interface defines the `GetOwner` / `SetOwner` type, which is an individual
   `UserId` corresponding to the `User` who created this record.
 - [ ] This `UserId` is generally intended to be immutable, but the `SysAdmin` role may
   have rights to change this field if need be
    - e.g. to reassign ownership of a `Season` to a new `Commissioner` during season play
    - However, even the `UserId` themselves that owns a record should not usually be able to do this
 - [x] The `EditableBy` field on the `Authorizable` interface allows the business logic to
   determine which `UserId` values can edit a `CrudRecord` at API-time
    - For example, this can be used in API middleware to prevent updates to a `Season` by non-commissioners   
 - [x] The `AccessibleTo` field on the `Authorizable` interface implements the corresponding logic for
    access rights.
    - E.g. some information like `Availability` is private from a viewing perspective to the `Team` that
      it pertains to only.
 - In general, the `GetOwner` / `SetOwner` logic is boilerplate in the database layer, but the logic
   in `EditableBy` and `AccessibleTo` is defined on a record-by-record basis.
    - This allows us to cleanly specify the business logic in-place w/ easy unit testing

## Record-specific rules
The types in the `model` directory implement all the specific `CrudRecord`s for the application

### Common access logic
- `captain_viewable`: a record that is `EditableBy` a `Season`'s `commissioner`, and viewable by the `Season`'s  `captain`s
   - examples are `Draft` and `CommissionerProposal`  
- `team_viewable`: a record that is viewable by members of a `Team` only
   - examples are `Availability` and `Lineup`s
- `user_modifiable_only`: a record that may only be modified by the record's owner
   - these are typically root types, e.g.  `facility`, `format`, `rating`, etc. that higher level types like `Draft` are composed of

### Base types
- `availability`: a `User`'s availability for a given `Week` in a `Season` 
   - `team_viewable` access rights
- `blurb`: a weekly summary or other league communication from the `Commissioner` or designated reporter for a `Season`
   - A `blurb` may only be created by a `Season`'s commissioner or designated `Reporter`(s)
   - A `blurb` may only be edited after-the-fact by the same
- `comment`: a `User`-supplied comment on a `blurb`
    - A `comment` may only be created by a player participating in a `Season`
    - A `comment` may be deleted (but not _edited_) by the `Commissioner` or `Reporter` on a season
- `commissioner_proposal`: a mid-`Season` proposal from the commissioner to captains for a rule change
    - `captain_viewable` access rules
- `facility`: a physical location that a `Season` is played out of
    - `user_modifiable_only` access rules
    - Cannot be deleted after it is in use somewhere
- `format`: a matchup format for a `Season` (e.g. the `rating`s and `line`s that are played)
    - `user_modifiable_only` access rules
    - Cannot be deleted (or modified) after it is in use somewhere
- `individual_match`: the results of a single match during `Season` play
    - this type is globally viewable without authorization
    - Can only be edited by designees (e.g. the `Match` participants, `captain`s, `commissioner`, etc.)
- `lineup`: a _planned_ set of players participating in a single `matchup` during `Season` play
    - `team_viewable` access rules
    - this type is copied into the globally-viewable `matchup` type before play begins
- `matchup`: the _real_ players participating in a single `matchup` during `Season` play
    - this type is globally viewable without authorization
    - Can only be edited by captains and commissioners after creation (e.g. for substitutions)
- `photo`: a photograph attached to a `blurb`
    - `user_modifiable_only` access rules  
- `playoff_structure`: a format for playoffs after the conclusion of the regular season portion of a `Season`
    - `user_modifiable_only` access rules
- `pre_draft_grade`: a `User` rating of another `User` for `Draft` pool calibration purposes
    - `user_modifiable_only` access rules
    - we might need to maintain a `draft_grader` assignment table to control who can create `pre_draft_grade`s for a `Draft`
- `rating`: a description type defining the different levels present in a given `Format`
    - `user_modifiable_only` access rules
    - globally accessible without authorization
    - cannot be deleted or modified after it's in-use
- `reaction`: a `User`-supplied emoji reaction to a `blurb` or a `comment` on a blurb
    - same rules as a `comment`
- `rule_amendment`: a modification to some portion of a `ruleset`
    - rule amendments can only be created by the owner of a `ruleset`
    - amendments are viewable in the same way as `ruleset`s
- `ruleset`: a set of league rules that is in-use for one or many `Season`s
    - these are composed of `ruleset_section`s with the same rights
    - `user_modifiable_only` access rules
    - generally these would be created by the `commissioner` of a `Season`, but not necessarily so
- `scoring_structure`: a scoring format for individual `match` play
    - `user_modifiable_only` access rules
    - can not be edited or deleted after creation
- `skill_info`: a `User`-supplied rating of ones own skill (e.g. from playing on an ALTA team or USTA self-rating)
    - `user_modifiable_only` access rights
- `user`: an individual person who uses the application and participates in `Season`s
    - `user_modifiable_only`
    - this is a very important base type for almost all access rules
- `week` a date-and-time where a week of play occurs during a `Season`
    - this is used as a reference type for `Season` `Matchup`s and individual `Lineup` and `Availability` records
    - only editable by the commissioner

### Meta / grouping types
- `draft`: a type corresponding to the actual draft event for a `Season`
    - `captain_viewable` access rules
- `schedule`: a list of `Matchup`s during a `Season`
    - editable only by `commissioner`, and likely not so after `Season` begins
    - viewable by all
- `season`: a complete season of play, the highest level grouping type in the application
    - editable only by the `commissioner`
    - this is created automatically when finalizing the `draft`
- `team`: an individual team competing in a given `Season`
    - members, roles and rating are publicly viewable
    - only editable by the commissioner
    - `co_captain`s can be managed by the `captain` of a `team
    - this grouping controls the `team_viewable` access-and-edit logic


### Join-table types
Join table types assign from one table to another and generally have their own boilerplate logic.

These types are not typically exposed as a raw `CrudRecord` type in the API but are rather viewed
and edited by business logic on another type. For example, the `TeamAssignment` type manages the 
`User`s who are members (and `captain`s / `co_captain`s) of a particular `Team`.

- `draft_captain`: an assignment of a `captain` to a `Team` inside a single `Draft` w/ draft position
- `draft_available_plater`: an assignment of a `User` as a draftable player in a `Draft` pool
- `draft_pick`: a selected player at a particular position in a `Draft`
- `season_commissioner`: the owner / administrator of a particular `Season`
- `season_late_add`: a late-added participant into a `season`
- `player`: a `User` who is participating as a player in a `season`
- `season_captain`: a `User` who is acting as a captain in a `draft` and subsequent `season`
- `team_match`: a home-and-away assignment of `Lineup`s in a given `Week` in a `Season`
- `weekly_matchup`: an assignment of which teams play who during a `Week` in a `Season`
