# pamixermidicontrol

pamixermidicontrol is a tool for controlling your PulseAudio setup from a MIDI controller.

It is based on https://github.com/solarnz/pamidicontrol/, with the following differences:
- no longer based on D-Bus as the drop-in replacement for PulseAudio server, pipewire-pulse, does not have D-Bus support (https://gitlab.freedesktop.org/pipewire/pipewire/-/issues/1127)
- SysEx capabilities to simplify configuration with automatic retrieval of controller setup (Korg nanoKontrol2, Akai LPD8 in "PAD" mode)
- handle multiple controllers simultaneously
- handle MIDI note and program change in addition to control change
- configuration checking

This has been tested on Arch Linux with these controllers:
- [KORG nanoKontrol2](https://www.korg.com/us/products/computergear/nanokontrol2/)
- [Akai LPD8](https://www.akaipro.com/lpd8.html)
- [M-Audio Keystation Mini 32](https://www.m-audio.com/keystation-mini32-mk3)

## Installation

Pre-requisites:
- go 1.22
- PulseAudio (no need for module-dbus-protocol)
- portmidi library
    - Arch linux: `pacman -S portmidi`
    - Debian-based linux: `apt-get install libportmidi-dev`

```
go get github.com/fluciotto/pamixermidicontrol
make
```

## Configuration

pamixermidicontrol requires the use of a configuration file.

Place the config file under `$HOME/.config/pamixermidicontrol/config.yaml`.

### Examples

Some configuration examples are available: [example configuration files](https://github.com/fluciotto/pamixermidicontrol/tree/master/config-examples).

### Configuration format

```
midiDevices:

  - name: <MIDI device custom name, must be unique accross midiDevices>
    type: <"Generic" | "KorgNanoKontrol2" | "AkaiLpd8">
    # pamixermidicontrol --list-midi
    midiInName: <MIDI device IN port name>
    midiOutName: <MIDI device OUT port name>
  - ...

rules:

  - midiMessage:
      deviceName: <MIDI device custom name>

      # Only if the device is "KorgNanoKontrol2" or "AkaiLpd8"
      deviceControlPath: <
        Korg nanoKontrol2:
          [
            Group[1-8]/[Slider|Knob|Solo|Mute|Record] |
            Transport/Track/[Prev|Next] |
            Transport/Cycle |
            Transport/Marker/[Set|Prev|Next] |
            Transport/[Rewind|FastForward|Stop|Play|Rec]
          ]
        Akai LPD8:
          [ Pad[1-8] | Knob[1-8] ]
      >

      # Optional if the device type is "KorgNanoKontrol2" or "AkaiLpd8"
      # Mandatory if the device is "Generic"
      type: <"Note" | "ControlChange" | "ProgramChange">
      channel: <0-15>
      # if type is "Note"
      note: <0-127>
      # else if type is "ControlChange"
      controller: <0-127>
      # else if type is "ProgramChange"
      program: <0-127>
      # only if type is "ControlChange"
      minValue: <0-127, optional, default 0>
      maxValue: <0-127, optional, default 127>

    actions:
      - type: <"SetVolume" | "ToggleMute" | "SetDefaultOutput">
        # if type is "SetVolume" or "ToggleMute"
        target:
          type: <"Output" | "Input" | "PlaybackStream" | "RecordStream">
          # pamixermidicontrol --list-pulse
          name: <PulseAudio output, input, playback stream or record stream name, can be "Default" if type is "Input" or "Output">
        # else if type is "SetDefaultOutput"
        target:
          # pamixermidicontrol --list-pulse
          name: <PulseAudio output name>
      - ...

  - ...
```

How to get available MIDI ports names?

run `pamixermidicontrol --list-midi`

How to get available PulseAudio objects names?

run `pamixermidicontrol --list-pulse`

pamixermidicontrol will print to stderr all of the midi control messages it gets, so you can easily build up your configuration file iteratively.
