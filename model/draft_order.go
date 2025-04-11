package model

type DraftOrderPattern interface {
	GetCaptainOnTheClock(round, pick, numberOfCaptains int) (captainIndex int)
}

type DraftOrderPatternSnake struct{}

func (d DraftOrderPatternSnake) GetCaptainOnTheClock(round, pick, numberOfCaptains int) (captainIndex int) {
	// if this is an even round, we draft in reverse order (snake draft)
	if round%2 == 0 {
		return numberOfCaptains - pick
	}

	// otherwise we draft in the order of the TeamCaptainAssignment
	return pick - 1
}

type DraftOrderPatternLastPickDouble struct{}

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

func (d DraftOrderPatternStraightUp) GetCaptainOnTheClock(round, pick, numberOfCaptains int) (captainIndex int) {
	return pick - 1
}
