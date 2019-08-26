package main

import "math/big"

type BattleStats struct {
	Strength string `json:"strength,omitempty"`
	Speed string `json:"speed,omitempty"`
	Dexterity string `json:"dexterity,omitempty"`
	Defense string `json:"defense,omitempty"`
}

const prec = 100

func Sub(l string, r string) string {
	zero, _ := new(big.Float).SetPrec(prec).SetString("0")
	ll, _ := new(big.Float).SetPrec(prec).SetString(l)
	rr, _ := new(big.Float).SetPrec(prec).SetString(r)
	result := new(big.Float).Sub(ll, rr)
	if result.Cmp(zero) == 0 {
		return ""
	}
	return result.Text('f', 4)
}

// Should only be run on diff
// TODO: Type alias to ensure this via extended interface
func (bs BattleStats) IsTrain() bool {
	zero, _ := new(big.Float).SetPrec(prec).SetString("0")
	str := ToFloat(bs.Strength)
	def := ToFloat(bs.Defense)
	dex := ToFloat(bs.Dexterity)
	spd := ToFloat(bs.Speed)
	return str.Cmp(zero) > 0 || def.Cmp(zero) > 0 || dex.Cmp(zero) > 0 || spd.Cmp(zero) > 0
}

func ToFloat(value string) *big.Float {
	if value == "" {
		value = "0"
	}
	ret, _ := new(big.Float).SetPrec(prec).SetString(value)
	return ret
}

func (bs BattleStats) Diff(bs2 BattleStats) BattleStats {
	var diff BattleStats
	diff.Defense = Sub(bs2.Defense, bs.Defense)
	diff.Dexterity = Sub(bs2.Dexterity, bs.Dexterity)
	diff.Speed = Sub(bs2.Speed, bs.Speed)
	diff.Strength = Sub(bs2.Strength, bs.Strength)	
	return diff
}