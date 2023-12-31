<h1>Table of contents</h1>


<!-- TOC -->
* [Open Sound Control Bridge](#open-sound-control-bridge)
  * [Example uses](#example-uses)
* [Install](#install)
  * [Docker](#docker)
* [Overview](#overview)
* [Configuration](#configuration)
  * [Example configuration](#example-configuration)
  * [Actions](#actions)
    * [Debouncing](#debouncing)
  * [Trigger chain](#trigger-chain)
    * [Conditions](#conditions)
      * [OSC_MATCH: Check if a single message exists](#oscmatch-check-if-a-single-message-exists)
        * [Trigger on change](#trigger-on-change)
      * [AND: Require all children condition to resolve to true](#and-require-all-children-condition-to-resolve-to-true)
      * [OR: Require at least one children to resolve to true](#or-require-at-least-one-children-to-resolve-to-true)
      * [NOT: Negate the single child's result.](#not-negate-the-single-childs-result)
  * [Sources](#sources)
    * [Digital Mixing Consoles](#digital-mixing-consoles)
    * [Dummy console](#dummy-console)
    * [OBS bridges](#obs-bridges)
    * [HTTP bridges](#http-bridges)
    * [Tickers](#tickers)
  * [Tasks](#tasks)
    * [HTTP request](#http-request)
    * [OBS Scene change](#obs-scene-change)
    * [OBS Vendor message](#obs-vendor-message)
    * [Delay](#delay)
    * [Run command](#run-command)
    * [Send OSC message](#send-osc-message)
* [Development](#development)
<!-- TOC -->

# Open Sound Control Bridge

OSCBridge is a tool to help automate operations with audio/streaming gear.

Input could come from various sources, such as:

* Digital Audio Mixer Console state (such as Behringer X32 or other that supports OSC)
* OBS Studio state
* A HTTP Request
* Time

OSCBridge currently supports the following "tasks":

* HTTP Request
* Delay (just wait)
* OBS Change preview scene
* OBS Change program scene
* OBS Send "vendor" message to any plugin that cares, e.g. to the amazingly
  excellent [Advanced Scene Switcher](https://github.com/WarmUpTill/SceneSwitcher).
* Excecute a command
* Send an OSC message

## Example uses

Here is just a few idea:

* When a microphone is unmuted, turn the PTZ camera to the speaker.
* When the stage is unmuted, turn the PTZ camera to the stage.
* When a special HTTP request arrives, mute/unmute something.
* When a special HTTP request arrives, set the volume of a channel to the specified value.
* At a specified time, unmute a microphone.
* At a specified time, switch to an OBS Scene.
* At a specified time, send an HTTP Request.
* When something is unmuted, switch to a scene in OBS.
* When a scene is activated in OBS unmute certain channels.
* When a microphone is unmuted, then turn the camera but only if a ceratin OBS scene is active.
* When ... send a command to Advanced Scene Switcher, to do
  a [zillion other things](https://github.com/WarmUpTill/SceneSwitcher/wiki)
* When ... then make Advanced Scene Switcher do an http request to execute some other actions through the oscbridge. (
  Btw A.S.S. can send OSC messages too.)

I think now you got the point!

# Install
 * Download the binary from the latest [release](https://github.com/KopiasCsaba/open_sound_control_bridge/releases)
 * Create a [config.yml](https://github.com/KopiasCsaba/open_sound_control_bridge#example-configuration) next to the binary.
 * Execute!

```bash
$:oscbridge$ ls
config.yml  oscbridge-6acaf3b4-linux-amd64.bin

$:oscbridge$ chmod +x oscbridge-6acaf3b4-linux-amd64.bin 

$:oscbridge$ ./oscbridge-6acaf3b4-linux-amd64.bin 
2023-11-13 07:30:51 [ INFO] OPEN SOUND CONTROL BRIDGE is starting.
2023-11-13 07:30:51 [ INFO] Version: v1.0.0 Revision: 6acaf3b4 
2023-11-13 07:30:51 [ INFO] Initializing OBS connections...
2023-11-13 07:30:51 [ INFO]     Connecting to streaming_pc_obs...
2023-11-13 07:30:51 [ INFO] Initializing OBS bridges...
2023-11-13 07:30:51 [ INFO] Initializing Open Sound Control (mixer consoles, etc) connections...

...
```

You may override the config.yml location with the environment variable `APP_CONFIG_FILE`, e.g.: `APP_CONFIG_FILE=/a/b/c/d/osc.yml`.


## Docker
A docker-hub version will be coming soon when my time permits.

# Overview

From a birds eye view, oscbridge provides a central "message store", to which "osc sources" can publish messages.
Every time a new message arrives, each action is checked, if their trigger_chain conditions are resolving to true based
on the current store.
If every the trigger chain resolves to true, then the action's tasks are executed.

So this is the control flow:
[OSC SOURCES] -> [OSC MESSAGE STORE] -> [ACTION TRIGGER CHAIN] -> [ACTION TASK]

# Configuration

## Example configuration

Below is the simplest example to showcase how the system works.
<details>
<summary>Click to see YAML</summary>

```yaml
obs_connections:
  - name: "streaming_pc_obs"
    host: 192.168.1.75
    port: 4455
    password: "foobar"

osc_sources:
  console_bridges:
    - name: "behringer_x32"
      enabled: false
      prefix: ""
      host: 192.168.2.99
      port: 10023
      osc_implementation: l
      init_command:
        address: /xinfo
      check_address: /ch/01/mix/on
      check_pattern: "^0|1$"
      subscriptions:
        - osc_command:
            address: /subscribe
            arguments:
              - type: string
                value: /ch/01/mix/on
              - type: int32
                value: 10
          repeat_millis: 8000

  dummy_connections:
    - name: "behringer_x32_dummy"
      enabled: true
      prefix: ""
      iteration_speed_secs: 1
      message_groups:
        - name: mic_1_on
          osc_commands:
            - address: /ch/01/mix/on
              comment: "headset mute (0: muted, 1: unmuted)"
              arguments:
                - type: int32
                  value: 1
        - name: mic_1_off
          osc_commands:
            - address: /ch/01/mix/on
              comment: "headset mute (0: muted, 1: unmuted)"
              arguments:
                - type: int32
                  value: 0

actions:
  to_pulpit:
    trigger_chain:
      type: osc_match
      parameters:
        address: /ch/01/mix/on
        arguments:
          - index: 0
            type: "int32"
            value: "1"
    tasks:
      - type: http_request
        parameters:
          url: "http://127.0.0.1:8888/cgi-bin/ptzctrl.cgi?ptzcmd&poscall&0&__TURN_TO_PULPIT"
          method: "get"
          timeout_secs: 1
      - type: obs_scene_change
        parameters:
          scene: "PULPIT"
          scene_match_type: regexp
          target: "program"
          connection: "streaming_pc_obs"
      - type: obs_scene_change
        parameters:
          scene: "STAGE"
          scene_match_type: regexp
          target: "preview"
          connection: "streaming_pc_obs"

  to_stage:
    trigger_chain:
      type: osc_match
      parameters:
        address: /ch/01/mix/on
        arguments:
          - index: 0
            type: "int32"
            value: "0"
    tasks:
      - type: http_request
        parameters:
          url: "http://127.0.0.1:8888/cgi-bin/ptzctrl.cgi?ptzcmd&poscall&1&__TURN_TO_STAGE"
          method: "get"
          timeout_secs: 1
      - type: obs_scene_change
        parameters:
          scene: "STAGE"
          scene_match_type: regexp
          target: "program"
          connection: "streaming_pc_obs"
      - type: obs_scene_change
        parameters:
          scene: "PULPIT"
          scene_match_type: regexp
          target: "preview"
          connection: "streaming_pc_obs"
```

</details>

In this configuration there are two OSC sources:

* Dummy (enabled)
* A Behringer X32 digital console (disabled)

The dummy source acts as if someone would press Ch1's mute button every second to toggle it.

Then there are two actions defined, "to_pulpit" and "to_stage".
Each has a single trigger, that matches /ch/01/mix/on to be 0 or 1.

Then for each action, there are three tasks:

* An HTTP request that would recall a PTZ Optics camera preset (0 and 1 respectively).
* An obs_scene_change to change the program scene.
* An obs_scene_change to change the preview scene.

You can see the results on this gif:

<a href="docs/assets/readme/example_config.mkv"><img src="docs/assets/readme/example_config.gif" width=300></a>

OBS is switching scenes based on the mute status, and at the bottom you can see the arriving requests.

You can just switch from the dummy to the console one, and your mute button is then tied to OBS scenes and the camera.

## Actions

Actions encapsulate a so called `trigger_chain` and a list of `tasks` together.

This is how actions look like:

```yaml
actions:
  change_to_pulpit:
    trigger_chain:
    # ... tree of conditions
    tasks:
    # ... 1 dimensional list of tasks to be executed in order, serially

  change_to_stage:
    trigger_chain:
    # ... tree of conditions
    tasks:
    # ... 1 dimensional list of tasks to be executed in order, serially

  start_live_stream:
    trigger_chain:
    # ... tree of conditions
    tasks:
    # ... 1 dimensional list of tasks to be executed in order, serially
```

Each action has it's own name, that is shown in the logs upon evaluation/execution.

Whenever the internal store receives an update, OSCBridge checks each action's trigger_chain, the tree of conditions if
they match the store or not.
If the trigger_chain is evaluated to be true, then the tasks will be executed.

### Debouncing

There is an option, that can be specified for each action, called `debounce_millis`,
if provided then the logic changes a bit. Upon store change, if the trigger_chain resolves to true,
then after the specified ammount of milliseconds the trigger_chain is re-evaluated.
If it is still true, only then will the tasks be executed.

For example:

```yaml
actions:
  change_to_pulpit:
    trigger_chain:
    # ... tree of conditions
    tasks:
    # ... 1 dimensional list of tasks to be executed in order, serially
    debounce_millis: 500
```

This could protect against quick transients, e.g. an accidental unmute/mute. For example here,
if the trigger chain is watching for ch1's unmute, then it will only execute the tasks if it is unmute for more than
0.5seconds.
This can help avoid accidents, where you accidentally unmute something but then you immediately mute it back.

## Trigger chain

The trigger chain is a tree of conditions. Some conditions can be nested, some of them are just leafs on a tree, without
any children.

You can build very complex conditions into here, e.g. (in pseudo code):

```
IF
(mic1-is-muted AND mic2-is-unmuted) OR 
(ch10-is-unmuted AND 
    (
    ch11fader > 0.5 OR 
    ch12fader > 0.5
    )
) THEN
...

```

But the way to express these are a bit more complicated due to the YAML configuration we use.

### Conditions

#### OSC_MATCH: Check if a single message exists

The `osc_match` condition can nothave any children, and it is checking for a single message in the store.
It can check based on address, address regexp and also based on arguments.

Here is an example:

```yaml
actions:
  change_to_pulpit:
    trigger_chain:
      - type: osc_match
        parameters:
          address: /ch/01/mix/on
          arguments:
            - index: 0
              type: "int32"
              value: "1"
    tasks:
    # ...
```

This is a single condition on an action's trigger_chain.
This checks for a message with an exact address of  "/ch/01/mix/on" and with a single first argument, that is int32 and
the value is 1.

If such a message exists in the store, the tasks will be executed.

Parameters:

| Parameter          | Default value  | Possible values | Description                                                                   | Example values                   |
|--------------------|----------------|-----------------|-------------------------------------------------------------------------------|----------------------------------|
| address            | none, required |                 | The value for matching a message's address. Can be a regexp, see next option. | /ch/01/mix/on, /ch/0[0-9]/mix/on |
| address_match_type | `eq`           | `eq`, `regexp`  | Determines the way of address matching.                                       | `regexp`                         |
| trigger_on_change  | `true`         | `true`, `false` | See the [trigger on change](#trigger-on-change) paragraph.                    | `true`                           |
| arguments          | none, optional |                 | See the next table.                                                           | List of arguments                |

Arguments:

| Parameter        | Default value  | Possible values                  | Description                                                                     | Example values |
|------------------|----------------|----------------------------------|---------------------------------------------------------------------------------|----------------|
| index            | none, required | `0`                              | The 0 based index for the argument.                                             | `0`, `1`, `2`  |
| type             | none, required | `string`, `int32`, `float32`     | The type of the argument.                                                       | `string`       |
| value            | none, required |                                  | The value of the argument.                                                      | `1`            |
| value_match_type | `=`            | `regexp`, `<=`,`<`,`>`,`>=`,`!=` | The comparison method. In case of regexp, the value can be a regexp expression. | `=`            |

##### Trigger on change

The `trigger_on_change` option is a special one. Whenever a new message arrives that changes the store, every
trigger_chain is checked.

Now, during the execution of the trigger_chain, it is being monitored what messages those conditions accessed.
By default (when `trigger_on_change: true`) if the trigger chain did not access the NEWLY UPDATED message, so the one
that just arrived,
the tasks aren't going to be executed. This avoids unneccessary re-execution just because an unrelevant message updated
the store.

But this is also usable, to avoid re-execution in a case when a relevant message updated the store.

Practically this option decouples a condition from being a trigger. The condition is still required to match in order to
execute the tasks, but that single condition's change will not trigger execution.

You want to set this to false, when you don't want to re-execute the action upon the toggling of one of the parameters
your trigger_chain is watching for. This is an edge case, that comes handy sometimes.

For example, let's say you have the following trigger_chain (in pseudo-ish code):

```
IF ( OBS-scene-name-contains-foobar AND
     OR (ch1-unmuted OR ch2-unmuted OR stage-is-muted) 
    )
THEN
...
```

So you want to only execute the tasks, when certain things on the console match, but don't wanna re-execute just because
of an OBS scene change.
But you only want to execute the tasks, when certain things on the console match AND obs scene name contains foobar.

Then you can mark the OBS-scene-name-contains condition with `trigger_on_change: false`.
That will cause the tasks to be executed when the console state changes (and obs scene contains foobar), but will not
trigger if only obs changes would otherwise match.
E.g. you might switch from one scene to another that contains foobar in our pseudo example, but that would not
re-execute the tasks.

#### AND: Require all children condition to resolve to true

`And` as it's name implies requires all children to resolve to true.

The following example action requires both ch1 **AND** ch2 to be on.

```yaml
actions:
  change_to_pulpit:
    trigger_chain:
      type: and
      children:
        - type: osc_match
          parameters:
            address: /ch/01/mix/on
            arguments:
              - index: 0
                type: "int32"
                value: "1"
        - type: osc_match
          parameters:
            address: /ch/02/mix/on
            arguments:
              - index: 0
                type: "int32"
                value: "1"
  tasks:
  # ...
```

Now you see how conditions can be nested.

#### OR: Require at least one children to resolve to true

`Or` as it's name implies requires that at least one of the childrens would resolve to true.

The following example action executes the tasks if ch1 **OR** ch2 is be on.

```yaml
actions:
  change_to_pulpit:
    trigger_chain:
      type: or
      children:
        - type: osc_match
          parameters:
            address: /ch/01/mix/on
            arguments:
              - index: 0
                type: "int32"
                value: "1"
        - type: osc_match
          parameters:
            address: /ch/02/mix/on
            arguments:
              - index: 0
                type: "int32"
                value: "1"
  tasks:
  # ...
```

#### NOT: Negate the single child's result.

The `NOT` condition simply negates it's single child's result.

Here is how you would achieve this pseudo code:

```
AND(ch1-unmuted; NOT(OR(ch10-unmuted,ch20-unmuted)))
```

In yaml:

```yaml
actions:
  change_to_pulpit:
    trigger_chain:
      type: and
      children:
        - type: osc_match
          parameters:
            address: /ch/01/mix/on
            arguments:
              - index: 0
                type: "int32"
                value: "1"
        - type: not
          children:
            - type: or
              children:
                - type: osc_match
                  parameters:
                    address: /ch/10/mix/on
                    arguments:
                      - index: 0
                        type: "int32"
                        value: "1"
                - type: osc_match
                  parameters:
                    address: /ch/20/mix/on
                    arguments:
                      - index: 0
                        type: "int32"
                        value: "1"

  tasks:
  # ...
```

## Sources

Now that you know how to compose conditions, you need input sources, that would add messages to the internal store,
against which you can match your trigger chains.

### Digital Mixing Consoles

Many digital mixing consoles support a protocol
called "[Open Sound Control](https://en.wikipedia.org/wiki/Open_Sound_Control)",
this is a UDP based simple protocol. It is based on "Messages", where each message has an address, and 0 or more
arguments, and each argument can be a string, a float, an int, etc.

I have tested on Behringer X32, so most examples are based on this console.
See pmalliot's excellent work [here](https://sites.google.com/site/patrickmaillot/x32) on
X32's [OSC](https://drive.google.com/file/d/1Snbwx3m6us6L1qeP1_pD6s8hbJpIpD0a/view) implementation.

In the case of X32, we need to regularly(8-10 sec) issue a /subscribe command with proper arguments, to show that we are
interested in updates of a certain value from the console. Then the mixer is flooding us with the requested parameter.

So below is a real world example for behringer x32 OSC connection:

<details>
<summary>Click to see YAML</summary>

```yaml
osc_sources:
  console_bridges:
    # The name of this mixer
    - name: "behringer_x32"

      # If enabled, OSCBRIDGE will try to connect, and restart if fails.
      enabled: true

      # Prefix determines the message address prefix as it will be stored to the store.
      # E.g. if you'd have multiple consoles, you could prefix them "/console1", "/console2",
      # and you could match for /console1/ch/01/mix/on for example.
      prefix: ""

      host: 192.168.2.99
      port: 10023

      # The driver to use. We only have "l" for now.
      osc_implementation: l

      # This command is sent right after the connection is opened.
      # It can be used for authentication, or anything that is required.
      # X32 does not require anything, but for this it returns it's own name.
      init_command:
        address: /xinfo
        # You could specify arguments also.
        # arguments:
        #   - type: string
        #     value: "foobar"

      # There is a regular query running, for checking if the connection is still alive.
      # Specify an address here, and a regexp that matches the returned value.
      # If there is no response, or the response doesn't match, the connection is counted as broken and the app restarts.
      check_address: /ch/01/mix/on
      check_pattern: "^0|1$"
      # Subscriptions are commands that are sent regularly (repeat_millis) that cause the mixer to update us with the lates values for the subscribed thing.
      # Research your own mixer for the exact syntax, but this is how you do it for X32.
      subscriptions:
        - osc_command:
            # This command subscribes for channel 1's mute status. 0 is muted, 1 is unmuted.
            address: /subscribe
            arguments:
              - type: string
                value: /ch/01/mix/on
              - type: int32
                value: 10
          repeat_millis: 8000
```

</details>

### Dummy console

The dummy console implementation is just what it's name implies.
It has `message_groups`, and each `message_group` contains `messages`.
The dummy console iterates infinitely through the groups, and executes the messages in them.
Between each group it waits the configured ammount of time.

The below example configures two groups, called "mic_1_on" and "mic_1_off".

Therefore, it provides a way to test the logic even without a real connection to a mixer.
You can have a dummy emitting the same messages the real console would, and you can freely enable/disable any source,
so you can test, or you can switch to the real operation mode by enabling the console connection.

<details>
<summary>Click to see YAML</summary>

```yaml
osc_sources:
  dummy_connections:
    - name: "behringer_x32_dummy"
      # Use this source, or not.
      enabled: true

      # Prefix determines the message address prefix as it will be stored to the store.
      prefix: ""

      # How much delay should be between each group?
      iteration_speed_secs: 1

      # Message groups are set of messages being emitted at once.
      message_groups:
        - name: mic_1_on
          osc_commands:
            - address: /ch/01/mix/on
              comment: "headset mute (0: muted, 1: unmuted)"
              arguments:
                - type: int32
                  value: 1

        - name: mic_1_off
          osc_commands:
            - address: /ch/01/mix/on
              comment: "headset mute (0: muted, 1: unmuted)"
              arguments:
                - type: int32
                  value: 0

```

</details>

### OBS bridges

OSCBridge can be configured to connect to an OBS Studio instance via websocket, and it will subscribe to some events in
OBS.

These events are the following:

* CurrentPreviewSceneChanged
    * Message:
        * Address: /obs/preview_scene
        * Argument[0]: string, value: NAME_OF_SCENE
* CurrentProgramSceneChanged
    * Message:
        * Address: /obs/program_scene
        * Argument[0]: string, value: NAME_OF_SCENE
* RecordStateChanged
    * Message:
        * Address: /obs/recording
        * Argument[0]: int32, value: 0 or 1
* StreamStateChanged
    * Message:
        * Address: /obs/streaming
        * Argument[0]: int32, value: 0 or 1

In order to configure an OBS Bridge, you'll also need to configure an OBS Connection.

<details>
<summary>Click to see YAML</summary>

```yaml

obs_connections:
  - name: "streampc_obs"
    host: 192.168.1.75
    port: 4455
    password: "foobar12345"

osc_sources:
  obs_bridges:
    - name: "obsbridge1"
      # You may choose to disable it.
      enabled: true

      # Prefix determines the message address prefix as it will be stored to the store.
      prefix: ""

      # The name of the obs connection, see above.
      connection: "streampc_obs"

```

</details>

### HTTP bridges

HTTP Bridges in OSCBridge enables you to open a port on a network interface and start a HTTP server on them.
The server can receive special HTTP GET requests, and converts them to OSC messages and stores them in the message
store.
Then you can write actions that check for that value, and may even execute tasks based on it.

The message can be put away under some namespace by using the prefix option, but you could also use it to override an
existing message.

To insert an OSC Message like this:

```
Message(address: /foo/bar/baz, arguments: [Argument(string:hello), Argument(int32:1)])
```

Execute a GET request like this:

```bash
curl "127.0.0.1:7878/?address=/foo/bar/baz&args[]=string,hello&args[]=int32,1"
```

<details>
<summary>Click to see YAML</summary>

```yaml
osc_sources:
  http_bridges:
    - name: "httpbridge1"
      # You may choose to disable it.
      enabled: true
      # Prefix determines the message address prefix as it will be stored to the store.
      prefix: ""
      port: 7878
      host: 0.0.0.0
```

</details>

### Tickers

You can enable "Tickers", that would regularly update the store with messages representing the current date/time.

The ticker publishes several packages under "/time/" (if you don't specify a prefix), with names that might be weird for
the first time,
if you are not familiar with how golang's time formatting works.

You may see the full reference [here](https://cs.opensource.google/go/go/+/refs/tags/go1.21.3:src/time/format.go;l=9).

Currently these messages are being emitted in every iteration:

```
Message(address: /time/rfc3339,         arguments: [Argument(string:2023-11-07T08:53:06Z)])
Message(address: /time/parts/2006,      arguments: [Argument(string:2023)])
Message(address: /time/parts/06,        arguments: [Argument(string:23)])
Message(address: /time/parts/Jan,       arguments: [Argument(string:Nov)])
Message(address: /time/parts/January,   arguments: [Argument(string:November)])
Message(address: /time/parts/01,        arguments: [Argument(string:11)])
Message(address: /time/parts/1,         arguments: [Argument(string:11)])
Message(address: /time/parts/Mon,       arguments: [Argument(string:Tue)])
Message(address: /time/parts/Monday,    arguments: [Argument(string:Tuesday)])
Message(address: /time/parts/2,         arguments: [Argument(string:7)])
Message(address: /time/parts/_2,        arguments: [Argument(string: 7)])
Message(address: /time/parts/02,        arguments: [Argument(string:07)])
Message(address: /time/parts/__2,       arguments: [Argument(string:311)])
Message(address: /time/parts/002,       arguments: [Argument(string:311)])
Message(address: /time/parts/15,        arguments: [Argument(string:08)])
Message(address: /time/parts/3,         arguments: [Argument(string:8)])
Message(address: /time/parts/03,        arguments: [Argument(string:08)])
Message(address: /time/parts/4,         arguments: [Argument(string:53)])
Message(address: /time/parts/04,        arguments: [Argument(string:53)])
Message(address: /time/parts/5,         arguments: [Argument(string:6)])
Message(address: /time/parts/05,        arguments: [Argument(string:06)])
Message(address: /time/parts/PM,        arguments: [Argument(string:AM)])
```

So if you want to match for hour:minute, then you want to match the values of /time/15 and /time/04 respectively in the
trigger chain (to be explained later).

<details>
<summary>Click to see YAML</summary>

```yaml
osc_sources:
  tickers:
    - name: "ticker1"
      # You may choose to disable it.
      enabled: true

      # Prefix determines the message address prefix as it will be stored to the store.
      prefix: ""

      # How often updates should occur
      refresh_rate_millis: 1000
```

</details>

## Tasks

Now you have actions, trigger_chains and sources, the final piece is to have tasks that will be executed if the
trigger_chain evaluates to true.

### HTTP request

The `http_request` task executes a specific http request upon evaluation.

Parameters:

| Parameter    | Default value  | Possible values | Description                   | Example values                                           |
|--------------|----------------|-----------------|-------------------------------|----------------------------------------------------------|
| url          | none, required |                 | The URL for the request.      | http://127.0.0.1/?foo=bar                                |
| body         | empty string   |                 | The request body.             | {"json":"or something else"}                             |
| timeout_secs | 30             |                 | The timeout for the request.  | 1                                                        |
| method       | `GET`          | `GET`, `POST`   | The method for the request.   | `POST`                                                   |
| headers      | empty          |                 | A list of "Key: value" pairs. | <pre>- "Content-Type: text/json"<br>- "X-Foo: bar"</pre> |

Example:

```yaml
actions:
  to_pulpit:
    trigger_chain:
    # ...
    tasks:
      - type: http_request
        parameters:
          url: "http://127.0.0.1:8888/cgi-bin/ptzctrl.cgi?ptzcmd&poscall&0&__TURN_TO_PULPIT"
          method: "get"
          timeout_secs: 1
          headers:
            - "X-Foo: bar"
            - "X-Foo2: baz"
          body: "O HAI"
```

The request that will be made:

```
GET /cgi-bin/ptzctrl.cgi?ptzcmd&poscall&0&__TURN_TO_PULPIT HTTP/1.1
Host: 127.0.0.1:8888
User-Agent: Go-http-client/1.1
Content-Length: 5
X-Foo: bar
X-Foo2: baz
Accept-Encoding: gzip

O HAI
```

### OBS Scene change

The `obs_scene_change` task changes the live or program scene on a remote OBS instance.

Parameters:

| Parameter        | Default value  | Possible values      | Description                                               | Example values    |
|------------------|----------------|----------------------|-----------------------------------------------------------|-------------------|
| scene            | none, required |                      | The name of the scene to which we need to switch.         | `PULPIT`, `STAGE` |
| connection       | none, required |                      | The name of the obs connection that this task should use. | `streampc_obs`    |
| scene_match_type | `exact`        | `exact`, `regexp`    | How to match the scene name.                              | `regexp`          |
| target           | none, required | `program`, `preview` | Which side of OBS should be switched.                     | `program`         |

Example:

```yaml
obs_connections:
  - name: "streampc_obs"
    host: 192.168.1.75
    port: 4455
    password: "foobar12345"


actions:
  to_pulpit:
    trigger_chain:
    # ...
    tasks:
      - type: obs_scene_change
        parameters:
          scene: "PULPIT.*"
          scene_match_type: regexp
          target: "program"
          connection: "streaming_pc_obs"

      - type: obs_scene_change
        parameters:
          scene: "STAGE"
          scene_match_type: exact
          target: "preview"
          connection: "streaming_pc_obs"
```

### OBS Vendor message

It is possible to send
a [VendorEvent](https://github.com/obsproject/obs-websocket/blob/master/docs/generated/protocol.md#vendorevent) to OBS
via a websocket connection.
Different plugins can listen for these events, one example is the
marvelous [Advanced Scene Switcher](https://github.com/WarmUpTill/SceneSwitcher/),
which [supports](https://github.com/WarmUpTill/SceneSwitcher/wiki/Websockets#websocket-condition) this.

So given that you are listening in that plugin for "IF Websocket Message waas received: foobar_notice",
you can execute macros remotely with OSCBridge:

```yaml
obs_connections:
  - name: "streampc_obs"
    host: 192.168.1.75
    port: 4455
    password: "foobar12345"

actions:
  to_pulpit:
    trigger_chain:
    # ...
    tasks:
      - type: obs_vendor_request
        parameters:
          connection: "streampc_obs"
          vendorName: "AdvancedSceneSwitcher"
          requestType: "AdvancedSceneSwitcherMessage"
          requestData:
            message: "foobar_notice"
```

Parameters:

| Parameter   | Default value  | Description                                               | Example values                 |
|-------------|----------------|-----------------------------------------------------------|--------------------------------|
| connection  | none, required | The name of the obs connection that this task should use. | `streampc_obs`                 |
| vendorName  | none, required |                                                           | `AdvancedSceneSwitcher`        |
| requestType | none, required |                                                           | `AdvancedSceneSwitcherMessage` |
| requestData | none, required |                                                           | `message: whatever`            |

### Delay

The `delay` simply delays the serial execution of the tasks, taking up as much time as you configure.

Parameters:

| Parameter    | Default value  | Description                    | Example values          |
|--------------|----------------|--------------------------------|-------------------------|
| delay_millis | none, required | How much milliseconds to wait. | `1500` (for 1.5 second) |

Example:

```yaml
actions:
  to_pulpit:
    trigger_chain:
    # ...
    tasks:
      - type: delay
        parameters:
          delay_millis: 1500
```

### Run command

The `run_command` task simply executes the given command.

| Parameter         | Default value  | Description                                                                         | Example values                                          |
|-------------------|----------------|-------------------------------------------------------------------------------------|---------------------------------------------------------|
| command           | none, required | The path to the binary to execute.                                                  | /usr/bin/bash                                           |
| arguments         | optional       | The list of arguments.                                                              | <pre>- "-l"<br>- "-c"<br>- "date > /tmp/date.txt"</pre> |
| run_in_background | false          | Whether or not the serial execution of tasks should wait for the command to finish. |                                                         |
| directory         | optional       | The execution folder for the command.                                               |                                                         |

You need to [follow](https://pkg.go.dev/os/exec#example-Command) the classical way of specifying a binary and it's
arguments.
So you can not use `date > /tmp/date.txt` as the command, you need to specify `/usr/bin/bash` as the command, and then
the parameters.

Example:

```yaml
actions:
  to_pulpit:
    trigger_chain:
    # ...
    tasks:
      - type: run_command
        parameters:
          command: "/usr/bin/bash"
          arguments: [ "-l","-c","date > /tmp/date.txt" ]
```

### Send OSC message

The `send_osc_message` sends an open sound control message through the specified connection.
Currently only the `console_bridges` support sending a message. E.g. you can send a message back to your console.

| Parameter  | Default value  | Description                                                               | Example values                         |
|------------|----------------|---------------------------------------------------------------------------|----------------------------------------|
| connection | none, required | The OSC connection to use (the `name` from one of your `console_bridges`) | `behringer_x32`                        |
| address    | none, required | The address of the message.                                               | `/ch/10/mix/on`                        |
| arguments  | optional       | The arguments of the message.                                             | <pre>- type: int32<br>- value: 0</pre> |

Example:

(Unmute channel 10)

```yaml
actions:
  to_pulpit:
    trigger_chain:
    # ...
    tasks:
      - type: send_osc_message
        parameters:
          connection: "behringer_x32"
          address: "/ch/10/mix/on"
          arguments:
            - type: int32
              value: 1
```

# Development

You'll need "make" and "docker" installed.
After cloning the repository, run "make" to see the available commands. 

Run `make dev_start` to start the development environment. 

It'll look for a config.yml in the source root.