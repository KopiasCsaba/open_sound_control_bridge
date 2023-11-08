package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net.kopias.oscbridge/app/drivers/osc_connections/obs_bridge"

	"net.kopias.oscbridge/app/drivers/osc_conditions"

	"net.kopias.oscbridge/app/drivers/osc_connections/dummy_bridge"

	osc_ticker "net.kopias.oscbridge/app/drivers/osc_connections/ticker"

	"net.kopias.oscbridge/app/adapters/config"
	"net.kopias.oscbridge/app/drivers/actioncomposer"
	"net.kopias.oscbridge/app/drivers/messagestore"
	"net.kopias.oscbridge/app/drivers/obsremote"
	"net.kopias.oscbridge/app/drivers/osc_conditions/cond_and"
	"net.kopias.oscbridge/app/drivers/osc_conditions/cond_not"
	"net.kopias.oscbridge/app/drivers/osc_conditions/cond_or"
	"net.kopias.oscbridge/app/drivers/osc_conditions/cond_osc_msg_match"
	"net.kopias.oscbridge/app/drivers/osc_connections/console_bridge_l"
	"net.kopias.oscbridge/app/drivers/osc_connections/http_bridge"
	"net.kopias.oscbridge/app/drivers/osc_message"
	"net.kopias.oscbridge/app/drivers/tasks/delay"
	"net.kopias.oscbridge/app/drivers/tasks/httpreq"
	"net.kopias.oscbridge/app/drivers/tasks/obstasks"
	"net.kopias.oscbridge/app/drivers/tasks/run_command"
	"net.kopias.oscbridge/app/drivers/tasks/send_osc_message"
	"net.kopias.oscbridge/app/entities"
	"net.kopias.oscbridge/app/pkg/logger"
	"net.kopias.oscbridge/app/usecase"
	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

const Version = "v1.0.0"

var (

	// Revision is the git revision of the current executable, filled in with the build command.
	Revision string
	// BuildTime is the time of build of the current executable, filled in with the build command.
	BuildTime string
)

func main() {
	for {
		ctx := context.Background()

		log := logger.New()
		log.Infof(ctx, "OPEN SOUND CONTROL BRIDGE is starting.")
		log.Infof(ctx, "Version: %s Revision: %.8s Built at: %s", Version, Revision, BuildTime)

		err := startApp(ctx, log)
		if err != nil {
			log.Err(ctx, fmt.Errorf("%w: the application will restart now", err))
		} else {
			os.Exit(0)
		}
		// nolint:forbidigo
		fmt.Println("\n\n\n\n\n\n ")
		time.Sleep(2 * time.Second)
	}
}

