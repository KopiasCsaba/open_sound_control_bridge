package run_command

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"net.kopias.oscbridge/app/drivers/paramsanitizer"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionTask = &RunCommandTask{}

// RunCommandTask executes the given command as-is.
type RunCommandTask struct {
	log             usecaseifs.ILogger
	debug           bool
	configError     error
	command         string
	arguments       []string
	runInBackground bool
	directory       string
}

const (
	ParamCommand    = "command"
	ParamArguments  = "arguments"
	ParamBackground = "run_in_background"
	ParamDirectory  = "directory"
)

func NewFactory(log usecaseifs.ILogger, debug bool) usecaseifs.ActionTaskFactory {
	return func() usecaseifs.IActionTask {
		return &RunCommandTask{log: log, debug: debug, arguments: []string{}}
	}
}

func (o *RunCommandTask) SetParameters(m map[string]interface{}) {
	sanitized, err := paramsanitizer.SanitizeParams(m, []paramsanitizer.ParameterDefinition{
		{
			Name:     ParamCommand,
			Optional: false,
			Type:     []string{"string"},
		}, {
			Name:         ParamArguments,
			Optional:     true,
			DefaultValue: []interface{}{},
			Type:         []string{"[]interface {}"},
		}, {
			Name:         ParamBackground,
			Optional:     true,
			DefaultValue: false,
			Type:         []string{"bool"},
		}, {
			Name:         ParamDirectory,
			Optional:     true,
			DefaultValue: "",
			Type:         []string{"string"},
		},
	})
	if err != nil {
		o.configError = fmt.Errorf("failed to verify parameters: %w", err)
		return
	}

	// nolint:forcetypeassert
	o.command = sanitized[ParamCommand].(string)
	// nolint:forcetypeassert
	o.runInBackground = sanitized[ParamBackground].(bool)
	// nolint:forcetypeassert
	o.directory = sanitized[ParamDirectory].(string)

	args, ok := sanitized[ParamArguments]
	if !ok {
		o.configError = fmt.Errorf("key %s was not found", ParamArguments)
		return
	}
	// nolint:forcetypeassert
	argsSlice := args.([]interface{})

	for i, argument := range argsSlice {
		argumentString, ok := argument.(string)
		if !ok {
			o.configError = fmt.Errorf("failed to convert argument %d to string", i)
			return
		}
		o.arguments = append(o.arguments, argumentString)
	}
}

func (o *RunCommandTask) Validate() error {
	return o.configError
}

func (o *RunCommandTask) Execute(ctx context.Context, store usecaseifs.IMessageStore) error {
	o.log.Infof(ctx, "\tExecuting task: Run command")

	if o.runInBackground {
		go o.execute(ctx)
	} else {
		o.execute(ctx)
	}
	return nil
}

func (o *RunCommandTask) execute(ctx context.Context) {
	// nolint: gosec
	cmd := exec.Command(o.command, o.arguments...)
	if o.directory != "" {
		cmd.Dir = o.directory
	}
	err := cmd.Run()
	if err != nil {
		o.log.Err(ctx, fmt.Errorf("failed to execute %s %s: %w", o.command, strings.Join(o.arguments, " "), err))
	}
	o.log.Infof(ctx, "Command exit code: %d", cmd.ProcessState.ExitCode())
}
