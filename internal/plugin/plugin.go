package plugin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	commandargsv1 "github.com/adiom-data/commandargs/gen/commandargs/v1"
	"github.com/adiom-data/commandargs/internal/naming"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type FlagType string

const (
	FlagTypeString    FlagType = "string"
	FlagTypeBool      FlagType = "bool"
	FlagTypeInt       FlagType = "int"
	FlagTypeInt64     FlagType = "int64"
	FlagTypeFloat64   FlagType = "float64"
	FlagTypeDuration  FlagType = "duration"
	FlagTypeTimestamp FlagType = "timestamp"

	FlagTypeStringList FlagType = "string_list"
	FlagTypeIntList    FlagType = "int_list"
	FlagTypeInt64List  FlagType = "int64_list"
)

type WellKnownTypeInfo struct {
	FlagType FlagType
}

var WellKnownTypes = map[string]WellKnownTypeInfo{
	"google.protobuf.Duration":    {FlagType: FlagTypeDuration},
	"google.protobuf.Timestamp":   {FlagType: FlagTypeTimestamp},
	"google.protobuf.StringValue": {FlagType: FlagTypeString},
	"google.protobuf.BoolValue":   {FlagType: FlagTypeBool},
	"google.protobuf.Int32Value":  {FlagType: FlagTypeInt},
	"google.protobuf.Int64Value":  {FlagType: FlagTypeInt64},

	"google.protobuf.FloatValue":  {FlagType: FlagTypeFloat64},
	"google.protobuf.DoubleValue": {FlagType: FlagTypeFloat64},
}

type FlagInfo struct {
	Name          string
	Usage         string
	Default       string
	Aliases       []string
	Required      bool
	Hidden        bool
	Category      string
	AllowedValues []string
	Type          FlagType
	IsList        bool
}

type CommandInfo struct {
	Name        string
	Description string
	Aliases     []string
	Usage       string
	Hidden      bool
	Category    string
	Flags       []FlagInfo
	Msg         *protogen.Message
}

func GetWellKnownType(field *protogen.Field) (WellKnownTypeInfo, bool) {
	if field.Message == nil {
		return WellKnownTypeInfo{}, false
	}
	wkt, ok := WellKnownTypes[string(field.Message.Desc.FullName())]
	return wkt, ok
}

func IsWellKnownScalar(field *protogen.Field) bool {
	_, ok := GetWellKnownType(field)
	return ok
}

func IsUnsupportedWellKnown(field *protogen.Field) bool {
	if field.Message == nil {
		return false
	}
	pkg := field.Message.Desc.ParentFile().Package()
	if pkg == "google.protobuf" {
		_, supported := WellKnownTypes[string(field.Message.Desc.FullName())]
		return !supported
	}
	return false
}

func GetCommandOptions(msg *protogen.Message) *commandargsv1.CommandOptions {
	opts := msg.Desc.Options()
	if opts == nil {
		return nil
	}
	if !proto.HasExtension(opts, commandargsv1.E_Command) {
		return nil
	}
	cmdOpts, ok := proto.GetExtension(opts, commandargsv1.E_Command).(*commandargsv1.CommandOptions)
	if !ok {
		return nil
	}
	return cmdOpts
}

func GetFlagOptions(field *protogen.Field) *commandargsv1.FlagOptions {
	opts := field.Desc.Options()
	if opts == nil {
		return nil
	}
	if !proto.HasExtension(opts, commandargsv1.E_Flag) {
		return nil
	}
	flagOpts, ok := proto.GetExtension(opts, commandargsv1.E_Flag).(*commandargsv1.FlagOptions)
	if !ok {
		return nil
	}
	return flagOpts
}

func FlagNameForField(field *protogen.Field) string {
	opts := GetFlagOptions(field)
	if opts != nil && opts.GetName() != "" {
		return opts.GetName()
	}
	return naming.SnakeToKebab(string(field.Desc.Name()))
}

func CollectCommandMessages(file *protogen.File) []*protogen.Message {
	var out []*protogen.Message
	for _, msg := range file.Messages {
		collectCommandMessagesRecursive(msg, &out)
	}
	return out
}

