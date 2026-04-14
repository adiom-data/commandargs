package main

import (
	"testing"

	"github.com/adiom-data/commandargs/internal/plugin"
)

func TestFlagTypeToCLI(t *testing.T) {
	cases := []struct {
		ft   plugin.FlagType
		want string
	}{
		{plugin.FlagTypeString, "StringFlag"},
		{plugin.FlagTypeBool, "BoolFlag"},
		{plugin.FlagTypeInt, "IntFlag"},
		{plugin.FlagTypeInt64, "Int64Flag"},
		{plugin.FlagTypeFloat64, "Float64Flag"},
		{plugin.FlagTypeDuration, "DurationFlag"},
		{plugin.FlagTypeStringList, "StringSliceFlag"},
		{plugin.FlagTypeIntList, "IntSliceFlag"},
		{plugin.FlagTypeInt64List, "Int64SliceFlag"},
	}
	for _, c := range cases {
		got := flagTypeToCLI(c.ft)
		if got != c.want {
			t.Errorf("flagTypeToCLI(%q) = %q, want %q", c.ft, got, c.want)
		}
	}
}
