package maptools

import "fmt"

func GetKeys[T comparable](param map[T]any) []T {
	result := []T{}

	for k := range param {
		result = append(result, k)
	}
	return result
}

func GetStringValue[T comparable](m map[T]any, key T) (string, error) {
	value, ok := m[key]
	if !ok {
		return "", fmt.Errorf("map does not contain key '%v'", key)
	}
	vCasted, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("the key's ('%v') value can not be casted to string", key)
	}

	return vCasted, nil
}

func GetStringSliceValue[T comparable](m map[T]any, key T) ([]string, error) {
	value, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("map does not contain key '%v'", key)
	}
	vCasted, ok := value.([]string)
	if !ok {
		return nil, fmt.Errorf("the key's ('%v') value can not be casted to string", key)
	}

	return vCasted, nil
}

func GetInt64Value[T comparable](m map[T]any, key T) (int64, error) {
	value, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("map does not contain key '%v'", key)
	}
	vCasted, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("the key's ('%v') value can not be casted to int64", key)
	}

	return vCasted, nil
}

func GetIntValue[T comparable](m map[T]any, key T) (int, error) {
	value, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("map does not contain key '%v'", key)
	}
	vCasted, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("the key's ('%v') value can not be casted to int", key)
	}

	return vCasted, nil
}

func GetBoolValue[T comparable](m map[T]any, key T) (bool, error) {
	value, ok := m[key]
	if !ok {
		return false, fmt.Errorf("map does not contain key '%v'", key)
	}
	vCasted, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("the key's ('%v') value can not be casted to bool", key)
	}

	return vCasted, nil
}
