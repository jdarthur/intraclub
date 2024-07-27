package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
)

var tempKey = "5df41d43-5893-4101-8126-b5148e5f3185"

func DummyHome() *MatchScores {
	m := NewMatchScores()

	m.OneOne.Pairing.Player1.Name = "Clay DeFriece"
	m.OneOne.Pairing.Player2.Name = "Austin Reynolds"

	m.OneTwo.Pairing.Player1.Name = "Michael Bulostin"
	m.OneTwo.Pairing.Player2.Name = "Chris Wilson"

	m.OneThree.Pairing.Player1.Name = "JD Arthur"
	m.OneThree.Pairing.Player2.Name = "Jake Maloch"

	m.TwoTwo.Pairing.Player1.Name = "Hayden Van Dyke"
	m.TwoTwo.Pairing.Player2.Name = "Connor DelPrete"

	m.TwoThree.Pairing.Player1.Name = "Andy Lascik"
	m.TwoThree.Pairing.Player2.Name = "Eli Cohen"

	m.ThreeThree.Pairing.Player1.Name = "Don Schmal"
	m.ThreeThree.Pairing.Player2.Name = "Scott Chenoweth"

	return m
}

func DummyAway() *MatchScores {
	m := NewMatchScores()

	m.OneOne.Pairing.Player1.Name = "Ethan Moland"
	m.OneOne.Pairing.Player2.Name = "Sean Pulanski"

	m.OneTwo.Pairing.Player1.Name = "Josh Turknett"
	m.OneTwo.Pairing.Player2.Name = "Sean Connelly"

	m.OneThree.Pairing.Player1.Name = "Tomer Wagshal"
	m.OneThree.Pairing.Player2.Name = "Justin Chan"

	m.TwoTwo.Pairing.Player1.Name = "Jim Bernard"
	m.TwoTwo.Pairing.Player2.Name = "Dave Lindsay"

	m.TwoThree.Pairing.Player1.Name = "Dan Huber"
	m.TwoThree.Pairing.Player2.Name = "Ami Busel"

	m.ThreeThree.Pairing.Player1.Name = "Jim Byrd"
	m.ThreeThree.Pairing.Player2.Name = "Jonathan Link"

	return m
}

// MatchScores is a collection of all the Matchup s for a particular team,
// i.e. a collection of six Pairing s each of which has a set of MatchSetScores
type MatchScores struct {
	OneOne     Matchup `json:"one_one"`
	OneTwo     Matchup `json:"one_two"`
	OneThree   Matchup `json:"one_three"`
	TwoTwo     Matchup `json:"two_two"`
	TwoThree   Matchup `json:"two_three"`
	ThreeThree Matchup `json:"three_three"`
}

func NewMatchScores() *MatchScores {
	return &MatchScores{
		OneOne:     Matchup{Pairing: NewPairing(1, 1)},
		OneTwo:     Matchup{Pairing: NewPairing(1, 2)},
		OneThree:   Matchup{Pairing: NewPairing(1, 3)},
		TwoTwo:     Matchup{Pairing: NewPairing(2, 2)},
		TwoThree:   Matchup{Pairing: NewPairing(2, 3)},
		ThreeThree: Matchup{Pairing: NewPairing(3, 3)},
	}
}

// GetMatchup returns the particular Matchup in the MatchScores struct
// given a line key, e.g. "1-1"
func (m *MatchScores) GetMatchup(line string) Matchup {
	if line == "1-1" {
		return m.OneOne
	} else if line == "1-2" {
		return m.OneTwo
	} else if line == "1-3" {
		return m.OneThree
	} else if line == "2-2" {
		return m.TwoTwo
	} else if line == "2-3" {
		return m.TwoThree
	} else if line == "3-3" {
		return m.ThreeThree
	} else {
		fmt.Printf("invalid line: %s\n", line)
		return Matchup{}
	}
}

// SetMatchup sets the particular Matchup in the MatchScores struct
// for a given line key, e.g. setting new values for the player names
// or game scores for the line "1-1"
func (m *MatchScores) SetMatchup(line string, matchup Matchup) {
	if line == "1-1" {
		m.OneOne = matchup
	} else if line == "1-2" {
		m.OneTwo = matchup
	} else if line == "1-3" {
		m.OneThree = matchup
	} else if line == "2-2" {
		m.TwoTwo = matchup
	} else if line == "2-3" {
		m.TwoThree = matchup
	} else if line == "3-3" {
		m.ThreeThree = matchup
	} else {
		fmt.Printf("invalid line: %s\n", line)
	}
}

// Matchup is a container for a collection of MatchSetScores
// that a particular Pairing has at any given time in a match
type Matchup struct {
	Pairing   Pairing        `json:"pairing"`
	SetScores MatchSetScores `json:"set_scores"`
}

