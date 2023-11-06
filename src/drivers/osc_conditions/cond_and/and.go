package cond_and

import (
	"context"
	"fmt"

	"net.kopias.oscbridge/app/drivers/osc_conditions"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionCondition = &AndCondition{}

// AndCondition makes sure all it's children return true for matching the current store.
type AndCondition struct {
	path             string
	children         []usecaseifs.IActionCondition
	conditionTracker *osc_conditions.ConditionTracker
}

func NewFactory(conditionTracker *osc_conditions.ConditionTracker) usecaseifs.ActionConditionFactory {
	return func(path string) usecaseifs.IActionCondition {
		return &AndCondition{path: path, conditionTracker: conditionTracker}
	}
}

func (a *AndCondition) GetType() string {
	return "AND"
}

func (a *AndCondition) Evaluate(ctx context.Context, store usecaseifs.IMessageStore) (bool, error) {
	for i, child := range a.children {
		a.conditionTracker.Log(ctx, a.path, "Checking on child %d...", i)
		matched, err := child.Evaluate(ctx, store)
		if err != nil {
			return a.conditionTracker.R(ctx, false, a.path, "Child %d evaluation failed: %s", i, err.Error()), err
		}

		if !matched {
			return a.conditionTracker.R(ctx, false, a.path, "Child %d returned false", i), nil
		}
	}

	return a.conditionTracker.R(ctx, true, a.path, "all child returned true"), nil
}

func (a *AndCondition) SetParameters(_ map[string]interface{}) {
	// noop
}

func (a *AndCondition) AddChild(condition usecaseifs.IActionCondition) {
	a.children = append(a.children, condition)
}

func (a *AndCondition) Validate() error {
	if len(a.children) == 0 {
		return fmt.Errorf("%s: this node has no children", a.GetType())
	}

	for _, child := range a.children {
		if err := child.Validate(); err != nil {
			return fmt.Errorf("AND failed to validate it's children: %w", err)
		}
	}

	return nil
}
