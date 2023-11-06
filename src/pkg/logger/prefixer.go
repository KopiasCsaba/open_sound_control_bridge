package logger

import (
	"context"
	"fmt"
	"strings"

	"net.kopias.oscbridge/app/pkg/slicetools"
)

type LogPrefixer interface {
	GetContextualLogPrefixer(ctx context.Context) []string
}

// GetPrefixesFunc is a function that returns an array of strings in regard to a context.
// For example a http context might contain a request id. The returned array of strings will be used as a prefix in the log lines.
type GetPrefixesFunc func(context.Context) []string

// AddPrefixerFunc adds a PrefixerFunc which will be called upon printing log messages.
// The PrefixerFunc-s can access the context, and extract any value from it and return it as a log prefix.
func (l *Logger) AddPrefixerFunc(prefixer GetPrefixesFunc) {
	l.prefixers = append(l.prefixers, prefixer)
}

// AddPrefixer adds a special Prefixer which will be called upon printing log messages.
// It is a shorthand for AddPrefixerFunc to extract simple string keys from the context.
func (l *Logger) AddPrefixer(key interface{}) {
	l.AddPrefixerFunc(func(ctx context.Context) []string {
		p := ctx.Value(key)
		if p != nil {
			pString, ok := p.(string)
			if ok {
				return []string{pString}
			}
		}
		return nil
	})
}

// GetPrefixForContext returns all the prefixes merged, that can be extracted from the contexts with the prefixers.
func (l *Logger) GetPrefixForContext(ctx context.Context) string {
	prefixString := ""
	prefixes := []string{}
	if ctx != nil {
		for i := range l.prefixers {
			prefixerResults := l.prefixers[i](ctx)
			prefixerResults = slicetools.Filter(prefixerResults, func(s string) bool {
				return s != ""
			})

			if len(prefixerResults) > 0 {
				prefixes = append(prefixes, prefixerResults...)
			}
		}

		if len(prefixes) > 0 {
			prefixString = fmt.Sprintf(" %s", strings.Join(prefixes, "/"))
		}
	}
	return prefixString
}
