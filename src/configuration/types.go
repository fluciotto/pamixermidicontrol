package configuration

import "gopkg.in/yaml.v3"

// MIDI Device

type MidiDeviceType string

const (
	Generic          MidiDeviceType = "Generic"
	AkaiLpd8         MidiDeviceType = "AkaiLpd8"
	KorgNanoKontrol2 MidiDeviceType = "KorgNanoKontrol2"
)

type MidiDevice struct {
	Name        string         `yaml:"name"`
	Type        MidiDeviceType `yaml:"type"`
	MidiInName  string         `yaml:"midiInName"`
	MidiOutName string         `yaml:"midiOutName"`
}

// Rule

type MidiMessageType string

const (
	None          MidiMessageType = ""
	Note          MidiMessageType = "Note"
	ControlChange MidiMessageType = "ControlChange"
	ProgramChange MidiMessageType = "ProgramChange"
)

type MidiMessage struct {
	DeviceName        string          `yaml:"deviceName"`
	DeviceControlPath string          `yaml:"deviceControlPath"`
	Type              MidiMessageType `yaml:"type"`
	Channel           uint8           `yaml:"channel"`
	Note              uint8           `yaml:"note"`
	Controller        uint8           `yaml:"controller"`
	Program           uint8           `yaml:"program"`
	MinValue          uint8           `yaml:"minValue"`
	MaxValue          uint8           `yaml:"maxValue"`
}

type PulseAudioActionType string

const (
	SetVolume        PulseAudioActionType = "SetVolume"
	ToggleMute       PulseAudioActionType = "ToggleMute"
	SetDefaultOutput PulseAudioActionType = "SetDefaultOutput"
)

type PulseAudioTargetType string

const (
	PlaybackStream PulseAudioTargetType = "PlaybackStream"
	RecordStream   PulseAudioTargetType = "RecordStream"
	OutputDevice   PulseAudioTargetType = "OutputDevice"
	InputDevice    PulseAudioTargetType = "InputDevice"
)

type Target struct {
	Name string `yaml:"name"`
}

type TypedTarget struct {
	Type PulseAudioTargetType `yaml:"type"`
	Name string               `yaml:"name"`
}

type Action struct {
	Type      PulseAudioActionType `yaml:"type"`
	RawTarget yaml.Node            `yaml:"target"`
	Target    interface{}          `yaml:"-"`
}

type Rule struct {
	MidiMessage MidiMessage `yaml:"midiMessage"`
	Actions     []Action    `yaml:"actions"`
}

// Configuration

type Config struct {
	MidiDevices []MidiDevice `yaml:"midiDevices"`
	Rules       []Rule       `yaml:"rules"`
}
