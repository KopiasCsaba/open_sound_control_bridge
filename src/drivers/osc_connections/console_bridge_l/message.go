package console_bridge_l

import (
	"fmt"
	"strconv"

	"github.com/loffa/gosc"
	"net.kopias.oscbridge/app/drivers/osc_message"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

// MessageFromOSCMessage converts the internal gosc Message to an IOSCMessage.
func MessageFromOSCMessage(oscMsg gosc.Message) (usecaseifs.IOSCMessage, error) {
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

// OSCMessageFromMessage converts an IOSCMessage into the internal gosc Message.
func OSCMessageFromMessage(msg usecaseifs.IOSCMessage) (*gosc.Message, error) {
	oscMessage := gosc.Message{}
	oscMessage.Address = msg.GetAddress()

	arguments := []any{}

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

// MessageArgumentFromOSCArgument converts the internal gosc Message to an IOSCMessageArgument.
func MessageArgumentFromOSCArgument(arg any) (usecaseifs.IOSCMessageArgument, error) {
	msgType := fmt.Sprintf("%T", arg)
	var msgValue string
	switch t := arg.(type) {
	case int32:
		msgValue = fmt.Sprintf("%d", t)
	case float32:
		msgValue = fmt.Sprintf("%f", t)
	case string:
		msgValue = t
	default:
		return nil, fmt.Errorf("response type %T is not supported", arg)
	}

	return osc_message.NewMessageArgument(msgType, msgValue), nil
}

// OSCArgumentFromMessageArgument converts an IOSCMessageArgument into the internal gosc MessageArgument (any).
func OSCArgumentFromMessageArgument(arg usecaseifs.IOSCMessageArgument) (any, error) {
	var v any
	switch arg.GetType() {
	case "int32":
		i64, err := strconv.ParseInt(arg.GetValue(), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message argument to int32, it should be %s but seems to be %T", arg.GetType(), arg.GetValue())
		}
		v = int32(i64)
	case "float32":
		f64, err := strconv.ParseFloat(arg.GetValue(), 32)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message argument to int32, it should be %s but seems to be %T", arg.GetType(), arg.GetValue())
		}
		v = float32(f64)
	case "string":
		v = arg.GetValue()
	default:
		return nil, fmt.Errorf("argument type %T is not supported", arg)
	}
	return v, nil
}
