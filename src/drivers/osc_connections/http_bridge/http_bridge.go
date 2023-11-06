package http_bridge

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"net.kopias.oscbridge/app/drivers/osc_message"

	"net.kopias.oscbridge/app/pkg/chantools"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

// This implementation uses loffa/gosc , which is lacking disconnect options.

var _ usecaseifs.IOSCConnection = &HTTPBridge{}

type Config struct {
	Debug bool
	Host  string
	Port  int64
}

// HTTPBridge starts an HTTP server receiving GET requests and converts them to osc messages.
type HTTPBridge struct {
	log      usecaseifs.ILogger
	messages chan usecaseifs.IOSCMessage
	// A channel that shows when the client exited with an error.
	quit   chan any
	cfg    Config
	notify chan error
	mux    *http.ServeMux
	srv    *http.Server
}

func NewHTTPBridge(log usecaseifs.ILogger, cfg Config) usecaseifs.IOSCConnection {
	return &HTTPBridge{
		log:      log,
		cfg:      cfg,
		quit:     make(chan any),
		messages: make(chan usecaseifs.IOSCMessage, 1),
		notify:   make(chan error, 1),
	}
}

func (c *HTTPBridge) Start(ctx context.Context) error {
	c.mux = http.NewServeMux()
	c.mux.HandleFunc("/", c.getRoot)

	go c.run(ctx)
	return nil
}

func (c *HTTPBridge) run(ctx context.Context) {
	c.log.Infof(ctx, "Starting server at port %s:%d", c.cfg.Host, c.cfg.Port)
	c.srv = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port),
		Handler:      c.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	if err := c.srv.ListenAndServe(); err != nil {
		c.notify <- err
		c.Stop(ctx)
	}
}

// getRoot is the handler for '/', converts GET requests to OSC messages.
// The 'address' query parameter will be the address.
// The 'args[]' array will become a list of arguments.
//
//	Each value must be in the format: 'type,value'
func (c *HTTPBridge) getRoot(w http.ResponseWriter, r *http.Request) {
	// Parse the query parameters
	queryParams := r.URL.Query()

	// Get the value of the 'address' parameter
	address := queryParams.Get("address")

	if address == "" {
		err := fmt.Errorf("invalid address: '%s'", address)
		c.badRequest(r.Context(), w, err)
		return
	}
	// Get the values of the 'args' parameter as an array
	args := queryParams["args[]"]

	// Process the values
	// response := fmt.Sprintf("Address: %s\nArgs: %s", address, strings.Join(args, ", "))

	oscMsgArgs := []usecaseifs.IOSCMessageArgument{}
	for i, arg := range args {
		before, after, found := strings.Cut(arg, ",")
		if !found {
			err := fmt.Errorf("invalid request: call argument[%d] does not contain a comma", i)
			c.badRequest(r.Context(), w, err)
			return
		}
		oscMsgArgs = append(oscMsgArgs, osc_message.NewMessageArgument(before, after))
	}
	oscMsg := osc_message.NewMessage(address, oscMsgArgs)

	if _, err := io.WriteString(w, "OK"); err != nil {
		c.log.Err(r.Context(), fmt.Errorf("failed to respond to request: %w", err))
	}
	c.messages <- oscMsg
}

func (c *HTTPBridge) badRequest(ctx context.Context, w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)
	if _, err := io.WriteString(w, err.Error()); err != nil {
		c.log.Err(ctx, fmt.Errorf("failed to respond to request: %w", err))
	}
}

// Notify returns the notification channel that can be used to listen for the client's exit
func (c *HTTPBridge) Notify() <-chan error {
	return c.notify
}

func (c *HTTPBridge) Stop(ctx context.Context) {
	_ = c.srv.Shutdown(context.Background())

	if chantools.ChanIsOpenReader(c.quit) {
		close(c.quit)
	}
}

func (c *HTTPBridge) GetEventChan(ctx context.Context) <-chan usecaseifs.IOSCMessage {
	return c.messages
}

func (c *HTTPBridge) SendMessage(ctx context.Context, msg usecaseifs.IOSCMessage) error {
	return fmt.Errorf("HTTPBridges does not support sending messages")
}
