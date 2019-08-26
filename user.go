package main

import (
	"encoding/json"
	"errors"
)

type RawUser struct {
	// BattleStats
	Strength string `json:"strength,omitempty"`
	Speed string `json:"speed,omitempty"`
	Dexterity string `json:"dexterity,omitempty"`
	Defense string `json:"defense,omitempty"`

	// Bars
	Energy Energy `json:"energy,omitempty"`
	Happy Happy `json:"happy,omitempty"`

	// Fields
	PlayerId uint `json:"player_id,omitempty"`

	// Well-structured crap
	JobPoints RawJobPoints `json:"jobpoints,omitempty"`
	PersonalStats PersonalStats `json:"personalstats,omitempty"`
	Refills Refills `json:"refills,omitempty"`
}

type User struct {
	UserId uint `json:"userId,omitempty"`
	BattleStats BattleStats `json:"battlestats,omitempty"`
	Bars Bars `json:"bars,omitempty"`
	Jobs []Job `json:"jobs,omitempty"`
	PersonalStats PersonalStats `json:"personalstats,omitempty"`
	Refills Refills `json:"refills,omitempty"`
}

func (u User) Equals(other interface{}) bool {
	if other == nil {
		return false
	}
	return u.UserId == other.(User).UserId &&
		u.BattleStats == other.(User).BattleStats &&
		u.Bars == other.(User).Bars &&
		Eq(u.Jobs, other.(User).Jobs) &&
		u.PersonalStats == other.(User).PersonalStats &&
		u.Refills == other.(User).Refills
}

func (raw RawUser) Bars() Bars {
	return Bars{raw.Energy, raw.Happy}
}

func (raw RawUser) BattleStats() BattleStats {
	return BattleStats{raw.Strength, raw.Speed, raw.Dexterity, raw.Defense}
}

func (raw RawUser) User() (*User, error) {
	jobs, err := raw.JobPoints.ToJobs()
	if err != nil {
		return nil, err
	}
	return &User{raw.PlayerId, raw.BattleStats(), raw.Bars(), jobs, raw.PersonalStats, raw.Refills}, nil
}

func (u User) MarshalJson() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalJSON(b []byte) error {
	// Determine type based on fields
	responseType, err := GetUserResponseType(b)
	if err != nil {
		return err
	}

	switch *responseType {
	case "Torn":
		var raw RawUser
		if err = json.Unmarshal(b, &raw); err == nil {
			innerUser, innerErr := raw.User()
			err = innerErr
			if innerUser != nil {
				*u = *innerUser
			}
		}
		break
	case "Kafka":
		type User2 User
		var user2 User2
		err = json.Unmarshal(b, &user2)
		*u = User(user2)
		break
	case "Error":
		err = errors.New("Unable to convert Torn Error into User: " + string(b))
		break
	}
	return err
}

func GetUserResponseType(b []byte) (*string, error) {
	// Determine type based on fields
	var body map[string]*json.RawMessage
	err := json.Unmarshal(b, &body)
	if err != nil {
		return nil, err
	}
	var resp string
	if _, fromTornApi := body["strength"]; fromTornApi {
		resp = "Torn"
	} else if _, fromKafka := body["bars"]; fromKafka {
		resp = "Kafka"
	} else if _, errorLike := body["error"]; errorLike {
		resp = "Error"
	}
	return &resp, nil
}