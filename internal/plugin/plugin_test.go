package plugin

import (
	"testing"
)

func TestValidateDefault(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		ft      FlagType
		value   string
		wantErr bool
	}{
		{"valid bool true", "verbose", FlagTypeBool, "true", false},
		{"valid bool false", "verbose", FlagTypeBool, "false", false},
		{"invalid bool", "verbose", FlagTypeBool, "yes", true},
		{"valid int", "port", FlagTypeInt, "8080", false},
		{"invalid int", "port", FlagTypeInt, "abc", true},
		{"int overflow", "port", FlagTypeInt, "99999999999999", true},
		{"valid int64", "size", FlagTypeInt64, "9999999999", false},
		{"invalid int64", "size", FlagTypeInt64, "abc", true},
		{"valid float", "rate", FlagTypeFloat64, "3.14", false},
		{"invalid float", "rate", FlagTypeFloat64, "abc", true},
		{"valid duration", "timeout", FlagTypeDuration, "30s", false},
		{"invalid duration", "timeout", FlagTypeDuration, "abc", true},
		{"valid timestamp", "start", FlagTypeTimestamp, "2025-01-15T10:30:00Z", false},
		{"invalid timestamp", "start", FlagTypeTimestamp, "not-a-time", true},
		{"string always valid", "name", FlagTypeString, "anything", false},
		{"string list always valid", "tags", FlagTypeStringList, "a, b, c", false},
		{"valid int list", "codes", FlagTypeIntList, "1, 2, 3", false},
		{"invalid int list element", "codes", FlagTypeIntList, "1, abc, 3", true},
		{"valid int64 list", "ids", FlagTypeInt64List, "100, 200", false},
		{"invalid int64 list element", "ids", FlagTypeInt64List, "100, xyz", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDefault(tt.flag, tt.ft, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDefault(%q, %q, %q) error = %v, wantErr %v", tt.flag, tt.ft, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDefaultAllowedValues(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		ft      FlagType
		value   string
		allowed []string
		wantErr bool
	}{
		{"no allowed values", "level", FlagTypeString, "any", nil, false},
		{"valid scalar", "level", FlagTypeString, "info", []string{"debug", "info", "warn"}, false},
		{"invalid scalar", "level", FlagTypeString, "trace", []string{"debug", "info", "warn"}, true},
		{"valid list all match", "tags", FlagTypeStringList, "a, b", []string{"a", "b", "c"}, false},
		{"invalid list element", "tags", FlagTypeStringList, "a, x", []string{"a", "b", "c"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDefaultAllowedValues(tt.flag, tt.ft, tt.value, tt.allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDefaultAllowedValues(%q, %q, %q, %v) error = %v, wantErr %v", tt.flag, tt.ft, tt.value, tt.allowed, err, tt.wantErr)
			}
		})
	}
}
