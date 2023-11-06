package console_bridge_l

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"net.kopias.oscbridge/app/drivers/osc_message"

	"net.kopias.oscbridge/app/pkg/chantools"

	"net.kopias.oscbridge/app/adapters/config"

	"github.com/loffa/gosc"
	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

// This implementation uses loffa/gosc, which is lacking disconnect options.

var _ usecaseifs.IOSCConnection = &Connection{}

type Config struct {
	Debug         bool
	Subscriptions []config.ConsoleSubscription
	Port          int64
	Host          string
	CheckAddress  string
	CheckPattern  string
}

type Connection struct {
	cfg      Config
	client   *gosc.Client
	messages chan usecaseifs.IOSCMessage
	log      usecaseifs.ILogger

	// Signals that the client is stopped.
	quit chan any

	// A channel that shows when the client exited with an error.
	notify chan error
}

func NewConnection(log usecaseifs.ILogger, cfg Config) usecaseifs.IOSCConnection {
	return &Connection{
		log:      log,
		cfg:      cfg,
		quit:     make(chan any),
		messages: make(chan usecaseifs.IOSCMessage),
		notify:   make(chan error, 1),
	}
}

func (c *Connection) Start(ctx context.Context) error {
	// Set up the client.
	client, err := gosc.NewClient(fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to resolve udp addr: %w", err)
	}

	c.client = client

	err = c.client.ReceiveMessageFunc(".*", func(oscMessage *gosc.Message) {
		msg, err2 := MessageFromOSCMessage(*oscMessage)
		if err2 != nil {
			c.log.Err(ctx, err2)
		}
		if c.cfg.Debug {
			c.log.Infof(ctx, "Received message: %v", msg)
		}
		c.messages <- msg
	})

	if err != nil {
		return fmt.Errorf("failed to register dispatcher: %w", err)
	}

	go c.watchdog(ctx)
	go c.manageSubscriptions(ctx)

	return nil
}

func (c *Connection) manageSubscriptions(ctx context.Context) {
	// Initial execution & ticker setup for repeating subscriptions as they time out.
	tickers := []*time.Ticker{}
	for i, sub := range c.cfg.Subscriptions {
		c.executeSub(ctx, i)
		t := time.NewTicker(time.Duration(sub.RepeatMillis) * time.Millisecond)
		tickers = append(tickers, t)
	}

	// Stop the tickers when returning...
	defer func() {
		for _, ticker := range tickers {
			ticker.Stop()
		}
	}()

	// Check on all the tickers and refresh subscriptions...
	for {
		for i, ticker := range tickers {
			select {
			case <-ticker.C:
				c.executeSub(ctx, i)
			case <-c.quit:
				return
			default:
			}
		}

		// Wait a bit and restart
		<-time.After(100 * time.Millisecond)
	}
}

// executeSub sends the configured OSC message to signal to the console that we are interested in certain updates.
func (c *Connection) executeSub(ctx context.Context, i int) {
	sub := c.cfg.Subscriptions[i]
	if c.cfg.Debug {
		c.log.Infof(ctx, "Subscribing to: %v (%s)", sub, sub.OSCCommand.Comment)
	}

	msg := &gosc.Message{
		Address:   sub.OSCCommand.Address,
		Arguments: []any{},
	}

	for _, a := range sub.OSCCommand.Arguments {
		var oscArgument usecaseifs.IOSCMessageArgument = osc_message.NewMessageArgument(a.Type, a.Value)

		oscMsgArg, err := OSCArgumentFromMessageArgument(oscArgument)
		if err != nil {
			c.log.Err(ctx, err)
			return
		}
		msg.Arguments = append(msg.Arguments, oscMsgArg)
	}

	err := c.client.SendMessage(msg)
	if err != nil {
		c.log.Err(ctx, fmt.Errorf("failed to subscribe/check subscription[%d] %v: %w", i, sub.OSCCommand, err))
	}
}

func (c *Connection) watchdog(ctx context.Context) {
	for {
		if c.cfg.Debug {
			c.log.Infof(ctx, "OSC conn checking connection...")

			// The checkConnection should finish in 10seconds, or we emit an error signalling that this connection is dead.
			stop := make(chan any, 1)

			go func() {
				select {
				case <-time.After(10 * time.Second):
					c.notify <- fmt.Errorf("timeout without check connection response")
				case <-stop:
					return
				}
			}()

			err := c.checkConnection(ctx)
			stop <- true
			if err != nil {
				c.notify <- fmt.Errorf("osc connection is broken: %w", err)
				c.Stop(ctx)
			}

			// Wait on stop to finish, or retry...
			select {
			case <-c.quit:
				return
			case <-time.After(5 * time.Second):
			}
		}
	}
}

// CheckConnection sends a check OSC Message for which some response is expected.
// The resulting response's first argument will be matched against a pattern.
func (c *Connection) checkConnection(ctx context.Context) error {
	msg, err := OSCMessageFromMessage(osc_message.NewMessage(c.cfg.CheckAddress, nil))
	if err != nil {
		return fmt.Errorf("failed to convert check message: %w", err)
	}
	resp, err := c.client.SendAndReceiveMessage(msg)
	if err != nil {
		return err
	}

	if len(resp.Arguments) == 0 {
		return fmt.Errorf("the received message has no arguments to check")
	}

	p, err := regexp.Compile(c.cfg.CheckPattern)
	if err != nil {
		return fmt.Errorf("failed to compile check pattern: %w", err)
	}

	if !p.MatchString(fmt.Sprintf("%v", resp.Arguments[0])) {
		return fmt.Errorf("failed to match '%s' against '%v'", c.cfg.CheckPattern, resp.Arguments[0])
	}

	return nil
}

// Notify returns the notification channel that can be used to listen for the client's exit
func (c *Connection) Notify() <-chan error {
	return c.notify
}

func (c *Connection) Stop(ctx context.Context) {
	if chantools.ChanIsOpenReader(c.quit) {
		close(c.quit)
	}
}

func (c *Connection) GetEventChan(ctx context.Context) <-chan usecaseifs.IOSCMessage {
	return c.messages
}

func (c *Connection) SendMessage(ctx context.Context, msg usecaseifs.IOSCMessage) error {
	if c.cfg.Debug {
		c.log.Infof(ctx, "Sending message %v", msg)
	}
	pkg, err := OSCMessageFromMessage(msg)
	if err != nil {
		return err
	}
	return c.client.SendMessage(pkg)
}
