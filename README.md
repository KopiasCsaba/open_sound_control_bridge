<h1>Table of contents</h1>



<!-- TOC -->
* [Open Sound Control Bridge](#open-sound-control-bridge)
  * [Example uses](#example-uses)
* [Install](#install)
* [Configuration](#configuration)
  * [Example configuration](#example-configuration)
* [Sources](#sources)
  * [Digital Mixing Consoles](#digital-mixing-consoles)
<!-- TOC -->

# Open Sound Control Bridge

OSCBridge is a tool to help automate operations with audio/streaming gear.

If you want to automate things* based on:
 * Digital Audio Mixer Console state (such as Behringer X32 or other that supports OSC)
 * OBS Studio state
 * A HTTP Request
 * Time

What are those things?

OSCBridge is easily extendable, but currently it supports the following "tasks":
 * HTTP Request
 * Delay (just wait)
 * OBS Change preview scene
 * OBS Change program scene
 * OBS Send "vendor" message to any plugin that cares, e.g. to the amazingly excellent [Advanced Scene Switcher](https://github.com/WarmUpTill/SceneSwitcher).
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
 * When ... send a command to Advanced Scene Switcher, to do a [zillion other things](https://github.com/WarmUpTill/SceneSwitcher/wiki)
 * When ... then make Advanced Scene Switcher do an http request to execute some other actions through the oscbridge. (Btw A.S.S. can send OSC messages too.)

I think now you got the point!

# Install

This section is under construction.

# Overview

From a birds eye view, oscbridge provides a central "message store", to which "osc sources" can publish messages.
Every time a new message arrives, each action is checked, if their trigger_chain conditions are resolving to true based on the current store.
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

# Sources
## Digital Mixing Consoles
Many digital mixing consoles support a protocol called "[Open Sound Control](https://en.wikipedia.org/wiki/Open_Sound_Control)",
this is a UDP based simple protocol. It is based on "Messages", where each message has an address, and 0 or more arguments, and each argument can be a string, a float, an int, etc.

I have tested on Behringer X32 (See pmalliot's excellent work [here](https://sites.google.com/site/patrickmaillot/x32) on OSC).

In the case of X32, we need to regularly(8-10 sec) issue a /subscribe command with proper arguments, to show that we are 
interested in updates of a certain value from the console. Then the mixer is flooding us with the requested parameter.

So below is a real world example for behringer x32 OSC connection:

<details>
<summary>Click to see YAML</summary>

```yaml
  console_bridges:
    - name: "behringer_x32"             # The name of this mixer
      enabled: true                     # If enabled, OSCBRIDGE will try to connect, and restart if fails.
      prefix: ""                        # Prefix determines
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
```

</details>

<details>
<summary>Click to see YAML</summary>

```yaml

```

</details>

<details>
<summary>Click to see YAML</summary>

```yaml

```

</details>

<details>
<summary>Click to see YAML</summary>

```yaml

```

</details>

<details>
<summary>Click to see YAML</summary>

```yaml

```

</details>

<details>
<summary>Click to see YAML</summary>

```yaml

```

</details>

<details>
<summary>Click to see YAML</summary>

```yaml

```

</details>

<details>
<summary>Click to see YAML</summary>

```yaml

```

</details>

<details>
<summary>Click to see YAML</summary>

```yaml

```

</details>

<details>
<summary>Click to see YAML</summary>

```yaml

```

</details>
