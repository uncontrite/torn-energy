package model

type Energy struct {
	Previous int `json:"previous,omitempty"`
	Current int `json:"current,omitempty"`
	Diff int `json:"diff,omitempty"`
	Maximum int `json:"maximum,omitempty"`
	TickTime int `json:"ticktime,omitempty"`
}

type Happy struct {
	Previous int `json:"previous,omitempty"`
	Current int `json:"current,omitempty"`
	Diff int `json:"diff,omitempty"`
	Maximum int `json:"maximum,omitempty"`
}

type Bars struct {
	Energy Energy `json:"energy,omitempty"`
	Happy Happy `json:"happy,omitempty"`
}

func (b Bars) Equals(b2 Bars) bool {
	return b.Energy.Current == b2.Energy.Current &&
		b.Happy.Current == b2.Happy.Current
}

func (b Bars) Diff(b2 Bars) Bars {
	var diff Bars
	diff.Energy.Previous = b.Energy.Current
	diff.Energy.Current = b2.Energy.Current
	diff.Energy.Diff = b2.Energy.Current - b.Energy.Current
	diff.Energy.TickTime = b2.Energy.TickTime - b.Energy.TickTime

	diff.Happy.Previous = b.Happy.Current
	diff.Happy.Current = b2.Happy.Current
	diff.Happy.Diff = b2.Happy.Current - b.Happy.Current
	return diff
}