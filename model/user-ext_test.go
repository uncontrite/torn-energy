package model

import (
	"encoding/json"
	"reflect"
	"testing"
)

const BEFORE = `{
  "bars": {
    "energy": {
      "current": 95,
      "maximum": 150
    },
    "happy": {
      "current": 4905,
      "maximum": 5025
    }
  },
  "battlestats": {
    "defense": "350512165.7422",
    "dexterity": "351973455.6630",
    "speed": "357500922.9665",
    "strength": "1000271903.1330"
  },
  "jobs": [
    {
      "Name": "Adult Novelties",
      "Points": 2
    },
    {
      "Name": "Pub",
      "Points": 10
    },
    {
      "Name": "casino",
      "Points": 5
    },
    {
      "Name": "education",
      "Points": 4
    },
    {
      "Name": "law",
      "Points": 16
    },
    {
      "Name": "medical",
      "Points": 1
    }
  ],
  "personalstats": {
    "alcoholused": 382,
    "attacksassisted": 25,
    "attacksdraw": 96,
    "attackslost": 387,
    "attackswon": 8376,
    "booksread": 4,
    "boostersused": 2878,
    "candyused": 36,
    "cantaken": 45,
    "consumablesused": 746,
    "dumpsearches": 379,
    "energydrinkused": 328,
    "exttaken": 65,
    "logins": 1820,
    "lsdtaken": 1,
    "nerverefills": 184,
    "overdosed": 48,
    "refills": 795,
    "useractivity": 5573812,
    "xantaken": 2085,
    "yourunaway": 38
  },
  "refills": {
    "energy_refill_used": false
  },
  "userId": 2040809
}
`
const AFTER = `{
  "bars": {
    "energy": {
      "current": 85,
      "maximum": 150
    },
    "happy": {
      "current": 4964,
      "maximum": 5025
    }
  },
  "battlestats": {
    "defense": "350512165.7422",
    "dexterity": "351973455.6630",
    "speed": "391103118.8087",
    "strength": "1001314404.3932"
  },
  "jobs": [
    {
      "Name": "Adult Novelties",
      "Points": 2
    },
    {
      "Name": "Pub",
      "Points": 10
    },
    {
      "Name": "casino",
      "Points": 5
    },
    {
      "Name": "education",
      "Points": 4
    },
    {
      "Name": "law",
      "Points": 40
    },
    {
      "Name": "medical",
      "Points": 1
    }
  ],
  "personalstats": {
    "alcoholused": 382,
    "attacksassisted": 25,
    "attacksdraw": 96,
    "attackslost": 387,
    "attackswon": 8445,
    "booksread": 4,
    "boostersused": 2909,
    "candyused": 37,
    "cantaken": 45,
    "consumablesused": 748,
    "dumpsearches": 386,
    "energydrinkused": 329,
    "exttaken": 65,
    "logins": 1820,
    "lsdtaken": 1,
    "nerverefills": 191,
    "overdosed": 49,
    "refills": 802,
    "useractivity": 5615488,
    "xantaken": 2104,
    "yourunaway": 38
  },
  "refills": {
    "energy_refill_used": true
  },
  "userId": 2040809
}`

const DIFF = `{
  "bars": {
    "energy": {
      "previous": 95,
      "current": 85,
      "diff": -10
    },
    "happy": {
      "previous": 4905,
      "current": 4964,
      "diff": 59
    }
  },
  "battlestats": {
    "speed": "33602195.8422",
    "strength": "1042501.2602"
  },
  "jobs": [
    {
      "Name": "law",
      "Points": 24
    }
  ],
  "personalstats": {
    "alcoholused": 0,
    "attacksassisted": 0,
    "attacksdraw": 0,
    "attackslost": 0,
    "attackswon": 69,
    "booksread": 0,
    "boostersused": 31,
    "candyused": 1,
    "cantaken": 0,
    "consumablesused": 2,
    "dumpsearches": 7,
    "energydrinkused": 1,
    "exttaken": 0,
    "logins": 0,
    "lsdtaken": 0,
    "nerverefills": 7,
    "overdosed": 1,
    "refills": 7,
    "useractivity": 41676,
    "xantaken": 19,
    "yourunaway": 0
  },
  "refills": {
    "energy_refill_used": true
  },
  "userId": 2040809
}`

func TestSub(t *testing.T) {
	type args struct {
		l string
		r string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Basic",args{"2", "1"}, "1.0000"},
		{"Negative Outcome",args{"2", "3"}, "-1.0000"},
		{"Float",args{"391103118.8087", "383085829.4637"}, "8017289.3450"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sub(tt.args.l, tt.args.r); got != tt.want {
				t.Errorf("Sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_Diff(t *testing.T) {
	var before User
	var after User
	var want User
	if err := json.Unmarshal([]byte(BEFORE), &before); err != nil {
		t.Errorf("Unable to unmarshal BEFORE: %s", err)
	}
	if err := json.Unmarshal([]byte(AFTER), &after); err != nil {
		t.Errorf("Unable to unmarshal AFTER: %s", err)
	}
	if err := json.Unmarshal([]byte(DIFF), &want); err != nil {
		t.Errorf("Unable to unmarshal DIFF: %s", err)
	}
	if got := before.Diff(after); !reflect.DeepEqual(got, want) {
		t.Errorf("Diff() = %+v, want %+v", got, want)
	}
}

func TestCalculateBoosterSplit(t *testing.T) {
	type args struct {
		prevHappy     int
		currHappy     int
		ecstasyTaken  int
		boostersTaken int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
	}{
		{"FHC1", args{4779, 5209, 0, 1}, 1, 0},
		{"FHC2", args{4983, 7874, 0, 7}, 7, 0},
		{"FHC3", args{5364, 6134, 0, 2}, 2, 0},
		{"FHC4", args{6134, 7486, 0, 3}, 3, 0},
		{"FHC5", args{7486, 7914, 0, 1}, 1, 0},
		{"EDVDs", args{9500, 33776, 1, 3}, 0, 3},
		{"Mix", args{5000, 7920, 0, 2}, 1, 1},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CalculateBoosterSplit(tt.args.prevHappy, tt.args.currHappy, tt.args.ecstasyTaken, tt.args.boostersTaken)
			if got != tt.want {
				t.Errorf("CalculateBoosterSplit() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CalculateBoosterSplit() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}