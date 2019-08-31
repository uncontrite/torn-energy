package model

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
)

type RawUser struct {
	// BattleStats
	Strength string `json:"strength,omitempty"`
	Speed string `json:"speed,omitempty"`
	Dexterity string `json:"dexterity,omitempty"`
	Defense string `json:"defense,omitempty"`

	// Bars
	Energy Energy `json:"energy,omitempty"`
	Happy  Happy  `json:"happy,omitempty"`

	// Fields
	Name string `json:"name,omitempty"`
	PlayerId uint `json:"player_id,omitempty"`

	// Well-structured crap
	JobPoints     RawJobPoints  `json:"jobpoints,omitempty"`
	PersonalStats PersonalStats `json:"personalstats,omitempty"`
	Refills       Refills       `json:"refills,omitempty"`
	Inventory     []Item		`json:"inventory,omitempty"`
}

type User struct {
	UserId        uint          `json:"userId,omitempty"`
	Name	      string        `json:"name,omitempty"`
	BattleStats   BattleStats   `json:"battlestats,omitempty"`
	Bars          Bars          `json:"bars,omitempty"`
	Jobs          []Job         `json:"jobs,omitempty"`
	PersonalStats PersonalStats `json:"personalstats,omitempty"`
	Refills       Refills       `json:"refills,omitempty"`
	Items		  []Item		`json:"inventory,omitempty"`
}

const (
	FHC = 367
	EDVD = 366
)

type Item struct {
	Id int `json:"ID,omitempty"` // FHC, EDVD consts above
	Quantity int `json:"quantity,omitempty"`
}

func (i Item) Diff(other Item) Item {
	return Item{
		Id: i.Id,
		Quantity: other.Quantity - i.Quantity,
	}
}

func CalculateItemDiffs(prev []Item, curr []Item) []Item {
	items := make(map[int]int)
	for _, item := range curr {
		items[item.Id] = item.Quantity
	}
	for _, item := range prev {
		items[item.Id] -= item.Quantity
	}
	var result []Item
	for id, quantity := range items {
		result = append(result, Item{id, quantity})
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Id < result[j].Id
	})
	return result
}

func (u User) Equals(other interface{}) bool {
	if other == nil {
		return false
	}
	return u.UserId == other.(User).UserId &&
		u.BattleStats == other.(User).BattleStats &&
		u.Bars.Equals(other.(User).Bars) &&
		Eq(u.Jobs, other.(User).Jobs) &&
		u.PersonalStats == other.(User).PersonalStats &&
		u.Refills == other.(User).Refills &&
		reflect.DeepEqual(u.Items, other.(User).Items)
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
	return &User{raw.PlayerId, raw.Name, raw.BattleStats(), raw.Bars(), jobs, raw.PersonalStats, raw.Refills, raw.Inventory}, nil
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
				var newItems []Item
				for _, item := range u.Items {
					if item.Id == FHC || item.Id == EDVD {
						newItems = append(newItems, item)
					}
				}
				u.Items = newItems
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