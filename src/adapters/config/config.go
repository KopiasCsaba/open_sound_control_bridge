// Package config loads, parses, verifies and enables the retrieval of the configuration.
package config

import "net.kopias.oscbridge/app/usecase/usecaseifs"

var _ usecaseifs.IConfiguration = &MainConfig{}

type (
	// MainConfig represents the config YAML structure.
	MainConfig struct {
		OSCSources     OSCSource       `yaml:"osc_sources"`
		OBSConnections []OBSConnection `yaml:"obs_connections" `
		App            `yaml:"app"`
		Actions        map[string]Action `yaml:"actions"`
	}

	// OSCSource is the tree for all the different sources where OSC Messages can be received.
	OSCSource struct {
		ConsoleBridges   []ConsoleBridge   `yaml:"console_bridges"`
		DummyConnections []DummyConnection `yaml:"dummy_connections"`
		OBSBridges       []OBSBridge       `yaml:"obs_bridges"`
		HTTPBridges      []HTTPBridge      `yaml:"http_bridges"`
		Tickers          []Ticker          `yaml:"tickers"`
	}

	// ConsoleBridge connects to an OSC source device, e.g. to a mixer console and receives messages, executes subscription commands.
	ConsoleBridge struct {
		Name              string                `yaml:"name"`
		Prefix            string                `yaml:"prefix"`
		Enabled           bool                  `yaml:"enabled"`
		OSCImplementation string                `yaml:"osc_implementation"`
		Port              int64                 `yaml:"port"`
		Host              string                `yaml:"host"`
		Subscriptions     []ConsoleSubscription `yaml:"subscriptions"`
		InitCommand       *OSCCommand           `yaml:"init_command"`
		CheckAddress      string                `yaml:"check_address"`
		CheckPattern      string                `yaml:"check_pattern"`
	}

	// ConsoleSubscription contains messages to be repeated at certain intervals to subscribe events on a mixer console.
	ConsoleSubscription struct {
		OSCCommand   OSCCommand `yaml:"osc_command"`
		RepeatMillis int64      `yaml:"repeat_millis"`
	}

	// DummyConnection generates pre-defined set of messages to trigger actions. Useful for testing without the actual mixer console.
	DummyConnection struct {
		Name               string                     `yaml:"name"`
		Prefix             string                     `yaml:"prefix"`
		Enabled            bool                       `yaml:"enabled"`
		IterationSpeedSecs int64                      `yaml:"iteration_speed_secs"`
		MessageGroups      []DummyConsoleMessageGroup `yaml:"message_groups"`
	}

	// DummyConsoleMessageGroup is a set of messages to be emitted at the same time.
	DummyConsoleMessageGroup struct {
		Name        string       `yaml:"name"`
		Comment     string       `yaml:"comment"`
		OSCCommands []OSCCommand `yaml:"osc_commands"`
	}

	// An OBSBridge is an OSCSource, that uses an OBSConnection by its name to receive/poll status and convert it to OSCMessages.
	OBSBridge struct {
		Name       string `yaml:"name"`
		Prefix     string `yaml:"prefix"`
		Enabled    bool   `yaml:"enabled"`
		Connection string `yaml:"connection"`
	}

	// A HTTPBridge is an OSCSource, that listens on a port for requests and converts them to OSCMessages.
	HTTPBridge struct {
		Name    string `yaml:"name"`
		Prefix  string `yaml:"prefix"`
		Enabled bool   `yaml:"enabled"`
		Port    int64  `yaml:"port"`
		Host    string `yaml:"host"`
	}

	// A Ticker is an OSCSource, that emits OSCMessages containing the time.
	Ticker struct {
		Name              string `yaml:"name"`
		Prefix            string `yaml:"prefix"`
		Enabled           bool   `yaml:"enabled"`
		RefreshRateMillis int64  `yaml:"refresh_rate_millis"`
	}

	// OSCCommand represent an OSC message
	OSCCommand struct {
		Address   string        `yaml:"address"`
		Comment   string        `yaml:"comment"`
		Arguments []OSCArgument `yaml:"arguments"`
	}

	// OSCArgument is a single argument for an OSC message.
	OSCArgument struct {
		Type  string `yaml:"type"`
		Value string `yaml:"value"`
	}

	// OBSConnection represetnts a single connection to an OBS instance through websocket.
	OBSConnection struct {
		Name     string `yaml:"name"`
		Port     int64  `yaml:"port"`
		Host     string `yaml:"host"`
		Password string `yaml:"password"`
	}

	// App contains general app settings.
	App struct {
		Debug            Debug  `yaml:"debug"`
		StorePersistPath string `yaml:"store_persist_path"`
	}

	Debug struct {
		DebugOSCConnection bool `yaml:"debug_osc_connection"`
		DebugOSCConditions bool `yaml:"debug_osc_conditions"`
		DebugTasks         bool `yaml:"debug_tasks"`
		DebugOBSRemote     bool `yaml:"debug_obs_remote"`
	}

	// Action contains a set of conditions that may trigger a set of tasks. E.g. If Channel 1 is muted, then do an HTTP request.
	Action struct {
		DebounceMillis int64                  `yaml:"debounce_millis"`
		TriggerChain   ActionConditionChecker `yaml:"trigger_chain"`
		Tasks          []ActionTask           `yaml:"tasks"`
	}

	ActionTask struct {
		Type       string                 `yaml:"type"`
		Parameters map[string]interface{} `yaml:"parameters"`
	}

	ActionConditionChecker struct {
		Type       string                   `yaml:"type"`
		Parameters map[string]interface{}   `yaml:"parameters"`
		Children   []ActionConditionChecker `yaml:"children"`
	}
)

func (c *MainConfig) ShouldDebugOSCConditions() bool {
	return c.App.Debug.DebugOSCConditions
}