func collectCommandMessagesRecursive(msg *protogen.Message, out *[]*protogen.Message) {
	if GetCommandOptions(msg) != nil {
		*out = append(*out, msg)
	}
	for _, nested := range msg.Messages {
		collectCommandMessagesRecursive(nested, out)
	}
}

func ValidateMessage(msg *protogen.Message) error {
	names := make(map[string]string)
	return validateFieldsRecursive(msg, names)
}

func validateFieldsRecursive(msg *protogen.Message, names map[string]string) error {
	for _, field := range msg.Fields {
		if field.Message != nil && !IsWellKnownScalar(field) {
			if field.Desc.IsMap() {
				return fmt.Errorf("unsupported field type: map field %q in message %q",
					field.Desc.Name(), msg.Desc.FullName())
			}
			if IsUnsupportedWellKnown(field) {
				return fmt.Errorf("unsupported well-known type %q for field %q in message %q",
					field.Message.Desc.FullName(), field.Desc.Name(), msg.Desc.FullName())
			}
			if err := validateFieldsRecursive(field.Message, names); err != nil {
				return err
			}
			continue
		}

		if field.Message == nil && !isSupportedKind(field.Desc.Kind()) {
			return fmt.Errorf("unsupported field type %q for field %q in message %q",
				field.Desc.Kind(), field.Desc.Name(), msg.Desc.FullName())
		}

		if field.Message == nil && field.Desc.IsList() && !isSupportedRepeatedKind(field.Desc.Kind()) {
			return fmt.Errorf("unsupported repeated field type %q for field %q in message %q",
				field.Desc.Kind(), field.Desc.Name(), msg.Desc.FullName())
		}

		flagName := FlagNameForField(field)
		source := string(msg.Desc.FullName()) + "." + string(field.Desc.Name())

		if existing, ok := names[flagName]; ok {
			return fmt.Errorf("duplicate flag name %q: defined by %s and %s",
				flagName, existing, source)
		}
		names[flagName] = source

		opts := GetFlagOptions(field)
		if opts != nil {
			for _, alias := range opts.GetAliases() {
				if existing, ok := names[alias]; ok {
					return fmt.Errorf("duplicate flag name %q (alias): defined by %s and %s",
						alias, existing, source)
				}
				names[alias] = source
			}
		}
	}
	return nil
}

func isSupportedKind(k protoreflect.Kind) bool {
	switch k {
	case protoreflect.BoolKind,
		protoreflect.Int32Kind, protoreflect.Sint32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind,
		protoreflect.FloatKind, protoreflect.DoubleKind,
		protoreflect.StringKind:
		return true
	}
	return false
}

func isSupportedRepeatedKind(k protoreflect.Kind) bool {
	switch k {
	case protoreflect.StringKind,
		protoreflect.Int32Kind, protoreflect.Sint32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind:
		return true
	}
	return false
}

func protoKindToFlagType(kind protoreflect.Kind, isList bool) FlagType {
	if isList {
		switch kind {
		case protoreflect.StringKind:
			return FlagTypeStringList
		case protoreflect.Int32Kind, protoreflect.Sint32Kind:
			return FlagTypeIntList
		case protoreflect.Int64Kind, protoreflect.Sint64Kind:
			return FlagTypeInt64List
		}
	}
	switch kind {
	case protoreflect.BoolKind:
		return FlagTypeBool
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		return FlagTypeInt
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		return FlagTypeInt64
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return FlagTypeFloat64
	default:
		return FlagTypeString
	}
}

func ExtractCommand(msg *protogen.Message) (CommandInfo, error) {
	cmdOpts := GetCommandOptions(msg)
	ci := CommandInfo{Msg: msg}
	if cmdOpts != nil {
		ci.Name = cmdOpts.GetName()
		ci.Description = cmdOpts.GetDescription()
		ci.Aliases = cmdOpts.GetAliases()
		ci.Usage = cmdOpts.GetUsage()
		ci.Hidden = cmdOpts.GetHidden()
		ci.Category = cmdOpts.GetCategory()
	}
	if ci.Name == "" {
		ci.Name = naming.PascalToKebab(string(msg.Desc.Name()))
	}
	if err := collectFlagsRecursive(msg, &ci.Flags); err != nil {
		return CommandInfo{}, err
	}
	return ci, nil
}

