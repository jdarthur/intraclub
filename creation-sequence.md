# Creation Sequence

## 1) Create base objects before a draft

### a) Create a user
Each individual person who wants to use the app has to do this.
You must have a user to create anything else

 - [x] Endpoint to self-register
 - [ ] Endpoint to get all existing users
 - [ ] Self-registration should fail if email already exists

### b) Create user tokens
Tokens are necessary to call authenticated endpoints such as create, update, delete
 - [ ] "Start login" endpoint should create an authentication token and email it to a user
 - [x] "Start login" endpoint should fail if a user doesn't already exist
 - [x] Token endpoint should require a one-time-use token / "magic link"
 - [x] Token endpoint should return a JWT signed by local keypair
 - [x] Token should be validate-able on subsequent requests
 - [x] Token endpoint should fail if one-time-use token is already used
 - [x] Token endpoint should fail if one-time-use token is not valid (different error)

### c) Create a facility
 - [x] Only one person has to create the Martin's Landing River Club facility
 - [x] Formats may not overlap
   - [x] uniqueness on name
   - [x] uniqueness on address
 - [x] Formats that are assigned to a season may not 

### d) Create other users
You must have some users to be able to create a draft (at least like 8 or so users)
 - [x] Endpoint to create a user from CSV
 - [x] Created users should not belong to CSV endpoint caller
 - [ ] CSV endpoint should parse already-created users so they are only modified (if necessary)
 - [ ] CSV endpoint should return errors, new users, modified users, & created users

### e) Create rating types
This is most likely just essentially `1`, `2`, and `3` and is needed before we can create a format
- [x] create a new rating
- [x] validate that the rating has values + a real User ID
- [x] validate that the rating is update-able by system administrator
- [x] validate that the rating cannot be deleted if it is in use

### f) Create a format
This will divide the drafted players into groups. It is a named list of ratings, high-skill-level to low-skill-level
- [x] create a new format
- [x] validate that the format has a name + a real UserId
- [x] validate that the format has some valid possible ratings
- [x] validate that the format has at least one valid line
- [x] validate that all rating IDs in each line are located in the possible ratings list
- [x] validate that the format cannot be deleted if it is in-use by a season
- [x] validate that the format cannot be edited if it is in-use by a season
- [x] validate that the format has no duplicate lines 
- [x] validate that the format has no reversed duplicate lines

### g) Create a playoff structure

- This only needs to be created once. An example format could be "3 teams make the playoffs, 1 team gets a bye"
- [x] validate that user ID is a real user
- [x] validate that no-bye setup has a power-of-2 number of teams
- [x] validate that in-use playoff structure cannot be deleted
- [x] validate that in-use playoff structure cannot be edited
- [x] validate that first round w/ byes includes an even number of teams
- [x] validate that second round w/ byes is a power-of-two number of teams
- [x] validate that number of teams must be at least two
- [x] validate that number of byes can't be >= number of teams

## 2) Draft players to teams

### a) Create a draft
- [x] Initialize a draft with a list of captain IDs
- [x] Each captain ID must be a valid user ID
- [x] Format must be assigned
- [x] List of players must be all real players
- [x] Selections must be empty when initializing
- [x] Captains can only pick when it's their turn
- [x] Captains can only be selected by themselves
- [x] Players cannot be double-selected
- [x] Players not in draft list cannot be selected

### b) Assign some draft-able players to the draft
You should have an even-ish number to divide among teams, e.g. 10 players per team
- [x] You can add new players to the available list
- [x] You can re-add players without changing the available list

### c) Assign pre-draft grades to players in the draft
Grading system where players may be marked as a strong, average, or weak version
of a particular `Rating`, e.g a high 1 or a mid 3

- [x] player ID must be valid
- [x] Draft ID must be valid
- [x] Grader ID must be a valid user
- [x] Player ID must be draftable in `Draft`
- [x] Rating must be present in `Draft`'s format
- [x] +/- modifier must be 0, 1, or 2
- [x] draft must not be completed when creating or updating a grade


### d) Assign some captains to each team in the draft
- [x] users must exist for all captains
- [ ] captain IDs must all be distinct


### e) Do the draft

- Owner of the draft (the commissioner) must start the draft
- Owner of the draft must assign a draft order
- Each captain can select a player when draft is started and it's their turn

### f) Complete the draft
- [x] Mark the draft as closed after the last pick
- [x] Create a season from the draft results with the assigned players and captains

## Configure a season

### a) Set up base values

- [x] Assign a name, facility, and start time
- [ ] Add co-commissioners if desired

### b) Create a list of weeks
- [ ] Each week must be in the future
- [ ] Each week must point to a valid `Draft`
   - the Availability records for a user can be created before the Season
   - each `Season` will have a one-to-one relationship with a `Draft`
- [ ] Weeks for a `Season` / `Draft` can be queried 
- Week records should get actually created right before configuration `POST`

### c) Create a season schedule

- Assign team matchups for each week in the season
- Assign a playoff structure

## Set up teams
### a) Notify players of their teams 
- We can probably automatically email the users
### b) Assign co-captains
- Each team captain can do this if they want
### c) Input availability
- Each team member can do this

## Set lineups and play a match
### Set lineups
- Team captain (or co-captain) can configure a weekly lineup based on player's ratings and format
- Lineup must be confirmed before the commissioner marks it as "official"
### Play matches
- Record game scores as they go on
### Determine a winner
- Commissioner can close a week when all matches are complete
- Use the game scores w/ rules to determine who wins a week head-to-head and overall
- Update weekly standings
