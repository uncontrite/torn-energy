package main

import "sort"

func (u User) Diff(u2 User) User {
	var diff User
	diff.BattleStats = u.BattleStats.Diff(u2.BattleStats)
	diff.Bars = u.Bars.Diff(u2.Bars)
	diff.PersonalStats = u.PersonalStats.Diff(u2.PersonalStats)
	diff.Jobs = Diff(u2.Jobs, u.Jobs)
	diff.Refills = u.Refills.Diff(u2.Refills)
	diff.UserId = u.UserId
	return diff
}

func (u User) IsDiffRelevant() ([]string, map[string]struct{}) {
	var reasons []string
	if u.IsTrain() {
		reasons = append(reasons, "train")
	}
	if u.Refills.EnergyRefillUsed || u.Refills.SpecialRefillsAvailable < 0 {
		reasons = append(reasons, "prf")
	}
	for _, j := range u.Jobs {
		if j.Points < 0 {
			reasons = append(reasons, "jp")
			break
		}
	}
	reasons = append(reasons, u.PersonalStats.IsDiffRelevant()...)
	sort.SliceStable(reasons, func(i, j int) bool {
		return reasons[i] < reasons[j]
	})
	m := make(map[string]struct{})
	for _, r := range reasons {
		m[r] = struct{}{}
	}
	return reasons, m
}

func (u User) IsTrain() bool {
	return u.BattleStats.IsTrain()
}

// returns (fhc, edvd)
func CalculateBoosterSplit(prevHappy int, currHappy int, ecstasyTaken int, boostersTaken int) (int, int) {
	var edvd int
	var fhc int
	boosters := boostersTaken
	if boosters > 0 {
		happy := currHappy
		if ecstasyTaken > 0 {
			happy /= 2
		}
		happy -= prevHappy
		// 400f + 2400e = h
		// f + e = b => 400f + 400e = 400b
		// --------
		// 2000e = h - 400b
		// e = (h-400b)/2000
		edvd = (happy - (400 * boostersTaken)) / 2000
		fhc = boostersTaken - edvd
	}
	return fhc, edvd
}

func CalculateEnergyTrained(prev User, curr User) int {
	diff := prev.Diff(curr)
	if !diff.IsTrain() {
		return 0
	}
	prfEnergy := diff.PersonalStats.Refills * prev.Bars.Energy.Maximum
	xanEnergy := 250 * diff.PersonalStats.XanaxTaken
	lsdEnergy := 50 * diff.PersonalStats.LsdTaken
	ps := diff.PersonalStats
	attacks := ps.AttacksWon + ps.AttacksLost + ps.AttacksDraw + ps.AttacksAssisted + ps.YouRunAway
	attacksEnergy := -25 * attacks
	unspentEnergy := -1 * curr.Bars.Energy.Current
	dumpEnergy := -5 * ps.DumpSearches
	energyDrinkEnergy := 30 * ps.EnergyDrinkUsed
	// Heuristic to split Booster into FHC v. EDVD
	fhc, _ := CalculateBoosterSplit(prev.Bars.Happy.Current, curr.Bars.Happy.Current, diff.PersonalStats.EcstasyTaken,
		diff.PersonalStats.BoostersUsed)
	fhcEnergy := 150 * fhc

	var gameShop, candle, farm, furniture, pub, restaurant int
	for _, j := range diff.Jobs {
		if j.Points > 0 {
			continue
		}
		if j.Name == "Game Shop" {
			gameShop = -1 * j.Points
		} else if j.Name == "Candle Shop" {
			candle = -1 * j.Points
		} else if j.Name == "Farm" {
			farm = -1 * j.Points
		} else if j.Name == "Furniture Store" {
			furniture = -1 * j.Points
		} else if j.Name == "Pub" {
			pub = -1 * j.Points
		} else if j.Name == "Restaurant" {
			restaurant = -1 * j.Points
		}
	}
	jobEnergy := (5 * gameShop) + (5 * candle) + (7 * farm) + (3 * furniture) + (3 * pub) + (3 * restaurant)
	eTrained := prev.Bars.Energy.Current + prfEnergy + xanEnergy + lsdEnergy + unspentEnergy + attacksEnergy +
		dumpEnergy + energyDrinkEnergy + fhcEnergy + jobEnergy
	return eTrained
}