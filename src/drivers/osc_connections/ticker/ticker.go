package osc_ticker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"net.kopias.oscbridge/app/drivers/osc_message"

	"net.kopias.oscbridge/app/pkg/chantools"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IOSCConnection = &Ticker{}

type Config struct {
	Debug             bool
	RefreshRateMillis int64
}

// Ticker just emits OSC Messages about the time.
type Ticker struct {
	log      usecaseifs.ILogger
	messages chan usecaseifs.IOSCMessage
	// A channel that shows when the client exited with an error.
	quit   chan any
	cfg    Config
	notify chan error
}

func NewTicker(log usecaseifs.ILogger, cfg Config) usecaseifs.IOSCConnection {
	return &Ticker{
		log:      log,
		cfg:      cfg,
		quit:     make(chan any),
		messages: make(chan usecaseifs.IOSCMessage, 1),
		notify:   make(chan error, 1),
	}
}

func (c *Ticker) Start(ctx context.Context) error {
	go c.run(ctx)
	return nil
}

func (c *Ticker) run(ctx context.Context) {
	t := time.NewTicker(time.Duration(c.cfg.RefreshRateMillis) * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			c.log.Debugf(ctx, "Updating time...")
			c.updateTime(ctx)
		case <-c.quit:
			return
		}
	}
}

// Notify returns the notification channel that can be used to listen for the client's exit
func (c *Ticker) Notify() <-chan error {
	return c.notify
}

func (c *Ticker) Stop(ctx context.Context) {
	if chantools.ChanIsOpenReader(c.quit) {
		close(c.quit)
	}
}

func (c *Ticker) GetEventChan(ctx context.Context) <-chan usecaseifs.IOSCMessage {
	return c.messages
}

func (c *Ticker) SendMessage(ctx context.Context, msg usecaseifs.IOSCMessage) error {
	return fmt.Errorf("ticker does not support sending messages")
}

func (c *Ticker) updateTime(ctx context.Context) {
	c.log.Debugf(ctx, "Updating time...")
	rfcDateTime := time.Now().Format(time.RFC3339)

	c.messages <- osc_message.NewMessage("/time/rfc3339", []usecaseifs.IOSCMessageArgument{osc_message.NewMessageArgument("string", rfcDateTime)})

	parts := strings.Split("2006,06,Jan,January,01,1,Mon,Monday,2,_2,02,__2,002,15,3,03,4,04,5,05,PM", ",")
	for _, part := range parts {
		formattedPart := time.Now().Format(part)
		c.messages <- osc_message.NewMessage("/time/parts/"+part, []usecaseifs.IOSCMessageArgument{osc_message.NewMessageArgument("string", formattedPart)})
	}
}
