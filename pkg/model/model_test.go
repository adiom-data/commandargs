package model

import "testing"

func TestFlagTypeString(t *testing.T) {
	cases := []struct {
		ft   FlagType
		want string
	}{
		{FlagTypeString, "string"},
		{FlagTypeBool, "bool"},
		{FlagTypeInt, "int"},
		{FlagTypeInt64, "int64"},
		{FlagTypeFloat64, "float64"},
		{FlagTypeStringSlice, "[]string"},
		{FlagTypeIntSlice, "[]int"},
		{FlagTypeInt64Slice, "[]int64"},
		{FlagTypeFloat64Slice, "[]float64"},
	}
	for _, c := range cases {
		got := c.ft.String()
		if got != c.want {
			t.Errorf("FlagType(%d).String() = %q, want %q", c.ft, got, c.want)
		}
	}
}
