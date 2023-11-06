package dummy_bridge

import (
	"context"
	"time"

	"net.kopias.oscbridge/app/drivers/osc_message"

	"net.kopias.oscbridge/app/adapters/config"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IOSCConnection = &Connection{}

// Connection acts as a dummy mixer console, that emits pre-configured recurring messages in groups, imitating changing scenarios, enabling testing of actions.
type Connection struct {
	messages     chan usecaseifs.IOSCMessage
	log          usecaseifs.ILogger
	mockSettings config.DummyConnection

	// A channel that shows when the client exited with an error.
	notify chan error
	quit   chan interface{}
	debug  bool
}

func NewConnection(log usecaseifs.ILogger, mockSettings config.DummyConnection, debug bool) *Connection {
	return &Connection{
		log:          log,
		mockSettings: mockSettings,
		debug:        debug,
		notify:       make(chan error, 1),
		quit:         make(chan interface{}, 1),
		messages:     make(chan usecaseifs.IOSCMessage, 1),
	}
}

func (c *Connection) Start(ctx context.Context) error {
	if len(c.mockSettings.MessageGroups) == 0 {
		c.log.Warn(ctx, "No message_groups were specified!")
		return nil
	}
	go c.clientDispatch(ctx)

	return nil
}

func (c *Connection) clientDispatch(ctx context.Context) {
	changeTicker := time.NewTicker(time.Duration(c.mockSettings.IterationSpeedSecs) * time.Second)
	defer changeTicker.Stop()
	round := 0

	time.Sleep(1 * time.Second)

	for {
		index := round % len(c.mockSettings.MessageGroups)
		currentGroup := c.mockSettings.MessageGroups[index]
		c.log.Infof(ctx, "Simulating console state: %s %s", currentGroup.Name, currentGroup.Comment)

		for _, command := range currentGroup.OSCCommands {
			args := []usecaseifs.IOSCMessageArgument{}
			for _, a := range command.Arguments {
				args = append(args, osc_message.NewMessageArgument(a.Type, a.Value))
			}
			msg := osc_message.NewMessage(command.Address, args)
			c.messages <- msg
		}

		round++

		select {
		case <-c.quit:
			c.log.Info(ctx, "Dummy implementation quiting...")
			return
		case <-changeTicker.C:
		}
	}
}

// Notify returns the notification channel that can be used to listen for the client's exit
func (c *Connection) Notify() <-chan error {
	return c.notify
}

func (c *Connection) Stop(ctx context.Context) {
	c.quit <- true
}

func (c *Connection) GetEventChan(ctx context.Context) <-chan usecaseifs.IOSCMessage {
	return c.messages
}

func (c *Connection) SendMessage(ctx context.Context, msg usecaseifs.IOSCMessage) error {
	c.log.Infof(ctx, "Sending message: %v", msg)
	c.messages <- msg
	return nil
}
