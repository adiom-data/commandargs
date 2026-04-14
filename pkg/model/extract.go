package model

import (
	commandargsv1 "github.com/adiom-data/commandargs/gen/commandargs/v1"
	"github.com/adiom-data/commandargs/internal/naming"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func ExtractFile(file *protogen.File) *File {
	var commands []*Command
	for _, msg := range file.Messages {
		collectCommands(msg, &commands)
	}
	if len(commands) == 0 {
		return nil
	}
	return &File{Commands: commands}
}

func collectCommands(msg *protogen.Message, commands *[]*Command) {
	if getCommandOptions(msg) != nil {
		cmd := extractCommand(msg)
		*commands = append(*commands, cmd)
	}
	for _, nested := range msg.Messages {
		collectCommands(nested, commands)
	}
}

func extractCommand(msg *protogen.Message) *Command {
	cmdOpts := getCommandOptions(msg)

	cmd := &Command{}
	if cmdOpts != nil {
		cmd.Name = cmdOpts.GetName()
		cmd.Description = cmdOpts.GetDescription()
		cmd.Aliases = cmdOpts.GetAliases()
		cmd.Usage = cmdOpts.GetUsage()
		cmd.Hidden = cmdOpts.GetHidden()
		cmd.Category = cmdOpts.GetCategory()
	}
	if cmd.Name == "" {
		cmd.Name = naming.PascalToKebab(string(msg.Desc.Name()))
	}

	for _, field := range msg.Fields {
		if field.Message != nil {
			inlineFlags(cmd, field.Message)
		} else {
			cmd.Flags = append(cmd.Flags, flagFromField(field))
		}
	}

	return cmd
}

func inlineFlags(cmd *Command, msg *protogen.Message) {
	for _, field := range msg.Fields {
		if field.Message != nil {
			inlineFlags(cmd, field.Message)
		} else {
			cmd.Flags = append(cmd.Flags, flagFromField(field))
		}
	}
}

func flagFromField(field *protogen.Field) Flag {
	f := Flag{
		GoName: field.GoName,
		Type:   protoKindToFlagType(field.Desc),
	}

	opts := getFlagOptions(field)
	if opts != nil {
		f.Name = opts.GetName()
		f.Usage = opts.GetUsage()
		f.Default = opts.GetDefault()
		f.Required = opts.GetRequired()
		f.Hidden = opts.GetHidden()
		f.Category = opts.GetCategory()
		f.Aliases = opts.GetAliases()
		f.AllowedValues = opts.GetAllowedValues()
	}

	if f.Name == "" {
		f.Name = naming.SnakeToKebab(string(field.Desc.Name()))
	}

	return f
}

func protoKindToFlagType(fd protoreflect.FieldDescriptor) FlagType {
	if fd.IsList() {
		switch fd.Kind() {
		case protoreflect.StringKind:
			return FlagTypeStringSlice
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind:
			return FlagTypeIntSlice
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind:
			return FlagTypeInt64Slice
		case protoreflect.FloatKind, protoreflect.DoubleKind:
			return FlagTypeFloat64Slice
		default:
			return FlagTypeStringSlice
		}
	}

	switch fd.Kind() {
	case protoreflect.BoolKind:
		return FlagTypeBool
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind:
		return FlagTypeInt
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind:
		return FlagTypeInt64
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return FlagTypeFloat64
	default:
		return FlagTypeString
	}
}

func getCommandOptions(msg *protogen.Message) *commandargsv1.CommandOptions {
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

func getFlagOptions(field *protogen.Field) *commandargsv1.FlagOptions {
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


