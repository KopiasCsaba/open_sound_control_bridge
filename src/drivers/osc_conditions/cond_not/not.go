package cond_not

import (
	"context"
	"fmt"

	"net.kopias.oscbridge/app/drivers/osc_conditions"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionCondition = &NotCondition{}

// NotCondition simply negates its SINGLE children's result for Evaluate.
type NotCondition struct {
	path             string
	children         []usecaseifs.IActionCondition
	conditionTracker *osc_conditions.ConditionTracker
}

func NewFactory(conditionTracker *osc_conditions.ConditionTracker) usecaseifs.ActionConditionFactory {
	return func(path string) usecaseifs.IActionCondition {
		return &NotCondition{path: path, conditionTracker: conditionTracker}
	}
}

func (a *NotCondition) GetType() string {
	return "NOT"
}

func (a *NotCondition) Evaluate(ctx context.Context, store usecaseifs.IMessageStore) (bool, error) {
	matched, err := a.children[0].Evaluate(ctx, store)
	if err != nil {
		return a.conditionTracker.R(ctx, false, a.path, "Child evaluation failed: %s", err.Error()), err
	}

	return a.conditionTracker.R(ctx, !matched, a.path, "Child returned %t", matched), nil
}

func (a *NotCondition) SetParameters(_ map[string]interface{}) {
	// noop
}

func (a *NotCondition) AddChild(condition usecaseifs.IActionCondition) {
	a.children = append(a.children, condition)
}

func (a *NotCondition) Validate() error {
	if len(a.children) == 0 {
		return fmt.Errorf("%s: this node has no children", a.GetType())
	}
	if len(a.children) > 1 {
		return fmt.Errorf("%s: this node has more than one children", a.GetType())
	}

	for _, child := range a.children {
		if err := child.Validate(); err != nil {
			return fmt.Errorf("OR failed to validate it's children: %w", err)
		}
	}
	return nil
}
