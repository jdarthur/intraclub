package model

import (
	"fmt"
	"testing"
)

func DumpDraftPicks(p DraftOrderPattern, numberOfCaptains, numberOfPicks int) {
	rounds := numberOfPicks / numberOfCaptains
	if numberOfPicks%numberOfCaptains != 0 {
		rounds += 1
	}

	for i := 0; i < rounds; i++ {
		for j := 0; j < numberOfCaptains; j++ {
			round := i + 1
			pick := j + 1
			captainIndex := p.GetCaptainOnTheClock(round, pick, numberOfCaptains)
			fmt.Printf("%d.%d: team %d\n", round, pick, captainIndex+1)
		}
		fmt.Println()
	}

}

func assertCaptainOnTheClock(t *testing.T, pattern DraftOrderPattern, round, pick, numberOfCaptains, expectedIndex int) {
	p := pattern.GetCaptainOnTheClock(round, pick, numberOfCaptains)
	if p != expectedIndex {
		t.Fatalf("%d.%d should be captain index %d (got %d)", round, pick, expectedIndex, p)
	}
}

func TestSnakeDraft1(t *testing.T) {

	pattern := DraftOrderPatternSnake{}

	assertCaptainOnTheClock(t, pattern, 1, 1, 4, 0)
	assertCaptainOnTheClock(t, pattern, 1, 2, 4, 1)
	assertCaptainOnTheClock(t, pattern, 1, 3, 4, 2)
	assertCaptainOnTheClock(t, pattern, 1, 4, 4, 3)

	assertCaptainOnTheClock(t, pattern, 2, 1, 4, 3)
	assertCaptainOnTheClock(t, pattern, 2, 2, 4, 2)
	assertCaptainOnTheClock(t, pattern, 2, 3, 4, 1)
	assertCaptainOnTheClock(t, pattern, 2, 4, 4, 0)

	assertCaptainOnTheClock(t, pattern, 3, 1, 4, 0)
	assertCaptainOnTheClock(t, pattern, 3, 2, 4, 1)
	assertCaptainOnTheClock(t, pattern, 3, 3, 4, 2)
	assertCaptainOnTheClock(t, pattern, 3, 4, 4, 3)

	DumpDraftPicks(pattern, 4, 12)
}

func TestLastPickDoubleDraft(t *testing.T) {
	pattern := DraftOrderPatternLastPickDouble{}

	assertCaptainOnTheClock(t, pattern, 1, 1, 4, 0)
	assertCaptainOnTheClock(t, pattern, 1, 2, 4, 1)
	assertCaptainOnTheClock(t, pattern, 1, 3, 4, 2)
	assertCaptainOnTheClock(t, pattern, 1, 4, 4, 3)

	assertCaptainOnTheClock(t, pattern, 2, 1, 4, 3)
	assertCaptainOnTheClock(t, pattern, 2, 2, 4, 0)
	assertCaptainOnTheClock(t, pattern, 2, 3, 4, 1)
	assertCaptainOnTheClock(t, pattern, 2, 4, 4, 2)

	assertCaptainOnTheClock(t, pattern, 3, 1, 4, 2)
	assertCaptainOnTheClock(t, pattern, 3, 2, 4, 3)
	assertCaptainOnTheClock(t, pattern, 3, 3, 4, 0)
	assertCaptainOnTheClock(t, pattern, 3, 4, 4, 1)

	DumpDraftPicks(pattern, 4, 48)
}

func TestStraightUpDraft(t *testing.T) {
	pattern := DraftOrderPatternStraightUp{}

	assertCaptainOnTheClock(t, pattern, 1, 1, 4, 0)
	assertCaptainOnTheClock(t, pattern, 1, 2, 4, 1)
	assertCaptainOnTheClock(t, pattern, 1, 3, 4, 2)
	assertCaptainOnTheClock(t, pattern, 1, 4, 4, 3)

	assertCaptainOnTheClock(t, pattern, 2, 1, 4, 0)
	assertCaptainOnTheClock(t, pattern, 2, 2, 4, 1)
	assertCaptainOnTheClock(t, pattern, 2, 3, 4, 2)
	assertCaptainOnTheClock(t, pattern, 2, 4, 4, 3)

	assertCaptainOnTheClock(t, pattern, 3, 1, 4, 0)
	assertCaptainOnTheClock(t, pattern, 3, 2, 4, 1)
	assertCaptainOnTheClock(t, pattern, 3, 3, 4, 2)
	assertCaptainOnTheClock(t, pattern, 3, 4, 4, 3)

	DumpDraftPicks(pattern, 4, 48)
}
