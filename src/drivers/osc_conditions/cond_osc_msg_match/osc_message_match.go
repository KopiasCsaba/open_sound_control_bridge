package cond_osc_msg_match

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"net.kopias.oscbridge/app/drivers/osc_conditions"

	"net.kopias.oscbridge/app/drivers/paramsanitizer"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionCondition = &OSCCondition{}

const (
	AddressKey          = "address"
	AddressMatchTypeKey = "address_match_type"
	ArgumentsKey        = "arguments"
	TriggerOnChangeKey  = "trigger_on_change"

	AddressMatchTypeEq     = "eq"
	AddressMatchTypeRegexp = "regexp"

	ArgIndexKey          = "index"
	ArgTypeKey           = "type"
	ArgValueKey          = "value"
	ArgValueMatchTypeKey = "value_match_type"

	ValueMatchTypeRegexp = "regexp"
	ValueMatchTypeEq     = "="
	ValueMatchTypeLTE    = "<="
	ValueMatchTypeGTE    = ">="
	ValueMatchTypeLT     = "<"
	ValueMatchTypeGT     = ">"
	ValueMatchTypeNOT    = "!="
	// @TODO ADD MOD
)

type argumentCondition struct {
	index                  int
	variableType           string
	variableValue          string
	variableValueRegexp    *regexp.Regexp
	variableValueMatchType string
}

func (ac argumentCondition) String() string {
	return fmt.Sprintf("ArgumentCondition(type: %s, value: %s, matchType: %s)", ac.variableType, ac.variableValue, ac.variableValueMatchType)
}

// OSCCondition matches an entire OSC Message by address and arguments if applicable.
type OSCCondition struct {
	path     string
	children []usecaseifs.IActionCondition

	configError error

	addressPattern string

	addressMatchType string
	argumentPatterns []argumentCondition
	triggerOnChange  bool
	conditionTracker *osc_conditions.ConditionTracker
}

func NewFactory(conditionTracker *osc_conditions.ConditionTracker) usecaseifs.ActionConditionFactory {
	return func(path string) usecaseifs.IActionCondition {
		return &OSCCondition{path: path, conditionTracker: conditionTracker}
	}
}

func (a *OSCCondition) SetParameters(m map[string]interface{}) {
	sanitized, err := paramsanitizer.SanitizeParams(m, []paramsanitizer.ParameterDefinition{
		{
			Name:         AddressKey,
			Optional:     false,
			DefaultValue: nil,
			Type:         []string{"string"},
		}, {
			Name:         AddressMatchTypeKey,
			Optional:     true,
			DefaultValue: AddressMatchTypeEq,
			ValuePattern: fmt.Sprintf("^%s|%s$", AddressMatchTypeEq, AddressMatchTypeRegexp),
			Type:         []string{"string"},
		}, {
			Name:         TriggerOnChangeKey,
			Optional:     true,
			DefaultValue: true,
			Type:         []string{"bool"},
		}, {
			Name:     ArgumentsKey,
			Optional: true,
			Type:     []string{"[]interface {}"},
		},
	})
	if err != nil {
		a.configError = fmt.Errorf("%s failed to verify parameters: %w", a.path, err)
		return
	}

	// nolint:forcetypeassert
	a.addressPattern = sanitized[AddressKey].(string)

	// nolint:forcetypeassert
	a.addressMatchType = sanitized[AddressMatchTypeKey].(string)

	// nolint:forcetypeassert
	a.triggerOnChange = sanitized[TriggerOnChangeKey].(bool)

	if sanitized[AddressMatchTypeKey] == AddressMatchTypeRegexp {
		_, err = regexp.Compile(sanitized[AddressKey].(string))
		if err != nil {
			a.configError = fmt.Errorf("%s is not a valid regexp: %w", AddressKey, err)
			return
		}
	}

	args, ok := sanitized[ArgumentsKey]
	if !ok {
		a.configError = fmt.Errorf("key %s was not found", ArgumentsKey)
		return
	}
	// nolint:forcetypeassert
	argsSlice := args.([]interface{})

	for i, argParams := range argsSlice {
		err = a.setArgumentParameters(argParams)
		if err != nil {
			a.configError = fmt.Errorf("%s failed to verify parameters: Agrgument[%d]: %w", a.path, i, err)
			return
		}
	}
}