// nolint:cyclop,gocognit,gocyclo
func startApp(ctx context.Context, log *logger.Logger) error {
	log.AddPrefixerFunc(usecase.GetContextualLogPrefixer)

	// == Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("mainConfig error: %w", err)
	}

	// == OBS Connections
	log.Infof(ctx, "Initializing OBS connections...")
	obsConnections := map[string]*obsremote.OBSRemote{}
	defer stopObsConnections(ctx, obsConnections)

	for _, c := range cfg.OBSConnections {
		log.Infof(ctx, "\tConnecting to %s...", c.Name)
		obsRemoteCfg := obsremote.Config{
			Host:     c.Host,
			Port:     c.Port,
			Password: c.Password,
			Debug:    cfg.App.Debug.DebugOBSRemote,
		}
		obsRemote := obsremote.NewOBSRemote(log, obsRemoteCfg)

		if err = obsRemote.Start(ctx); err != nil {
			return err
		}
		obsConnections[c.Name] = obsRemote
	}

	oscConnections := []entities.OscConnectionDetails{}
	defer stopOscConnections(ctx, oscConnections)

	// OBS Brdiges
	log.Infof(ctx, "Initializing OBS bridges...")
	for _, c := range cfg.OSCSources.OBSBridges {
		if !c.Enabled {
			continue
		}

		var oscConn usecaseifs.IOSCConnection

		log.Infof(ctx, "\tStarting obs bridge %s...", c.Name)
		conn, ok := obsConnections[c.Connection]
		if !ok {
			return fmt.Errorf("failed to start obs bridge: there is no connection named '%s'", c.Connection)
		}
		obsCfg := obs_bridge.Config{
			Debug:      cfg.App.Debug.DebugOSCConnection,
			Connection: conn,
		}

		oscConn = obs_bridge.NewOBSBridge(log, obsCfg)
		if err = oscConn.Start(ctx); err != nil {
			return fmt.Errorf("failed to start obs bridge: %w", err)
		}

		oscConnections = append(oscConnections, *entities.NewOscConnectionDetails(c.Name, c.Prefix, oscConn))
	}

	// == Console Bridges
	log.Infof(ctx, "Initializing Open Sound Control (mixer consoles, etc) connections...")
	for _, c := range cfg.OSCSources.ConsoleBridges {
		if !c.Enabled {
			continue
		}
		var oscConn usecaseifs.IOSCConnection

		log.Infof(ctx, "\tConnecting to %s...", c.Name)
		switch c.OSCImplementation {
		case "l":
			oscConnCfg := console_bridge_l.Config{
				Debug:         cfg.App.Debug.DebugOSCConnection,
				Subscriptions: c.Subscriptions,
				Port:          c.Port,
				Host:          c.Host,
				CheckAddress:  c.CheckAddress,
				CheckPattern:  c.CheckPattern,
			}
			oscConn = console_bridge_l.NewConnection(log, oscConnCfg)
			if err := oscConn.Start(ctx); err != nil {
				return fmt.Errorf("failed to start osc[l] connection: %w", err)
			}
		// case "s":
		// 	oscConnCfg := console_bridge_s.Config{
		// 		Debug:         cfg.App.Debug.DebugOSCConnection,
		// 		Subscriptions: c.Subscriptions,
		// 		Port:          c.Port,
		// 		Host:          c.Host,
		// 		CheckAddress:  c.CheckAddress,
		// 		CheckPattern:  c.CheckPattern,
		// 	}
		// 	oscConn = console_bridge_s.NewConnection(log, oscConnCfg)
		// 	if err := oscConn.Start(ctx); err != nil {
		// 		return fmt.Errorf("failed to start osc[s] connection: %w", err)
		// 	}
		default:
			return fmt.Errorf("unknown osc implementation: %s", c.OSCImplementation)
		}

		if c.InitCommand != nil {
			args := []usecaseifs.IOSCMessageArgument{}
			for _, a := range c.InitCommand.Arguments {
				args = append(args, osc_message.NewMessageArgument(a.Type, a.Value))
			}

			if err := oscConn.SendMessage(ctx, osc_message.NewMessage(c.InitCommand.Address, args)); err != nil {
				return fmt.Errorf("failed to query mixer: %w", err)
			}
		}

		oscConnections = append(oscConnections, *entities.NewOscConnectionDetails(c.Name, c.Prefix, oscConn))
	}

	// == Dummy Console Connections
	log.Infof(ctx, "Initializing Dummy Open Sound Control connections...")

	for _, c := range cfg.OSCSources.DummyConnections {
		if !c.Enabled {
			continue
		}

		var oscConn usecaseifs.IOSCConnection

		log.Infof(ctx, "\tConnecting to %s...", c.Name)

		oscConn = dummy_bridge.NewConnection(log, c, cfg.App.Debug.DebugOSCConnection)
		if err := oscConn.Start(ctx); err != nil {
			return fmt.Errorf("failed to start dummy osc connection: %w", err)
		}

		oscConnections = append(oscConnections, *entities.NewOscConnectionDetails(c.Name, c.Prefix, oscConn))
	}

	// == Tickers
	log.Infof(ctx, "Initializing tickers...")
	for _, c := range cfg.OSCSources.Tickers {
		if !c.Enabled {
			continue
		}

		var oscConn usecaseifs.IOSCConnection

		log.Infof(ctx, "\tStarting ticker %s...", c.Name)
		tickerCfg := osc_ticker.Config{
			Debug:             cfg.App.Debug.DebugOSCConnection,
			RefreshRateMillis: c.RefreshRateMillis,
		}

		oscConn = osc_ticker.NewTicker(log, tickerCfg)
		if err := oscConn.Start(ctx); err != nil {
			return fmt.Errorf("failed to start ticker: %w", err)
		}

		oscConnections = append(oscConnections, *entities.NewOscConnectionDetails(c.Name, c.Prefix, oscConn))
	}

	// == HTTP Bridges
	log.Infof(ctx, "Initializing http bridges...")
	for _, c := range cfg.OSCSources.HTTPBridges {
		if !c.Enabled {
			continue
		}

		var oscConn usecaseifs.IOSCConnection

		log.Infof(ctx, "\tStarting http bridge %s...", c.Name)
		hbCfg := http_bridge.Config{
			Debug: cfg.App.Debug.DebugOSCConnection,
			Host:  c.Host,
			Port:  c.Port,
		}

		oscConn = http_bridge.NewHTTPBridge(log, hbCfg)
		if err := oscConn.Start(ctx); err != nil {
			return fmt.Errorf("failed to start http2osc bridge: %w", err)
		}

		oscConnections = append(oscConnections, *entities.NewOscConnectionDetails(c.Name, c.Prefix, oscConn))
	}

	// == OSC Connection map
	oscConnectionMap := map[string]usecaseifs.IOSCConnection{}
	for _, c := range oscConnections {
		oscConnectionMap[c.Name] = c.Connection
	}

	// == Tasks
	log.Infof(ctx, "Initializing Tasks ...")

	registeredTasks := map[string]usecaseifs.ActionTaskFactory{
		"obs_scene_change":   obstasks.NewSceneChangerFactory(obsConnections, log, cfg.App.Debug.DebugTasks),
		"obs_vendor_request": obstasks.NewVendorRequestFactory(obsConnections, log, cfg.App.Debug.DebugTasks),
		"delay":              delay.NewFactory(log, cfg.App.Debug.DebugTasks),
		"http_request":       httpreq.NewFactory(log, cfg.App.Debug.DebugTasks),
		"send_osc_message":   send_osc_message.NewFactory(log, cfg.App.Debug.DebugTasks, oscConnectionMap),
		"run_command":        run_command.NewFactory(log, cfg.App.Debug.DebugTasks),
	}

	// == Conditions
	log.Infof(ctx, "Initializing Conditions ...")

	conditionTracker := osc_conditions.NewConditionTracker(log, cfg.App.Debug.DebugOSCConditions)

	registeredConditions := map[string]usecaseifs.ActionConditionFactory{
		"and":       cond_and.NewFactory(conditionTracker),
		"or":        cond_or.NewFactory(conditionTracker),
		"not":       cond_not.NewFactory(conditionTracker),
		"osc_match": cond_osc_msg_match.NewFactory(conditionTracker),
	}

	// == Composing actions
	actionComposer := actioncomposer.NewActionComposer(cfg.Actions, registeredConditions, registeredTasks)
	actions, err := actionComposer.GetActionList()
	if err != nil {
		return err
	}

	messageStore := messagestore.NewMessageStore()

	// == Compose use cases
	log.Infof(ctx, "Initializing Use cases...")
	ucs := usecase.New(
		log,
		cfg,
		oscConnections,
		messageStore,
		actions,
		cfg.StorePersistPath,
	)

	if err := ucs.Start(ctx); err != nil {
		return err
	}
	defer ucs.Stop(ctx)

	// == CTRL-C trap
	interrupt, trapStop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer trapStop()

	// == Serve & Watch for errors
	log.Info(ctx, "All services are up & running!")

	defer log.Infof(ctx, "Main thread quited.")

	select {
	case <-interrupt.Done():
		log.Info(ctx, "Kill signal received.")
		return nil
	case err := <-oscConnNotify(ctx, oscConnections):
		return fmt.Errorf("OSC bridge encountered an issue: %w", err)

	case err := <-obsConnNotify(ctx, obsConnections):
		return fmt.Errorf("OBS remote encountered an issue: %w", err)

	case err := <-ucs.Notify():
		return fmt.Errorf("USESCASES encountered an issue: %w", err)
	}
}

func stopObsConnections(ctx context.Context, connections map[string]*obsremote.OBSRemote) {
	for _, c := range connections {
		c.Stop(ctx)
	}
}

// obsConnNotify checks on all obs connections, and forwards the error if any of them emits an error.
func obsConnNotify(ctx context.Context, connections map[string]*obsremote.OBSRemote) chan error {
	e := make(chan error, 1)

	for _, c := range connections {
		c := c
		go func() {
			e <- <-c.Notify()
		}()
	}

	return e
}

func stopOscConnections(ctx context.Context, connectionDetails []entities.OscConnectionDetails) {
	for _, cd := range connectionDetails {
		cd.Connection.Stop(ctx)
	}
}

func oscConnNotify(ctx context.Context, connectionDetails []entities.OscConnectionDetails) chan error {
	e := make(chan error, 1)

	for _, cd := range connectionDetails {
		cd := cd
		go func() {
			e <- <-cd.Connection.Notify()
		}()
	}

	return e
}
