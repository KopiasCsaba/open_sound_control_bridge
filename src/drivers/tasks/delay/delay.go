package delay

import (
	"context"
	"fmt"
	"time"

	"net.kopias.oscbridge/app/pkg/maptools"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionTask = &Delay{}

// Delay simply pauses the execution of the next tasks by the specified amount of time.
type Delay struct {
	parameters map[string]interface{}
	log        usecaseifs.ILogger
	debug      bool
}

const (
	ParamDelayMillis = "delay_millis"
)

func NewFactory(log usecaseifs.ILogger, debug bool) usecaseifs.ActionTaskFactory {
	return func() usecaseifs.IActionTask { return &Delay{log: log, debug: debug} }
}

func (o *Delay) Validate() error {
	_, err := maptools.GetIntValue(o.parameters, ParamDelayMillis)
	if err != nil {
		return fmt.Errorf("unable to validate delay_millis: %w", err)
	}
	return nil
}

func (o *Delay) Execute(ctx context.Context, store usecaseifs.IMessageStore) error {
	o.log.Infof(ctx, "\tExecuting task: delay")
	delayMillis, err := maptools.GetIntValue(o.parameters, ParamDelayMillis)
	if err != nil {
		return fmt.Errorf("unable to validate delay_millis: %w", err)
	}
	if o.debug {
		o.log.Debugf(ctx, "Waiting %d milliseconds.", delayMillis)
	}

	// All that code for a bit of sleep...
	time.Sleep(time.Millisecond * time.Duration(delayMillis))
	if o.debug {
		o.log.Debugf(ctx, "Waiting %d milliseconds is over.", delayMillis)
	}

	return nil
}

func (o *Delay) SetParameters(m map[string]interface{}) {
	o.parameters = m
}
