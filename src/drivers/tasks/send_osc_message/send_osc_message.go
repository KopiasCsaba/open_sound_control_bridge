package send_osc_message

import (
	"context"
	"fmt"

	"net.kopias.oscbridge/app/drivers/osc_message"

	"net.kopias.oscbridge/app/drivers/paramsanitizer"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionTask = &SendOscMessageTask{}

// SendOscMessageTask sends an OSC Message to the named connection.
type SendOscMessageTask struct {
	log         usecaseifs.ILogger
	debug       bool
	configError error
	connections map[string]usecaseifs.IOSCConnection
	connection  string
	address     string
	arguments   []argument
}

type argument struct {
	variableType  string
	variableValue string
}

const (
	ParamConnectionKey = "connection"
	ParamAddress       = "address"
	ParamArguments     = "arguments"
	ParamArgumentType  = "type"
	ParamArgumentValue = "value"
)

func NewFactory(log usecaseifs.ILogger, debug bool, connections map[string]usecaseifs.IOSCConnection) usecaseifs.ActionTaskFactory {
	return func() usecaseifs.IActionTask {
		return &SendOscMessageTask{log: log, debug: debug, connections: connections}
	}
}

func (o *SendOscMessageTask) SetParameters(m map[string]interface{}) {
	sanitized, err := paramsanitizer.SanitizeParams(m, []paramsanitizer.ParameterDefinition{
		{
			Name:     ParamConnectionKey,
			Optional: false,
			Type:     []string{"string"},
		}, {
			Name:     ParamAddress,
			Optional: false,
			Type:     []string{"string"},
		}, {
			Name:     ParamArguments,
			Optional: true,
			Type:     []string{"[]interface {}"},
		},
	})
	if err != nil {
		o.configError = fmt.Errorf("failed to verify parameters: %w", err)
		return
	}

	// nolint:forcetypeassert
	o.connection = sanitized[ParamConnectionKey].(string)

	// nolint:forcetypeassert
	o.address = sanitized[ParamAddress].(string)

	args, ok := sanitized[ParamArguments]
	if !ok {
		o.configError = fmt.Errorf("key %s was not found", ParamArguments)
		return
	}
	// nolint:forcetypeassert
	argsSlice := args.([]interface{})

	for i, argParams := range argsSlice {
		err = o.setArgumentParameters(argParams)
		if err != nil {
			o.configError = fmt.Errorf("osc_message task failed to verify parameters: Agrgument[%d]: %w", i, err)
			return
		}
	}
}

func (o *SendOscMessageTask) setArgumentParameters(m interface{}) error {
	mCasted, ok := m.(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to cast supplied arguments")
	}
	sanitized, err := paramsanitizer.SanitizeParams(mCasted, []paramsanitizer.ParameterDefinition{
		{
			Name:     ParamArgumentType,
			Optional: false,
			Type:     []string{"string"},
		},
		{
			Name:     ParamArgumentValue,
			Optional: false,
			Type:     []string{"string"},
		},
	})
	if err != nil {
		return err
	}

	newArg := argument{}

	// nolint:forcetypeassert
	newArg.variableType = sanitized[ParamArgumentType].(string)
	// nolint:forcetypeassert
	newArg.variableValue = sanitized[ParamArgumentValue].(string)

	o.arguments = append(o.arguments, newArg)
	return nil
}

func (o *SendOscMessageTask) Validate() error {
	return o.configError
}

func (o *SendOscMessageTask) Execute(ctx context.Context, store usecaseifs.IMessageStore) error {
	o.log.Infof(ctx, "\tExecuting task: OSC Message send")

	conn, ok := o.connections[o.connection]
	if !ok {
		return fmt.Errorf("there is no osc connection named '%s'", o.connection)
	}

	args := []usecaseifs.IOSCMessageArgument{}

	for _, a := range o.arguments {
		arg := osc_message.NewMessageArgument(a.variableType, a.variableValue)
		args = append(args, arg)
	}
	msg := osc_message.NewMessage(o.address, args)

	err := conn.SendMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %s: %w", msg.String(), err)
	}

	return nil
}
