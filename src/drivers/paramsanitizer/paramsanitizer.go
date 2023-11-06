package paramsanitizer

import (
	"fmt"
	"regexp"
	"strings"

	"net.kopias.oscbridge/app/pkg/slicetools"
)

type ParameterDefinition struct {
	// Name is the name of the parameter.
	Name string

	// Optional determines if specifying this variable is required or not.
	Optional bool

	// DefaultValue is in effect, when the parameter is optional.
	DefaultValue interface{}

	// ValuePattern if specified determines a regexp that must match the given value.
	ValuePattern string

	// Type is a list of golang types, in the form as sprintf would say it for %T.
	Type []string
}

// SanitizeParams verifies the incoming parameters and enforces specific rules on them.
func SanitizeParams(parameters map[string]interface{}, keysAndTypes []ParameterDefinition) (map[string]interface{}, error) {
	sanitized := map[string]interface{}{}

	for _, pd := range keysAndTypes {
		value, found := parameters[pd.Name]

		if !found && !pd.Optional {
			return nil, fmt.Errorf("parameter '%s' is not specified", pd.Name)
		}

		if found {
			valueType := fmt.Sprintf("%T", value)

			if slicetools.IndexOf(pd.Type, valueType) == -1 {
				return nil, fmt.Errorf("parameter '%s' is of a wrong type (%s). Allowed types: %s", pd.Name, valueType, strings.Join(pd.Type, ", "))
			}

			if pd.ValuePattern != "" {
				valueString, ok := value.(string)
				if !ok {
					return nil, fmt.Errorf("failed to cast ValuePattern to string")
				}
				if !regexp.MustCompile(pd.ValuePattern).MatchString(valueString) {
					return nil, fmt.Errorf("parameter '%s' value does not match: %s", pd.Name, pd.ValuePattern)
				}
			}
			sanitized[pd.Name] = value
		} else {
			sanitized[pd.Name] = pd.DefaultValue
		}
	}

	return sanitized, nil
}
