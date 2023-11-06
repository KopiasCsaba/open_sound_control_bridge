package usecaseifs

import "context"

// Use cases
type (
	// IUseCases is the public facing api of the application's business logic.
	IUseCases interface {
		Start(context.Context) error
		Stop(context.Context)
		Notify() <-chan error
	}

	IOSCMessageStore interface{}
)
