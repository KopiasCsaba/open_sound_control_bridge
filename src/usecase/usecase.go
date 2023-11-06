package usecase

import (
	"context"

	"net.kopias.oscbridge/app/entities"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

// Private interfaces
type (
	iusecase interface {
		setUseCases(cases *UseCases)
	}
)

var _ usecaseifs.IUseCases = UseCases{}

// UseCases	are the root to all the usecase groups in the system.
type UseCases struct {
	oscMessageStore *oscMessageStoreManager
	oscListener     *oscListener

	notify chan error
	quit   chan interface{}
	log    usecaseifs.ILogger
}

func New(
	log usecaseifs.ILogger,
	cfg usecaseifs.IConfiguration,
	oscConnections []entities.OscConnectionDetails,
	store usecaseifs.IMessageStore,
	actions []usecaseifs.IAction,
	storePersistPath string,
) *UseCases {
	ucs := &UseCases{
		oscMessageStore: newOscMessageStoreManager(log, cfg, store, actions, storePersistPath),
		oscListener:     newOscListener(log, cfg, oscConnections),

		notify: make(chan error, 1),
		quit:   make(chan interface{}, 1),
		log:    log,
	}

	// The trick here is, to inject the object itself back to every member, then cross-calling is possible within the usecases.
	ucs.oscMessageStore.setUseCases(ucs)
	ucs.oscListener.setUseCases(ucs)

	return ucs
}

func (u UseCases) Start(ctx context.Context) error {
	if err := u.oscMessageStore.Start(ctx); err != nil {
		return err
	}

	return u.oscListener.Start(ctx)
}

func (u UseCases) Stop(ctx context.Context) {
	u.oscMessageStore.Stop(ctx)

	if err := u.oscListener.Stop(ctx); err != nil {
		u.log.Err(ctx, err)
	}

	u.quit <- true
}

func (u UseCases) Notify() <-chan error {
	return u.oscMessageStore.Notify()
}
