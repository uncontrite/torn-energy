package main

import (
	"encoding/json"
	"strconv"
)

type Bool struct {
	Value bool
}

func (v Bool) MarshalJson() ([]byte, error) {
	return json.Marshal(v.Value)
}

func (v *Bool) UnmarshalJSON(b []byte) error {
	var checkBool bool
	if err := json.Unmarshal(b, &checkBool); err == nil {
		*v = Bool{Value: checkBool}
		return nil
	}
	var checkString string
	if err := json.Unmarshal(b, &checkString); err == nil {
		if f, strconvErr := strconv.ParseFloat(checkString, 32); strconvErr == nil && f > 0 {
			*v = Bool{Value: true}
			return nil
		}
	}
	var checkInt int
	if err := json.Unmarshal(b, &checkInt); err == nil {
		*v = Bool{Value: 1 == checkInt}
	}
	return nil
}

func (v Bool) String() string {
	return strconv.FormatBool(v.Value)
}

type Float32 struct {
	Value string
}

func (v Float32) String() string {
	return v.Value
}

func (v Float32) MarshalJson() ([]byte, error) {
	return json.Marshal(v.Value)
}

func (v *Float32) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*v = Float32{Value : s}
		return nil
	}
	var i int64
	if err := json.Unmarshal(b, &i); err == nil {
		*v = Float32{Value: strconv.FormatInt(i, 10)}
	}
	return nil
}
