package main

import "encoding/json"

type AttacksResponse struct {
	Attacks Attacks `json:"attacks,omitempty"`
}

type Attacks map[string]*json.RawMessage

type Attack struct {
	Started uint `json:"timestamp_started,omitempty"`
	Ended uint `json:"timestamp_ended,omitempty"`
	Id uint `json:"id,omitempty,omitempty"`
	AttackerId uint `json:"attacker_id,omitempty"`
	AttackerName string `json:"attacker_name,omitempty"`
	AttackerFactionId uint `json:"attacker_faction,omitempty"`
	AttackerFactionName string `json:"attacker_factionname,omitempty"`
	DefenderId uint `json:"defender_id,omitempty"`
	DefenderName string `json:"defender_name,omitempty"`
	DefenderFactionId uint `json:"defender_faction,omitempty"`
	DefenderFactionName string `json:"defender_factionname,omitempty"`
	Result string `json:"result,omitempty"`
	Stealthed Bool `json:"stealthed,omitempty"`
	RespectGain Float32 `json:"respect_gain,omitempty"`
	Chain uint64 `json:"chain,omitempty"`
	AttackModifiers AttackModifiers `json:"modifiers,omitempty"`
}

type AttackModifiers struct {
	FairFight Float32 `json:"fairFight,omitempty"`
	War Float32 `json:"war,omitempty"`
	Retaliation Float32 `json:"retaliation,omitempty"`
	GroupAttack Float32 `json:"groupAttack,omitempty"`
	Overseas Float32 `json:"overseas,omitempty"`
	ChainBonus Float32 `json:"chainBonus,omitempty"`
}