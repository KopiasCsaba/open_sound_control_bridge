package entities

import (
	"context"
	"fmt"
	"strings"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IAction = &Action{}

// Action represents a living, composed set of instances of triggers and tasks
type Action struct {
	name string
	// triggerChain is a tree of conditions
	triggerChain usecaseifs.IActionCondition

	// tasks is a list of tasks that must be executed serially.
	tasks []usecaseifs.IActionTask

	// debounceMillis causes repeated evaluation with this delay to see if the condition is still true.
	debounceMillis int64
}

func (a *Action) GetDebounceMillis() int64 {
	return a.debounceMillis
}

func NewAction(name string, triggerChain usecaseifs.IActionCondition, tasks []usecaseifs.IActionTask, debounceMillis int64) *Action {
	return &Action{
		name:           name,
		triggerChain:   triggerChain,
		tasks:          tasks,
		debounceMillis: debounceMillis,
	}
}

func (a *Action) Evaluate(ctx context.Context, store usecaseifs.IMessageStore) (bool, error) {
	matched, err := a.triggerChain.Evaluate(ctx, store)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate trigger chain: %w", err)
	}
	return matched, nil
}

func (a *Action) Execute(ctx context.Context, store usecaseifs.IMessageStore) error {
	errs := []any{}
	for i, task := range a.tasks {
		if err := task.Execute(ctx, store); err != nil {
			errs = append(errs, fmt.Errorf("failed to execute task %d: %w", i, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to exectue task(s) "+strings.Repeat(": %w", len(errs)), errs...)
	}

	return nil
}

func (a *Action) GetName() string {
	return a.name
}
