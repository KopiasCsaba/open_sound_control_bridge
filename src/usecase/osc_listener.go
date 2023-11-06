package usecase

import (
	"context"
	"time"

	"net.kopias.oscbridge/app/entities"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ iusecase = &oscListener{}

// oscListener is watching for new incoming messages from all the different connections.
type oscListener struct {
	ucs *UseCases
	log usecaseifs.ILogger
	cfg usecaseifs.IConfiguration

	oscConnections []entities.OscConnectionDetails
	quit           chan interface{}
}

func newOscListener(log usecaseifs.ILogger, cfg usecaseifs.IConfiguration, oscConnections []entities.OscConnectionDetails) *oscListener {
	return &oscListener{
		log:            log,
		cfg:            cfg,
		oscConnections: oscConnections,
		quit:           make(chan interface{}, 1),
	}
}

func (e *oscListener) setUseCases(ucs *UseCases) {
	e.ucs = ucs
}

func (e *oscListener) Start(ctx context.Context) error {
	go e.listeningLoop(ctx)
	return nil
}

func (e *oscListener) Stop(ctx context.Context) error {
	e.quit <- true
	return nil
}

func (e *oscListener) listeningLoop(ctx context.Context) {
	for {
		//
		for _, cd := range e.oscConnections {
			// fmt.Println("CHECKING ON ", cd.Name)
			select {
			case msg := <-cd.Connection.GetEventChan(ctx):
				// e.log.Infof(ctx, "Incoming message from: %s: %s", cd.Name, msg.String())

				prefixedMessage := entities.NewPrefixedOSCMessage(cd.Prefix, msg)

				e.ucs.oscMessageStore.updateRecord(ctx, prefixedMessage)

			case <-e.quit:
				return

			default:
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}
