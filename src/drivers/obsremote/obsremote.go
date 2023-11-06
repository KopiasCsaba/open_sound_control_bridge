package obsremote

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IOBSRemote = &OBSRemote{}

type Config struct {
	Host     string
	Port     int64
	Password string
	Debug    bool
}

// OBSRemote uses goobs to connect to an OBS instance based on the config, and to provide the IOBSRemote interface.
type OBSRemote struct {
	client *goobs.Client
	logger usecaseifs.ILogger
	cfg    Config

	notify chan error
	quit   chan interface{}

	// Locking is introduced because goobs is not trhead safe apparently,
	// produces mismatched id errors if handled from multiple threads.
	m *sync.Mutex
}

func NewOBSRemote(logger usecaseifs.ILogger, cfg Config) *OBSRemote {
	return &OBSRemote{
		logger: logger,
		cfg:    cfg,
		notify: make(chan error, 1),
		quit:   make(chan interface{}, 1),
		m:      &sync.Mutex{},
	}
}

func (or *OBSRemote) Start(ctx context.Context) error {
	var err error
	or.m.Lock()

	opts := []goobs.Option{}

	if or.cfg.Password != "" {
		opts = append(opts, goobs.WithPassword(or.cfg.Password))
	}

	opts = append(opts, goobs.WithLogger(newOBSRemoteLogger(ctx, or.logger, or.cfg.Debug)))

	or.client, err = goobs.New(fmt.Sprintf("%s:%d", or.cfg.Host, or.cfg.Port), opts...)

	or.m.Unlock()

	if err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}

	if err := or.checkConnection(ctx); err != nil {
		return err
	}
	go or.watchdog(ctx)
	return nil
}

func (or *OBSRemote) Stop(ctx context.Context) {
	or.m.Lock()
	defer or.m.Unlock()

	if or.quit == nil {
		return
	}
	close(or.quit)
	or.quit = nil
	if or.client == nil {
		or.logger.Err(ctx, fmt.Errorf("not connected"))
	}

	if err := or.client.Disconnect(); err != nil {
		or.logger.Err(ctx, err)
	}
}

func (or *OBSRemote) checkConnection(ctx context.Context) error {
	or.m.Lock()
	defer or.m.Unlock()

	_, err := or.client.General.GetVersion()
	if err != nil {
		return err
	}
	return nil
}

func (or *OBSRemote) watchdog(ctx context.Context) {
	for {
		if or.cfg.Debug {
			or.logger.Infof(ctx, "OBS remote checking connection...")
		}
		err := or.checkConnection(ctx)
		if err != nil {
			or.Stop(ctx)
			or.notify <- fmt.Errorf("obsRemote connection is broken: %w", err)
		}

		select {
		case <-or.quit:
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (or *OBSRemote) Notify() chan error {
	return or.notify
}

func (or *OBSRemote) GetIncomingEvents() (chan interface{}, error) {
	if or.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	return or.client.IncomingEvents, nil
}
