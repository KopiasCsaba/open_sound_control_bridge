package osc_conditions

import (
	"context"
	"fmt"
	"strings"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

// ConditionTracker is a utility for tracking condition calls and debug why something matched or not.
type ConditionTracker struct {
	log     usecaseifs.ILogger
	enabled bool
}

func NewConditionTracker(log usecaseifs.ILogger, enabled bool) *ConditionTracker {
	return &ConditionTracker{log: log, enabled: enabled}
}

func (ct *ConditionTracker) Log(ctx context.Context, prefix string, message string, args ...interface{}) {
	if ct.enabled {
		ct.log.Debugf(ctx, prefix+" "+message, args...)
	}
}

func (ct *ConditionTracker) R(ctx context.Context, ret bool, prefix string, message string, args ...interface{}) bool {
	ct.Log(ctx, prefix, "returned %s because %s.", strings.ToUpper(fmt.Sprintf("%t", ret)), fmt.Sprintf(message, args...))
	return ret
}
