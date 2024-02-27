package midi

import (
	"github.com/fluciotto/pamixermidicontrol/src/configuration"
	akaiLpd8 "github.com/fluciotto/pamixermidicontrol/src/device/akai/lpd8"
	korgNanokontrol2 "github.com/fluciotto/pamixermidicontrol/src/device/korg/nanokontrol2"
	"github.com/fluciotto/pamixermidicontrol/src/pulseaudio"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/gomidi/midi/v2"

	driver "gitlab.com/gomidi/midi/v2/drivers/portmididrv"
)

func listDevices() ([]string, []string, error) {
	drv, err := driver.New()
	if err != nil {
		panic(err)
	}
	// make sure to close all open ports at the end
	defer drv.Close()
	// MIDI in
	ins, err := drv.Ins()
	if err != nil {
		return nil, nil, err
	}
	// MIDI out
	outs, err := drv.Outs()
	if err != nil {
		return nil, nil, err
	}
	// Get names
	inNames := make([]string, 0)
	outNames := make([]string, 0)
	for _, port := range ins {
		inNames = append(inNames, port.String())
	}
	for _, port := range outs {
		outNames = append(outNames, port.String())
	}
	return inNames, outNames, nil
}

func List() {
	log := log.Logger.With().Str("module", "Midi").Logger()
	ins, outs, err := listDevices()
	if err != nil {
		panic(err)
	}
	// List input ports
	for _, port := range ins {
		log.Info().Msgf("Found midi in device:\t%s", port)
	}
	// List output ports
	for _, port := range outs {
		log.Info().Msgf("Found midi out device:\t%s", port)
	}
}

type MidiClient struct {
	log        zerolog.Logger
	PAClient   *pulseaudio.PAClient
	MidiDevice configuration.MidiDevice
	Rules      []configuration.Rule
}

func NewMidiClient(paClient *pulseaudio.PAClient, device configuration.MidiDevice, rules []configuration.Rule) *MidiClient {
	client := &MidiClient{
		log:        log.With().Str("module", "Midi").Str("device", device.Name).Logger(),
		PAClient:   paClient,
		MidiDevice: device,
		Rules:      rules,
	}
	return client
}

func (client *MidiClient) Run() {
	drv, err := driver.New()
	if err != nil {
		panic(err)
	}

	// make sure to close all open ports at the end
	defer drv.Close()

	in, err := midi.FindInPort(client.MidiDevice.MidiInName)
	if err != nil {
		// panic(err)
		client.log.Error().Msgf("Could not find MIDI In %s", client.MidiDevice.MidiInName)
	}

	out, err := midi.FindOutPort(client.MidiDevice.MidiOutName)
	if err != nil {
		// panic(err)
		client.log.Error().Msgf("Could not find MIDI Out %s", client.MidiDevice.MidiInName)
	}

	if in == nil || out == nil {
		return
	}

	if err := in.Open(); err != nil {
		panic(err)
	}

	if err := out.Open(); err != nil {
		panic(err)
	}

	defer in.Close()
	defer out.Close()

	onMessage := func(sysExChannel chan []byte) func(msg midi.Message, timestampMs int32) {
		var doActions = func(rule configuration.Rule, value uint8) {
			for _, action := range rule.Actions {
				switch action.Type {
				case configuration.SetVolume:
					var minValue uint8
					var maxValue uint8
					if rule.MidiMessage.MinValue != 0 {
						minValue = rule.MidiMessage.MinValue
					} else {
						minValue = 0
					}
					if rule.MidiMessage.MaxValue != 0 {
						maxValue = rule.MidiMessage.MaxValue
					} else {
						maxValue = 0x7f
					}
					volumePercent := float32(value) / float32(maxValue-minValue)
					if err := client.PAClient.ProcessVolumeAction(action, volumePercent); err != nil {
						client.log.Error().Err(err)
					}
				case configuration.Mute:
					if value == 0 {
						return
					}
					if err := client.PAClient.ProcessToggleMute(action); err != nil {
						client.log.Error().Err(err)
					}
				case configuration.SetDefaultOutput:
					if value == 0 {
						return
					}
					if err := client.PAClient.SetDefaultOutput(action); err != nil {
						client.log.Error().Err(err)
					}
				default:
					client.log.Error().Msgf("Unknown action type %s in rule %+v", action.Type, rule)
				}
			}
		}
		return func(message midi.Message, timestampMs int32) {
			client.log.Debug().Msgf("Received MIDI message (%s) from in port %v", message.String(), in)
			switch message.Type() {
			case midi.NoteOnMsg, midi.NoteOffMsg:
				var channel uint8
				var note uint8
				var velocity uint8
				message.GetNoteOn(&channel, &note, &velocity)
				for _, rule := range client.Rules {
					if rule.MidiMessage.Type != configuration.Note {
						continue
					}
					if channel != rule.MidiMessage.Channel {
						continue
					}
					if note != rule.MidiMessage.Controller {
						continue
					}
					doActions(rule, velocity)
				}
			case midi.ControlChangeMsg:
				var channel uint8
				var controller uint8
				var ccValue uint8
				message.GetControlChange(&channel, &controller, &ccValue)
				for _, rule := range client.Rules {
					if rule.MidiMessage.Type != configuration.ControlChange {
						continue
					}
					if channel != rule.MidiMessage.Channel {
						continue
					}
					if controller != rule.MidiMessage.Controller {
						continue
					}
					doActions(rule, ccValue)
				}
			case midi.ProgramChangeMsg:
				var channel uint8
				var program uint8
				message.GetProgramChange(&channel, &program)
				for _, rule := range client.Rules {
					if rule.MidiMessage.Type != configuration.ProgramChange {
						continue
					}
					if channel != rule.MidiMessage.Channel {
						continue
					}
					doActions(rule, 0x7f)
				}
			case midi.SysExMsg:
				var bytes []byte
				message.GetSysEx(&bytes)
				sysExChannel <- bytes
			}
		}
	}

	sysExChannel := make(chan []byte)

	if _, err = midi.ListenTo(in, onMessage(sysExChannel), midi.UseSysEx()); err != nil {
		panic(err)
	}

	if client.MidiDevice.Type == configuration.AkaiLpd8 {
		device := akaiLpd8.New(client.MidiDevice.Name)
		// client.log.Info().Msgf("device %+v", device)
		// device.OnStart(sysExChannel, out)
		client.Rules = device.UpdateRules(client.Rules, sysExChannel, out)
	} else if client.MidiDevice.Type == configuration.KorgNanoKontrol2 {
		device := korgNanokontrol2.New(client.MidiDevice.Name)
		// client.log.Info().Msgf("device %+v", device)
		// device.OnStart(sysExChannel, out)
		client.Rules = device.UpdateRules(client.Rules, sysExChannel, out)
	}

	select {}
}
