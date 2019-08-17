package main

type Energy struct {
	Current uint `json:"current,omitempty"`
	Maximum uint `json:"maximum,omitempty"`
}

type Happy struct {
	Current uint `json:"current,omitempty"`
	Maximum uint `json:"maximum,omitempty"`
}

type Bars struct {
	Energy Energy `json:"energy,omitempty"`
	Happy Happy `json:"happy,omitempty"`
}