func (a *OSCCondition) setArgumentParameters(m interface{}) error {
	mCasted, ok := m.(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to cast supplied arguments")
	}
	sanitized, err := paramsanitizer.SanitizeParams(mCasted, []paramsanitizer.ParameterDefinition{
		{
			Name:     ArgIndexKey,
			Optional: false,
			Type:     []string{"int"},
		},
		{
			Name:         ArgTypeKey,
			Optional:     false,
			ValuePattern: "^string|int|float$",
			Type:         []string{"string"},
		},
		{
			Name:     ArgValueKey,
			Optional: false,
			Type:     []string{"string"},
		},
		{
			Name:         ArgValueMatchTypeKey,
			Optional:     true,
			DefaultValue: ValueMatchTypeEq,
			ValuePattern: fmt.Sprintf("^%s$", strings.Join([]string{
				ValueMatchTypeEq,
				ValueMatchTypeRegexp,
				ValueMatchTypeLTE,
				ValueMatchTypeGTE,
				ValueMatchTypeLT,
				ValueMatchTypeGT,
				ValueMatchTypeNOT,
			}, "|")),
			Type: []string{"string"},
		},
	})
	if err != nil {
		return err
	}

	newArgCondition := argumentCondition{}

	// nolint:forcetypeassert
	newArgCondition.index = sanitized[ArgIndexKey].(int)
	// nolint:forcetypeassert
	newArgCondition.variableType = sanitized[ArgTypeKey].(string)
	// nolint:forcetypeassert
	newArgCondition.variableValue = sanitized[ArgValueKey].(string)
	// nolint:forcetypeassert
	newArgCondition.variableValueMatchType = sanitized[ArgValueMatchTypeKey].(string)

	if newArgCondition.variableValueMatchType == ValueMatchTypeRegexp {
		newArgCondition.variableValueRegexp, err = regexp.Compile(newArgCondition.variableValue)
		if err != nil {
			return fmt.Errorf("failed to compile value regexp: %s: %w", newArgCondition.variableValue, err)
		}
	}

	a.argumentPatterns = append(a.argumentPatterns, newArgCondition)
	return nil
}

func (a *OSCCondition) GetType() string {
	return "MATCH"
}

func (a *OSCCondition) Evaluate(ctx context.Context, store usecaseifs.IMessageStore) (bool, error) {
	var err error
	var record usecaseifs.IMessageStoreRecord
	var found bool

	if a.addressMatchType == AddressMatchTypeRegexp {
		record, err = store.GetOneRecordByRegexp(a.addressPattern, a.triggerOnChange)
		if err != nil {
			err = fmt.Errorf("failed to get record by regexp: %w", err)
			return a.conditionTracker.R(ctx, false, a.path, err.Error()), err
		}
		if record == nil {
			return a.conditionTracker.R(ctx, false, a.path, "record not found by regexp on address: %s", a.addressPattern), nil
		}
	} else {
		record, found = store.GetRecord(a.addressPattern, a.triggerOnChange)
		if !found {
			return a.conditionTracker.R(ctx, false, a.path, "record not found by exact match: %s", a.addressPattern), nil
		}
	}

	var matched bool
	for i, ap := range a.argumentPatterns {
		matched, err = a.matchArguments(record, ap)
		if err != nil {
			err = fmt.Errorf("failed to match argument[%d]: %w", i, err)
			return a.conditionTracker.R(ctx, false, a.path, err.Error()), err
		}
		if !matched {
			return a.conditionTracker.R(ctx, false, a.path, "argument %d did not match '%s'", i, ap.String()), nil
		}
	}
	return a.conditionTracker.R(ctx, true, a.path, "all checks passed"), nil
}

// nolint: unparam,cyclop
func (a *OSCCondition) matchArguments(record usecaseifs.IMessageStoreRecord, ac argumentCondition) (bool, error) {
	// If it has no argument with the specified index
	if len(record.GetMessage().GetArguments())-1 < ac.index {
		return false, nil
	}
	arg := record.GetMessage().GetArguments()[ac.index]

	if arg.GetType() != ac.variableType {
		return false, nil
	}

	switch ac.variableValueMatchType {
	case ValueMatchTypeEq:
		if arg.GetValue() != ac.variableValue {
			return false, nil
		}
	case ValueMatchTypeRegexp:
		if !ac.variableValueRegexp.MatchString(arg.GetValue()) {
			return false, nil
		}
	case ValueMatchTypeLTE:
		if arg.GetValue() > ac.variableValue {
			return false, nil
		}
	case ValueMatchTypeGTE:
		if arg.GetValue() < ac.variableValue {
			return false, nil
		}
	case ValueMatchTypeLT:
		if arg.GetValue() >= ac.variableValue {
			return false, nil
		}
	case ValueMatchTypeGT:
		if arg.GetValue() <= ac.variableValue {
			return false, nil
		}
	case ValueMatchTypeNOT:
		if arg.GetValue() == ac.variableValue {
			return false, nil
		}
	}

	return true, nil
}

func (a *OSCCondition) AddChild(condition usecaseifs.IActionCondition) {
	a.children = append(a.children, condition)
}

func (a *OSCCondition) Validate() error {
	if len(a.children) != 0 {
		return fmt.Errorf("this node can not have children")
	}
	if a.configError != nil {
		return a.configError
	}

	return nil
}
