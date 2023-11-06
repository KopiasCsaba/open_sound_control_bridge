package stringtools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// GenerateUID returns a compressed uid like string in the length of [length].
func GenerateUID(length int) string {
	if length > 32 {
		length = 32
	}
	return strings.ReplaceAll(uuid.New().String(), "-", "")[0:length]
}

// JSONPrettyPrint pretty-prints the given object in JSON format.
func JSONPrettyPrint(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", fmt.Errorf("unable to marshal json: %w", err)
	}
	return string(val), nil
}
