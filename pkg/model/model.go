package model

type FlagType int

const (
	FlagTypeString FlagType = iota
	FlagTypeBool
	FlagTypeInt
	FlagTypeInt64
	FlagTypeFloat64
	FlagTypeStringSlice
	FlagTypeIntSlice
	FlagTypeInt64Slice
	FlagTypeFloat64Slice
)

func (f FlagType) String() string {
	switch f {
	case FlagTypeString:
		return "string"
	case FlagTypeBool:
		return "bool"
	case FlagTypeInt:
		return "int"
	case FlagTypeInt64:
		return "int64"
	case FlagTypeFloat64:
		return "float64"
	case FlagTypeStringSlice:
		return "[]string"
	case FlagTypeIntSlice:
		return "[]int"
	case FlagTypeInt64Slice:
		return "[]int64"
	case FlagTypeFloat64Slice:
		return "[]float64"
	default:
		return "string"
	}
}

type File struct {
	Commands []*Command
}

type Command struct {
	Name        string
	Description string
	Aliases     []string
	Usage       string
	Hidden      bool
	Category    string
	Flags       []Flag
}

type Flag struct {
	Name          string
	GoName        string
	Usage         string
	Default       string
	Required      bool
	Hidden        bool
	Category      string
	Aliases       []string
	AllowedValues []string
	Type          FlagType
}
