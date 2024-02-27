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
    Arch linux: `pacman -S portmidi`
    Debian-based linux: `apt-get install libportmidi-dev`

```
go get github.com/fluciotto/pamixermidicontrol
```

## Configuration

pamixermidicontrol requires the use of a configuration file. Place the config file under `$HOME/.config/pamixermidicontrol/config.yaml`.
You can checkout the [example configuration file](https://github.com/fluciotto/pamixermidicontrol/blob/master/config.yaml) to see how to configure pamixermidicontrol.
You must set a bare-minimum the Input and Output midi device names.

pamixermidicontrol will print to stderr all of the midi control messages it gets, so you can easily build up your configuration file iteratively.

# Troubleshooting

## panic: runtime error: invalid memory address or nil pointer dereference on startup

Make sure that the names you have configured for `InputMidiName` / `OutputMidiName` actually exist.
