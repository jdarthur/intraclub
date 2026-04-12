package model

import (
	"fmt"
	"net/http"

	"intraclub/common"

	"github.com/gin-gonic/gin"
)

type DraftOrderPattern interface {
	GetCaptainOnTheClock(round, pick, numberOfCaptains int) (captainIndex int)
	Name() string
	Description() string
}

type DraftOrderPatternSerial struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Example     [][]int `json:"example"`
}

var DraftOrderPatterns = []DraftOrderPattern{
	DraftOrderPatternSnake{},
	DraftOrderPatternLastPickDouble{},
	DraftOrderPatternStraightUp{},
}

func DraftOrderPatternFromString(s string) (DraftOrderPattern, error) {
	for _, pattern := range DraftOrderPatterns {
		if s == pattern.Name() {
			return pattern, nil
		}
	}
	return nil, fmt.Errorf("invalid draft order pattern: %s", s)
}

func GetDraftOrderPatternExample(d DraftOrderPattern, numberOfCaptains, numberOfRounds int) [][]int {
	output := make([][]int, 0, numberOfRounds)
	for i := 1; i <= numberOfRounds; i++ {
		round := make([]int, 0, numberOfCaptains)
		for j := 1; j <= numberOfCaptains; j++ {
			c := d.GetCaptainOnTheClock(i, j, numberOfCaptains)
			round = append(round, c+1)
		}
		output = append(output, round)
	}
	return output
}

func GetDraftOrderPatterns(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{common.ResourceKey: getDraftOrderPatterns()})
}

func getDraftOrderPatterns() []DraftOrderPatternSerial {
	output := make([]DraftOrderPatternSerial, 0, len(DraftOrderPatterns))
	for _, pattern := range DraftOrderPatterns {
		output = append(output, DraftOrderPatternSerial{
			Name:        pattern.Name(),
			Description: pattern.Description(),
			Example:     GetDraftOrderPatternExample(pattern, 4, 5),
		})
	}
	return output
}

type DraftOrderPatternSnake struct{}

func (d DraftOrderPatternSnake) Description() string {
	return "Order reverses each round"
}

func (d DraftOrderPatternSnake) Name() string {
	return "Snake"
}

func (d DraftOrderPatternSnake) GetCaptainOnTheClock(round, pick, numberOfCaptains int) (captainIndex int) {
	// if this is an even round, we draft in reverse order (snake draft)
	if round%2 == 0 {
		return numberOfCaptains - pick
	}

	// otherwise we draft in the order of the TeamCaptainAssignment
	return pick - 1
}

type DraftOrderPatternLastPickDouble struct{}

func (d DraftOrderPatternLastPickDouble) Name() string {
	return "Last pick double"
}

func (d DraftOrderPatternLastPickDouble) Description() string {
	return "Last pick of each round gets a back-to-back, then the furthest-away team picks next"
}

func (d DraftOrderPatternLastPickDouble) GetCaptainOnTheClock(round, pick, numberOfCaptains int) (captainIndex int) {

	if round == 1 {
		return pick - 1
	}

	if pick == 1 {
		return d.GetCaptainOnTheClock(round-1, numberOfCaptains, numberOfCaptains)
	}
	return d.GetCaptainFurthestAway(round, pick, numberOfCaptains)
}

func (d DraftOrderPatternLastPickDouble) GetCaptainBefore(currentRound, currentPick, numberOfCaptains, distance int) (captainIndex int) {
	newRound := currentRound
	newPick := currentPick

	for i := 0; i < distance; i++ {
		newPick -= 1
		if newPick == 0 {
			newRound -= 1
			newPick = numberOfCaptains
		}
	}
	return d.GetCaptainOnTheClock(newRound, newPick, numberOfCaptains)
}

func (d DraftOrderPatternLastPickDouble) GetCaptainFurthestAway(round, pick, numberOfCaptains int) (captainIndex int) {
	return d.GetCaptainBefore(round, pick, numberOfCaptains, numberOfCaptains+1)
}

type DraftOrderPatternStraightUp struct{}

func (d DraftOrderPatternStraightUp) Name() string {
	return "Straight-up"
}

func (d DraftOrderPatternStraightUp) Description() string {
	return "Teams pick in the same order each round"
}

func (d DraftOrderPatternStraightUp) GetCaptainOnTheClock(round, pick, numberOfCaptains int) (captainIndex int) {
	return pick - 1
}