// Pairing is a group of two Player s who are playing a match
type Pairing struct {
	Player1 Player `json:"player1"`
	Player2 Player `json:"player2"`
}

func NewPairing(line1, line2 int) Pairing {
	return Pairing{
		Player1: NewPlayer(line1),
		Player2: NewPlayer(line2),
	}
}

// Player is a person with a Name who is playing at a particular Line
type Player struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

func NewPlayer(line int) Player {
	return Player{
		Name: "",
		Line: line,
	}
}

// MatchSetScores stores the number of games won by a Pairing in a particular match
type MatchSetScores struct {
	Set1Games int `json:"set1_games"`
	Set2Games int `json:"set2_games"`
	Set3Games int `json:"set3_games"`
}

// FullScores stores the MatchScores for both the Home and Away teams.
type FullScores struct {
	Home     *MatchScores `json:"home_scores"`
	Away     *MatchScores `json:"away_scores"`
	HomeTeam Team         `json:"home_team"`
	AwayTeam Team         `json:"away_team"`
}

// full is the in-memory running tally of all the matchups in the active match
var full = FullScores{
	Home:     DummyHome(),
	Away:     DummyAway(),
	HomeTeam: Team{Name: "Blue", Color: "blue"},
	AwayTeam: Team{Name: "Green", Color: "green"},
}

func GetMatchScoreboard(c *gin.Context) {
	c.JSON(200, full)
}

// MatchScoreRequest is a PUT request to update the MatchSetScores
// for the Home or away team at a particular Line
type MatchScoreRequest struct {
	Line   string         `json:"line"`
	Home   bool           `json:"home"`
	Scores MatchSetScores `json:"scores"`
}

// getPairing returns the Pairing in the full datatype for either the
// home or away team at a particular line
func getPairing(line string, home bool) Pairing {
	if home {
		return full.Home.GetMatchup(line).Pairing
	} else {
		return full.Away.GetMatchup(line).Pairing
	}
}

// validateTempKey makes sure that an update request contains the
// tempKey value as a query param (so that PUT requests can only be sent
// by an authorized user who has the tempKey value)
func validateTempKey(c *gin.Context) error {
	key, ok := c.GetQuery("key")
	if !ok {
		return errors.New("missing key param")
	}
	if key != tempKey {
		return errors.New("invalid key param")
	}
	return nil
}

// UpdateMatchScore is called when we update the game scores for a particular
// set for one team in an individual match
func UpdateMatchScore(c *gin.Context) {
	err := validateTempKey(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	request := MatchScoreRequest{}
	err = c.BindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	pairing := getPairing(request.Line, request.Home)

	m := Matchup{
		Pairing:   pairing,
		SetScores: request.Scores,
	}

	if request.Home {
		full.Home.SetMatchup(request.Line, m)
	} else {
		full.Away.SetMatchup(request.Line, m)
	}

	UpdateOccurred()

	c.JSON(200, full)
}

// MatchNameRequest is supplied in a PUT request to update the name
// of one of the Player s in a particular match
type MatchNameRequest struct {
	Line    string `json:"matchup_line"`
	Home    bool   `json:"home"`
	Name    string `json:"name"`
	Player1 bool   `json:"player1"`
}

// UpdateMatchNames is called when we update the name for a particular
// Player for one team in an individual match
func UpdateMatchNames(c *gin.Context) {

	err := validateTempKey(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// validate that we got a good request
	request := MatchNameRequest{}
	err = c.BindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
	}

	pairing := getPairing(request.Line, request.Home)
	fmt.Printf("Pairing: %+v\n", pairing)

	if request.Player1 {
		pairing.Player1 = Player{Name: request.Name, Line: pairing.Player1.Line}
	} else {
		pairing.Player2 = Player{Name: request.Name, Line: pairing.Player2.Line}
	}

	m := full.Home.GetMatchup(request.Line)
	m.Pairing = pairing

	if request.Home {
		full.Home.SetMatchup(request.Line, m)
	} else {
		full.Away.SetMatchup(request.Line, m)
	}

	UpdateOccurred()

	c.JSON(200, full)
}

type Team struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type UpdateTeamRequest struct {
	Home  bool   `json:"home"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func UpdateMatchTeam(c *gin.Context) {
	err := validateTempKey(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	request := UpdateTeamRequest{}
	err = c.BindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if request.Home {
		full.HomeTeam.Color = request.Color
		full.HomeTeam.Name = request.Name
	} else {
		full.AwayTeam.Color = request.Color
		full.AwayTeam.Name = request.Name
	}

	UpdateOccurred()

	c.JSON(200, full)
}