func collectFlagsRecursive(msg *protogen.Message, flags *[]FlagInfo) error {
	for _, field := range msg.Fields {
		if field.Message != nil && !IsWellKnownScalar(field) {
			if err := collectFlagsRecursive(field.Message, flags); err != nil {
				return err
			}
			continue
		}

		fi := FlagInfo{Name: FlagNameForField(field), IsList: field.Desc.IsList()}
		if wkt, ok := GetWellKnownType(field); ok {
			if field.Desc.IsList() {
				fi.Type = FlagTypeStringList
			} else {
				fi.Type = wkt.FlagType
			}
		} else {
			fi.Type = protoKindToFlagType(field.Desc.Kind(), field.Desc.IsList())
		}

		opts := GetFlagOptions(field)
		if opts != nil {
			if opts.GetName() != "" {
				fi.Name = opts.GetName()
			}
			fi.Usage = opts.GetUsage()
			fi.Default = opts.GetDefault()
			fi.Aliases = opts.GetAliases()
			fi.Required = opts.GetRequired()
			fi.Hidden = opts.GetHidden()
			fi.Category = opts.GetCategory()
			fi.AllowedValues = opts.GetAllowedValues()
		}

		if fi.Default != "" {
			if err := validateDefault(fi.Name, fi.Type, fi.Default); err != nil {
				return err
			}
			if err := validateDefaultAllowedValues(fi.Name, fi.Type, fi.Default, fi.AllowedValues); err != nil {
				return err
			}
		}

		*flags = append(*flags, fi)
	}
	return nil
}

func validateDefault(flagName string, ft FlagType, value string) error {
	switch ft {
	case FlagTypeStringList:
		// All string elements are valid.
	case FlagTypeIntList:
		for _, elem := range splitDefault(value) {
			if _, err := strconv.ParseInt(elem, 10, 32); err != nil {
				return fmt.Errorf("invalid default element %q for int list flag %q: %w", elem, flagName, err)
			}
		}
	case FlagTypeInt64List:
		for _, elem := range splitDefault(value) {
			if _, err := strconv.ParseInt(elem, 10, 64); err != nil {
				return fmt.Errorf("invalid default element %q for int64 list flag %q: %w", elem, flagName, err)
			}
		}
	case FlagTypeBool:
		if value != "true" && value != "false" {
			return fmt.Errorf("invalid default %q for bool flag %q: must be \"true\" or \"false\"", value, flagName)
		}
	case FlagTypeInt:
		if _, err := strconv.ParseInt(value, 10, 32); err != nil {
			return fmt.Errorf("invalid default %q for int flag %q: %w", value, flagName, err)
		}
	case FlagTypeInt64:
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("invalid default %q for int64 flag %q: %w", value, flagName, err)
		}
	case FlagTypeFloat64:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("invalid default %q for float64 flag %q: %w", value, flagName, err)
		}
	case FlagTypeDuration:
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid default %q for duration flag %q: %w", value, flagName, err)
		}
	case FlagTypeTimestamp:
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			return fmt.Errorf("invalid default %q for timestamp flag %q: %w", value, flagName, err)
		}
	}
	return nil
}

func validateDefaultAllowedValues(flagName string, ft FlagType, value string, allowedValues []string) error {
	if len(allowedValues) == 0 {
		return nil
	}
	allowed := make(map[string]bool, len(allowedValues))
	for _, v := range allowedValues {
		allowed[v] = true
	}
	switch ft {
	case FlagTypeStringList, FlagTypeIntList, FlagTypeInt64List:
		for _, elem := range splitDefault(value) {
			if !allowed[elem] {
				return fmt.Errorf("default element %q for flag %q is not in allowed_values %v", elem, flagName, allowedValues)
			}
		}
	default:
		if !allowed[value] {
			return fmt.Errorf("default %q for flag %q is not in allowed_values %v", value, flagName, allowedValues)
		}
	}
	return nil
}

func SplitDefault(value string) []string {
	return splitDefault(value)
}

func splitDefault(value string) []string {
	parts := strings.Split(value, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
