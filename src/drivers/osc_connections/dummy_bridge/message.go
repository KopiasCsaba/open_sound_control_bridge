package dummy_bridge

import (
	"fmt"
	"strconv"

	"github.com/scgolang/osc"
	"net.kopias.oscbridge/app/drivers/osc_message"
	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

const (
	ArgTypeInt    = "int"
	ArgTypeFloat  = "float"
	ArgTypeBool   = "bool"
	ArgTypeString = "string"
)

// MessageFromOSCMessage converts the internal osc Message to an IOSCMessage.
func MessageFromOSCMessage(oscMsg osc.Message) (usecaseifs.IOSCMessage, error) {
	arguments := []usecaseifs.IOSCMessageArgument{}

	for i, oscArg := range oscMsg.Arguments {
		arg, err := MessageArgumentFromOSCArgument(oscArg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert arg %d: %w", i, err)
		}
		arguments = append(arguments, arg)
	}

	return osc_message.NewMessage(oscMsg.Address, arguments), nil
}

// OSCMessageFromMessage converts an IOSCMessage into the internal osc Message.
func OSCMessageFromMessage(msg usecaseifs.IOSCMessage) (*osc.Message, error) {
	oscMessage := osc.Message{}
	oscMessage.Address = msg.GetAddress()

	arguments := []osc.Argument{}

	for _, msgArg := range msg.GetArguments() {
		oscArg, err := OSCArgumentFromMessageArgument(msgArg)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, oscArg)
	}

	oscMessage.Arguments = arguments
	return &oscMessage, nil
}

// MessageArgumentFromOSCArgument converts the internal osc Message to an IOSCMessageArgument.
func MessageArgumentFromOSCArgument(arg osc.Argument) (usecaseifs.IOSCMessageArgument, error) {
	var msgType string
	var msgValue string
	var err error
	switch arg.Typetag() {
	case osc.TypetagInt:
		msgType = ArgTypeInt
		intValue, err := arg.ReadInt32()
		if err != nil {
			return nil, fmt.Errorf("failed to read Int32 argument: %w", err)
		}
		msgValue = string(intValue)

	case osc.TypetagFloat:
		msgType = ArgTypeFloat
		floatValue, err := arg.ReadFloat32()
		if err != nil {
			return nil, fmt.Errorf("failed to read float32 argument: %w", err)
		}
		msgValue = fmt.Sprintf("%f", floatValue)

	case osc.TypetagTrue:
		msgType = ArgTypeBool
		msgValue = "true"

	case osc.TypetagFalse:
		msgType = ArgTypeBool
		msgValue = "false"

	case osc.TypetagString:
		msgType = ArgTypeString
		msgValue, err = arg.ReadString()
		if err != nil {
			return nil, fmt.Errorf("failed to read string argument: %w", err)
		}
	// case osc.TypetagBlob:
	default:
		return nil, fmt.Errorf("unsupported type: %q", string(arg.Typetag()))
	}
	return osc_message.NewMessageArgument(msgType, msgValue), nil
}

// OSCArgumentFromMessageArgument converts an IOSCMessageArgument into the internal osc MessageArgument (any).
func OSCArgumentFromMessageArgument(arg usecaseifs.IOSCMessageArgument) (osc.Argument, error) {
	switch arg.GetType() {
	case ArgTypeInt:
		intVal, err := strconv.ParseInt(arg.GetValue(), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to convert string to int32: %w", err)
		}
		return osc.Int(intVal), nil

	case ArgTypeFloat:
		floatVal, err := strconv.ParseFloat(arg.GetValue(), 32)
		if err != nil {
			return nil, fmt.Errorf("failed to convert string to float32: %w", err)
		}
		return osc.Float(floatVal), nil

	case ArgTypeBool:
		if arg.GetValue() == "true" {
			return osc.Bool(true), nil
		}
		return osc.Bool(false), nil

	case ArgTypeString:
		return osc.String(arg.GetValue()), nil

	// case "blob":
	//	osc.TypetagBlob:

	default:
		return nil, fmt.Errorf("unsupported type: %s", arg.GetType())
	}
}
