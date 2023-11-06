package usecase

import (
	"context"
	"fmt"

	"net.kopias.oscbridge/app/pkg/stringtools"
)

type contextKey string

func getTaskExecutionSessionContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey("task_exec_session"), stringtools.GenerateUID(6))
}

func GetContextualLogPrefixer(ctx context.Context) []string {
	result := []string{}

	if value := ctx.Value(contextKey("task_exec_session")); value != nil {
		result = append(result, fmt.Sprintf("T:%s", value))
	}

	return result
}
