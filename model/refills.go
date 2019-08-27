package model

type Refills struct {
	EnergyRefillUsed bool `json:"energy_refill_used,omitempty"`
	SpecialRefillsAvailable int `json:"special_refills_available,omitempty"`
}

func (r Refills) Diff(r2 Refills) Refills {
	var diff Refills
	// True when user uses energy refill for the day, false otherwise
	diff.EnergyRefillUsed = !r.EnergyRefillUsed && r2.EnergyRefillUsed
	// _Decreases_ when used, i.e. negative diff means special refill was used
	diff.SpecialRefillsAvailable = r2.SpecialRefillsAvailable - r.SpecialRefillsAvailable
	return diff
}