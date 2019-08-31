package model

import (
	"fmt"
	"sort"
)

type UserDiff struct {
	User
	MaxEnergy int `json:"maxEnergy,omitempty"`
}

func (u User) Diff(u2 User) UserDiff {
	var diff UserDiff
	diff.BattleStats = u.BattleStats.Diff(u2.BattleStats)
	diff.Bars = u.Bars.Diff(u2.Bars)
	diff.PersonalStats = u.PersonalStats.Diff(u2.PersonalStats)
	diff.Jobs = Diff(u2.Jobs, u.Jobs)
	diff.Refills = u.Refills.Diff(u2.Refills)
	diff.UserId = u.UserId
	diff.MaxEnergy = u.Bars.Energy.Maximum
	return diff
}

func (u UserDiff) IsRelevant() ([]string, map[string]struct{}) {
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

func (u UserDiff) GetEvents() []string {
	var events []string
	if psEvents := u.PersonalStats.GetEvents(); len(psEvents) > 0 {
		events = append(events, psEvents...)
	}
	if jpEnergyGained, jpSpent := u.CalculateEnergyGainedFromJobPoints(); jpEnergyGained > 0 {
		events = append(events, fmt.Sprintf("gained %de by spending %d job points", jpEnergyGained, jpSpent))
	}
	fhc, edvd := CalculateBoosterSplit(u.Bars.Happy.Previous, u.Bars.Happy.Current, u.PersonalStats.EcstasyTaken,
		u.PersonalStats.BoostersUsed, u.PersonalStats.Overdosed, u.Bars.Energy.Current, u.IsTrain())
	if fhc > 0 {
		events = append(events, fmt.Sprintf("gained %de* by using %d FHCs", 150 * fhc, fhc))
	}
	if edvd > 0 {
		events = append(events, fmt.Sprintf("gained %d happy by watching %d eDVDs", edvd * 2500, edvd))
	}
	gains := u.BattleStats.GetTotalGains()
	trained := u.CalculateEnergyTrained()
	if t, _ := gains.Float64(); t > 0 || trained > 0 {
		events = append(events, fmt.Sprintf("trained %de gaining %s stats", trained, gains.Text('f', 4)))
	}
	return events
}

func (u UserDiff) IsTrain() bool {
	return u.BattleStats.IsTrain()
}

// returns (fhc, edvd)
func CalculateBoosterSplit(prevHappy int, currHappy int, ecstasyTaken int, boostersTaken int, od int, currEnergy int, train bool) (int, int) {
	var edvd int
	var fhc int
	boosters := boostersTaken
	if boosters > 0 {
		if currEnergy >= 400 { // Power training
			return 0, boostersTaken
		}
		happy := currHappy
		if ecstasyTaken > 0 {
			happy /= 2
		}
		happy -= prevHappy
		if happy <= -5000 && currHappy > 10 && od == 0 { // Happy reset
			if train {
				return boostersTaken, 0
			}
			return 0, boostersTaken
		}
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

func (u UserDiff) CalculateEnergyGainedFromJobPoints() (int, int) {
	var gameShop, candle, farm, furniture, pub, restaurant int
	for _, j := range u.Jobs {
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
	pointsSpent := gameShop + candle + farm + furniture + pub + restaurant
	jobEnergy := (5 * gameShop) + (5 * candle) + (7 * farm) + (3 * furniture) + (3 * pub) + (3 * restaurant)
	return jobEnergy, pointsSpent
}

func (u UserDiff) CalculateEnergyTrained() int {
	if !u.IsTrain() {
		return 0
	}
	prfEnergy := u.PersonalStats.Refills * u.MaxEnergy
	xanEnergy := 250 * u.PersonalStats.XanaxTaken
	lsdEnergy := 50 * u.PersonalStats.LsdTaken
	ps := u.PersonalStats
	attacks := ps.AttacksWon + ps.AttacksLost + ps.AttacksDraw + ps.AttacksAssisted + ps.YouRunAway
	attacksEnergy := -25 * attacks
	dumpEnergy := -5 * ps.DumpSearches
	energyDrinkEnergy := 30 * ps.EnergyDrinkUsed
	// Heuristic to split Booster into FHCs v. EDVDs
	fhc, _ := CalculateBoosterSplit(u.Bars.Happy.Previous, u.Bars.Happy.Current, u.PersonalStats.EcstasyTaken,
		u.PersonalStats.BoostersUsed, u.PersonalStats.Overdosed, u.Bars.Energy.Current, u.IsTrain())
	fhcEnergy := 150 * fhc

	unspentEnergy := -1 * u.Bars.Energy.Current
	jpEnergy, _ := u.CalculateEnergyGainedFromJobPoints()
	eTrained := u.Bars.Energy.Previous + prfEnergy + xanEnergy + lsdEnergy + unspentEnergy + attacksEnergy +
		dumpEnergy + energyDrinkEnergy + fhcEnergy + jpEnergy
	return eTrained
}

type UserSummary struct {
	User          uint
	Name          string
	Energy        int
	FHCs          int
	Xanax         int
	LSD           int
	EnergyDrinks  int
	Attacks       int
	EnergyRefills int
	EDVDs         int
	Dumps         int
	JpEnergy	  int
	Overdoses	  int
}

func (u UserDiff) AddToSummary(summary *UserSummary) {
	summary.EnergyRefills += u.PersonalStats.Refills
	summary.Xanax += u.PersonalStats.XanaxTaken
	summary.LSD += u.PersonalStats.LsdTaken
	ps := u.PersonalStats
	summary.Attacks += ps.AttacksWon + ps.AttacksLost + ps.AttacksDraw + ps.AttacksAssisted + ps.YouRunAway
	summary.Dumps += ps.DumpSearches
	summary.EnergyDrinks += ps.EnergyDrinkUsed

	// Heuristic to split Booster into FHCs v. EDVDs
	fhc, edvd := CalculateBoosterSplit(u.Bars.Happy.Previous, u.Bars.Happy.Current, u.PersonalStats.EcstasyTaken,
		u.PersonalStats.BoostersUsed, u.PersonalStats.Overdosed, u.Bars.Energy.Current, u.IsTrain())
	summary.Overdoses += ps.Overdosed
	summary.FHCs += fhc
	summary.EDVDs += edvd
	jpEnergy, _ := u.CalculateEnergyGainedFromJobPoints()
	summary.JpEnergy += jpEnergy
	summary.Energy += u.CalculateEnergyTrained()
}