package cond_or

import (
	"context"
	"fmt"

	"net.kopias.oscbridge/app/drivers/osc_conditions"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionCondition = &OrCondition{}

type OrCondition struct {
	path             string
	children         []usecaseifs.IActionCondition
	conditionTracker *osc_conditions.ConditionTracker
}

func NewFactory(conditionTracker *osc_conditions.ConditionTracker) usecaseifs.ActionConditionFactory {
	return func(path string) usecaseifs.IActionCondition {
		return &OrCondition{path: path, conditionTracker: conditionTracker}
	}
}

func (a *OrCondition) GetType() string {
	return "OR"
}

func (a *OrCondition) Evaluate(ctx context.Context, store usecaseifs.IMessageStore) (bool, error) {
	for i, child := range a.children {
		a.conditionTracker.Log(ctx, a.path, "Checking on child %d...", i)
		matched, err := child.Evaluate(ctx, store)
		if err != nil {
			return a.conditionTracker.R(ctx, false, a.path, "Child %d evaluation failed: %s", i, err.Error()), err
		}
		if matched {
			return a.conditionTracker.R(ctx, true, a.path, "Child %d returned true", i), nil
		}
	}
	return a.conditionTracker.R(ctx, false, a.path, "no child returned true"), nil
}

func (a *OrCondition) SetParameters(_ map[string]interface{}) {
	// noop
}

func (a *OrCondition) AddChild(condition usecaseifs.IActionCondition) {
	a.children = append(a.children, condition)
}

func (a *OrCondition) Validate() error {
	if len(a.children) == 0 {
		return fmt.Errorf("%s: this node has no children", a.GetType())
	}

	for _, child := range a.children {
		if err := child.Validate(); err != nil {
			return fmt.Errorf("OR failed to validate it's children: %w", err)
		}
	}
	return nil
}
