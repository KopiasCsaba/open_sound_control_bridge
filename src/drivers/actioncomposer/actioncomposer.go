package actioncomposer

import (
	"fmt"

	"net.kopias.oscbridge/app/adapters/config"
	"net.kopias.oscbridge/app/entities"
	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

// ActionComposer is responsible for composing actions from the YAML structure into an instance-structure.
type ActionComposer struct {
	// actions holds the name->configured action structure pairs
	actions map[string]config.Action

	// conditions holds name->factory pairs for the conditions.
	conditions map[string]usecaseifs.ActionConditionFactory

	// tasks holds name->factory pairs for the tasks.
	tasks map[string]usecaseifs.ActionTaskFactory
}

func NewActionComposer(
	actions map[string]config.Action,
	conditions map[string]usecaseifs.ActionConditionFactory,
	tasks map[string]usecaseifs.ActionTaskFactory,
) *ActionComposer {
	return &ActionComposer{
		actions:    actions,
		conditions: conditions,
		tasks:      tasks,
	}
}

// GetActionList processes the input actions and returns a list of IActions.
func (a *ActionComposer) GetActionList() ([]usecaseifs.IAction, error) {
	actionList := []usecaseifs.IAction{}
	childIndex := 0

	for actionName, cfgAction := range a.actions {
		// Convert the conditions for this action.
		condition, err := a.convertCondition(cfgAction.TriggerChain, actionName, childIndex)
		if err != nil {
			return nil, err
		}

		// Validate the condition parameters.
		if err := condition.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate %s's triggers: %w", actionName, err)
		}

		// Convert the tasks for this action.
		tasks, err := a.convertTasks(actionName, cfgAction.Tasks)
		if err != nil {
			return nil, fmt.Errorf("failed to load tasks: %w", err)
		}

		actionList = append(actionList, entities.NewAction(actionName, condition, tasks, cfgAction.DebounceMillis))
		childIndex++
	}
	return actionList, nil
}

// convertCondition instantiates and configures a single condition
// [path] tracks the hierarchy of the conditions, used for logging.
// [index] reveals the child-index of the current node.
func (a *ActionComposer) convertCondition(chain config.ActionConditionChecker, path string, index int) (usecaseifs.IActionCondition, error) {
	factory, ok := a.conditions[chain.Type]
	if !ok {
		return nil, fmt.Errorf("there are no condition implementation for type '%s'", chain.Type)
	}

	currentPath := fmt.Sprintf("%s/%s:%d", path, chain.Type, index)
	cond := factory(currentPath)
	cond.SetParameters(chain.Parameters)

	for childIndex, child := range chain.Children {
		children, err := a.convertCondition(child, currentPath, childIndex)
		if err != nil {
			return nil, err
		}
		cond.AddChild(children)
	}
	return cond, nil
}

// convertTasks instantiates and configures the tasks.
func (a *ActionComposer) convertTasks(actionName string, tasks []config.ActionTask) ([]usecaseifs.IActionTask, error) {
	result := []usecaseifs.IActionTask{}

	for i, task := range tasks {
		taskFactory, ok := a.tasks[task.Type]
		if !ok {
			return nil, fmt.Errorf("no such task type registered: %s", task.Type)
		}

		newTask := taskFactory()
		newTask.SetParameters(task.Parameters)
		if err := newTask.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate %s action's [%d-%s] task: %w", actionName, i, task.Type, err)
		}
		result = append(result, newTask)
	}
	return result, nil
}
