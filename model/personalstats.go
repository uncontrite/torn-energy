package model

type PersonalStats struct {
	AttacksWon int `json:"attackswon,omitempty"`
	DumpSearches int `json:"dumpsearches,omitempty"`
	UserActivity int `json:"useractivity,omitempty"`
	Logins int `json:"logins,omitempty"`
	AttacksLost int `json:"attackslost,omitempty"`
	XanaxTaken int `json:"xantaken,omitempty"`
	AttacksDraw int `json:"attacksdraw,omitempty"`
	LsdTaken int `json:"lsdtaken,omitempty"`
	EcstasyTaken int `json:"exttaken,omitempty"`
	Overdosed int `json:"overdosed,omitempty"`
	YouRunAway int `json:"yourunaway,omitempty"`
	AttacksAssisted int `json:"attacksassisted,omitempty"`
	CannabisTaken int `json:"cantaken,omitempty"`
	ConsumablesUsed int `json:"consumablesused,omitempty"`
	CandyUsed int `json:"candyused,omitempty"`
	AlcoholUsed int `json:"alcoholused,omitempty"`
	EnergyDrinkUsed int `json:"energydrinkused,omitempty"`
	BooksRead int `json:"booksread,omitempty"`
	NerveRefills int `json:"nerverefills,omitempty"`
	BoostersUsed int `json:"boostersused,omitempty"`
	Refills int `json:"refills,omitempty"`
}

func (ps PersonalStats) Diff(ps2 PersonalStats) PersonalStats {
	var diff PersonalStats
	diff.AttacksWon = ps2.AttacksWon - ps.AttacksWon
	diff.DumpSearches = ps2.DumpSearches - ps.DumpSearches
	diff.UserActivity = ps2.UserActivity - ps.UserActivity
	diff.Logins = ps2.Logins - ps.Logins
	diff.AttacksLost = ps2.AttacksLost - ps.AttacksLost
	diff.XanaxTaken = ps2.XanaxTaken - ps.XanaxTaken
	diff.AttacksDraw = ps2.AttacksDraw - ps.AttacksDraw
	diff.LsdTaken = ps2.LsdTaken - ps.LsdTaken
	diff.EcstasyTaken = ps2.EcstasyTaken - ps.EcstasyTaken
	diff.Overdosed = ps2.Overdosed - ps.Overdosed
	diff.YouRunAway = ps2.YouRunAway - ps.YouRunAway
	diff.AttacksAssisted = ps2.AttacksAssisted - ps.AttacksAssisted
	diff.CannabisTaken = ps2.CannabisTaken - ps.CannabisTaken
	diff.ConsumablesUsed = ps2.ConsumablesUsed - ps.ConsumablesUsed
	diff.CandyUsed = ps2.CandyUsed - ps.CandyUsed
	diff.AlcoholUsed = ps2.AlcoholUsed - ps.AlcoholUsed
	diff.EnergyDrinkUsed = ps2.EnergyDrinkUsed - ps.EnergyDrinkUsed
	diff.BooksRead = ps2.BooksRead - ps.BooksRead
	diff.NerveRefills = ps2.NerveRefills - ps.NerveRefills
	diff.BoostersUsed = ps2.BoostersUsed - ps.BoostersUsed
	diff.Refills = ps2.Refills - ps.Refills
	return diff
}

func (ps PersonalStats) IsDiffAttack() bool {
	return ps.AttacksWon > 0 || ps.AttacksLost > 0 || ps.AttacksDraw > 0 || ps.AttacksAssisted > 0 || ps.YouRunAway > 0
}

func (ps PersonalStats) IsDiffRelevant() []string {
	var reasons []string
	if ps.IsDiffAttack() {
		reasons = append(reasons, "attack")
	}
	if ps.DumpSearches > 0 {
		reasons = append(reasons, "dump")
	}
	if ps.LsdTaken > 0 {
		reasons = append(reasons, "lsd")
	}
	if ps.XanaxTaken > 0 {
		reasons = append(reasons, "xanax")
	}
	if ps.Overdosed > 0 {
		reasons = append(reasons, "od")
	}
	if ps.Refills > 0 {
		reasons = append(reasons, "psprf")
	}
	if ps.BooksRead > 0 {
		reasons = append(reasons, "book")
	}
	if ps.CannabisTaken > 0 || ps.EnergyDrinkUsed > 0 {
		reasons = append(reasons, "energydrink")
	}
	if ps.BoostersUsed > 0 {
		reasons = append(reasons, "booster")
	}
	if ps.ConsumablesUsed > 0 {
		reasons = append(reasons, "consumable")
	}
	return reasons
}