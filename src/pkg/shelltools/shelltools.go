package shelltools

import (
	"fmt"
	"strings"

	"net.kopias.oscbridge/app/pkg/slicetools"
)

// EscapeShellArg is based on the same function in php
// https://github.com/php/php-src/blob/master/ext/standard/exec.c#L388
func EscapeShellArg(arg string) (string, error) {
	if len(arg) > 4096-2-1 {
		return "", fmt.Errorf("argument exceeds the allowed length of 4096 bytes. (%s)", arg)
	}
	result := strings.Builder{}
	result.WriteString("'")
	argValidUtf8 := []rune(strings.ToValidUTF8(arg, ""))

	//nolint:staticcheck,gosimple
	for _, r := range argValidUtf8 {
		switch r {
		case []rune("'")[0]:
			result.WriteRune(r)
			result.WriteString("\\")
			result.WriteRune(r)
			result.WriteRune(r)
		default:
			result.WriteRune(r)
		}
	}
	result.WriteString("'")

	if result.Len() > 4096-1 {
		return "", fmt.Errorf("escaped argument exceeds the allowed length of 4096 bytes (%s)", arg)
	}
	return result.String(), nil
}

func EscapeShellArgImplode(args []string) (string, error) {
	var err error
	var escaped string

	result := strings.Join(
		slicetools.Map(args, func(arg string) string {
			if err != nil {
				return ""
			}
			escaped, err = EscapeShellArg(arg)
			return escaped
		}),
		" ")

	if err != nil {
		return "", err
	}
	return result, nil
}
