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
 - [x] Facilities may not overlap
   - e.g uniqueness on name, address fields
 - [x] Facilities that are assigned to a season may not 

### d) Create other users
You must have some users to be able to create a draft (at least like 8 or so users)
 - [x] Endpoint to create a user from CSV
 - [x] Created users should not belong to CSV endpoint caller
 - [ ] CSV endpoint should parse already-created users so they are only modified (if necessary)
 - [ ] CSV endpoint should return errors, new users, modified users, & created users

### e) Create rating types

- This is most likely just `1`, `2`, and `3`
- This is needed before we can create a format

### f) Create a format

- This will divide the drafted players into Ratings
- It is a list of ratings, high-skill-level to low-skill-level

### g) Create a playoff structure

- This only needs to be created once
- e.g. 3 teams make the playoffs, 1 team gets a bye

## 2) Draft players to teams

### a) Create a draft

- You must have a certain number of teams in mind
- Draft creation will create a list of Teams without an assigned captain
- Draft owner will be the first season commissioner
- Format must be assigned to partition players to a rating based on draft order

### b) Assign some draft-able players to the draft

- You should have an even-ish number to divide among teams, e.g. 10 players per team

### c) Assign pre-draft grades to players in the draft

- Grader

### d) Assign some captains to each team in the draft

- Users must exist already for each captain

### e) Do the draft

- Owner of the draft (the commissioner) must start the draft
- Owner of the draft must assign a draft order
- Each captain can select a player when draft is started and it's their turn

### f) Complete the draft

- Mark the draft as closed
- Create a season from the draft results with the assigned players and captains

## Configure a season

### a) Set up base values

- Assign a name, facility, and start time
- Add co-commissioners if desired

### b) Create a list of weeks

- Each week must be in the future
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
