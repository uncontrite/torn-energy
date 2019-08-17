package main

type Refills struct {
	EnergyRefillUsed bool `json:"energy_refill_used,omitempty"`
	SpecialRefillsAvailable uint `json:"special_refills_available,omitempty"`
}