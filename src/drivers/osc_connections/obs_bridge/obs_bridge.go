package obs_bridge

import (
	"context"
	"fmt"

	"github.com/andreykaipov/goobs/api/events"
	"net.kopias.oscbridge/app/drivers/osc_message"

	"net.kopias.oscbridge/app/drivers/obsremote"
	"net.kopias.oscbridge/app/pkg/chantools"
	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IOSCConnection = &OBSBridge{}

type Config struct {
	Debug      bool
	Connection *obsremote.OBSRemote
}

// OBSBridge uses a named OBS Connection, subscribes for certain events from OBS and emits an OSC Message when an update is received.
type OBSBridge struct {
	log      usecaseifs.ILogger
	messages chan usecaseifs.IOSCMessage
	// A channel that shows when the client exited with an error.
	quit   chan any
	cfg    Config
	notify chan error
}

func NewOBSBridge(log usecaseifs.ILogger, cfg Config) usecaseifs.IOSCConnection {
	return &OBSBridge{
		log:      log,
		cfg:      cfg,
		quit:     make(chan any),
		messages: make(chan usecaseifs.IOSCMessage, 10),
		notify:   make(chan error, 1),
	}
}

func (c *OBSBridge) Start(ctx context.Context) error {
	if err := c.initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	go c.listen(ctx)
	return nil
}

// Notify returns the notification channel that can be used to listen for the client's exit
func (c *OBSBridge) Notify() <-chan error {
	return c.notify
}

func (c *OBSBridge) Stop(ctx context.Context) {
	if chantools.ChanIsOpenReader(c.quit) {
		close(c.quit)
	}
}

func (c *OBSBridge) GetEventChan(ctx context.Context) <-chan usecaseifs.IOSCMessage {
	return c.messages
}

func (c *OBSBridge) SendMessage(ctx context.Context, msg usecaseifs.IOSCMessage) error {
	return fmt.Errorf("OBSBridge does not support sending messages")
}

// initialize retrieves initial values for the subscribed events.
func (c *OBSBridge) initialize(ctx context.Context) error {
	programScene, err := c.cfg.Connection.GetCurrentProgramScene(ctx)
	if err != nil {
		return err
	}
	c.handleObsEvent(ctx, &events.CurrentProgramSceneChanged{SceneName: programScene})

	previewScene, err := c.cfg.Connection.GetCurrentPreviewScene(ctx)
	if err != nil {
		return err
	}
	c.handleObsEvent(ctx, &events.CurrentPreviewSceneChanged{SceneName: previewScene})

	isStreaming, err := c.cfg.Connection.IsStreaming(ctx)
	if err != nil {
		return err
	}
	c.handleObsEvent(ctx, &events.StreamStateChanged{OutputActive: isStreaming})

	isRecording, err := c.cfg.Connection.IsRecording(ctx)
	if err != nil {
		return err
	}
	c.handleObsEvent(ctx, &events.RecordStateChanged{OutputActive: isRecording})

	return nil
}

// listen watches for incoming messages from OBS.
func (c *OBSBridge) listen(ctx context.Context) {
	eventChan, err := c.cfg.Connection.GetIncomingEvents()
	if err != nil {
		c.notify <- fmt.Errorf("failed to open incoming events channel for obs: %w", err)
	}
	for {
		select {
		case e := <-eventChan:
			c.handleObsEvent(ctx, e)
		case <-c.quit:
			return
		}
	}
}

// handleObsEvent filters the incoming events and converts the appropriate ones to OSC Messages.
func (c *OBSBridge) handleObsEvent(ctx context.Context, event interface{}) {
	// c.log.Debugf(ctx, "INCOMING OBS EVENT: %#v", event)

	switch t := event.(type) {
	case *events.CurrentPreviewSceneChanged:
		args := []usecaseifs.IOSCMessageArgument{osc_message.NewMessageArgument("string", t.SceneName)}
		c.messages <- osc_message.NewMessage("/obs/preview_scene", args)

	case *events.CurrentProgramSceneChanged:
		args := []usecaseifs.IOSCMessageArgument{osc_message.NewMessageArgument("string", t.SceneName)}
		c.messages <- osc_message.NewMessage("/obs/program_scene", args)

	case *events.RecordStateChanged:
		active := "0"
		if t.OutputActive {
			active = "1"
		}
		args := []usecaseifs.IOSCMessageArgument{osc_message.NewMessageArgument("int32", active)}
		c.messages <- osc_message.NewMessage("/obs/recording", args)

	case *events.StreamStateChanged:
		active := "0"
		if t.OutputActive {
			active = "1"
		}
		args := []usecaseifs.IOSCMessageArgument{osc_message.NewMessageArgument("int32", active)}
		c.messages <- osc_message.NewMessage("/obs/streaming", args)

	default:
		// c.log.Debugf(ctx, "UNHANDLED INCOMING OBS EVENT: %#v", t)
	}
}
