// SysEx based on https://github.com/charlesfleche/lpd8editor/blob/master/doc/SYSEX.md

package akaiLpd8

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"

	"github.com/fluciotto/pamixermidicontrol/src/configuration"
	"github.com/fluciotto/pamixermidicontrol/src/device"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/sysex"
)

type AkaiLpd8 struct {
	log zerolog.Logger
	// DeviceInfo device.DeviceInfo
	DeviceName string
}

func New(name string) *AkaiLpd8 {
	return &AkaiLpd8{
		log: log.With().Str("device", "Akai LPD8").Logger(),
		// DeviceInfo: device.DeviceInfo{
		// 	Manufacturer: device.Manufacturer{ManufacturerID: sysex.Akai, Name: sysex.Akai.String()},
		// 	Model:        "LPD8",
		// },
		DeviceName: name,
	}
}

func (d *AkaiLpd8) identityMessage(channel byte) *device.SysExMessage {
	request := sysex.IdentityRequest(channel)
	responseHandler := func(bytes []byte) (rawData []byte, processedData []byte, err error) {
		log := d.log.With().Str("SysEx", "IdentityMessage").Logger()
		if len(bytes) != 35-2 {
			log.Error().Msgf("Identity response has a bad length: %d", len(bytes))
			return []byte{}, []byte{}, fmt.Errorf("identity response has a bad length: %d", len(bytes))
		}
		globalMidiChannel := bytes[1]
		manufacturerId := bytes[4]
		familyId := binary.LittleEndian.Uint16([]byte{bytes[5], bytes[6]})
		memberId := binary.LittleEndian.Uint16([]byte{bytes[7], bytes[8]})
		version := fmt.Sprintf("%d.%d.%d", bytes[9], bytes[10], bytes[11])
		log.Info().Msgf("Global MIDI channel 0x%X", globalMidiChannel)
		log.Info().Msgf("Manufacturer ID 0x%X", manufacturerId)
		log.Info().Msgf("Family ID 0x%X", familyId)
		log.Info().Msgf("Member ID 0x%X", memberId)
		log.Info().Msgf("Version %s", version)
		return bytes, nil, nil
	}
	return device.NewSysExMessage(request, responseHandler)
}

func (d *AkaiLpd8) activeProgramRequestMessage() *device.SysExMessage {
	request := []byte{
		0xf0,
		0x47,       // Akai
		0x7f, 0x75, // LPD8
		0x64, 0x00, 0x00, // Request active program
		0xf7,
	}
	responseHandler := func(bytes []byte) (rawData []byte, processedData []byte, err error) {
		log := d.log.With().Str("SysEx", "activeProgramRequestMessage").Logger()
		if len(bytes) != 9-2 {
			log.Error().Msgf("Mode response has a bad length: %d", len(bytes))
			return []byte{}, []byte{}, fmt.Errorf("mode response has a bad length: %d", len(bytes))
		}
		activeProgram := bytes[6]
		return bytes, []byte{activeProgram}, nil
	}
	return device.NewSysExMessage(request, responseHandler)
}

func (d *AkaiLpd8) programRequestMessage(programNumber byte) *device.SysExMessage {
	request := []byte{
		0xf0,
		0x47,       // Akai
		0x7f, 0x75, // LPD8
		0x63, 0x00, 0x01, programNumber, // Request program
		0xf7,
	}
	responseHandler := func(bytes []byte) (rawData []byte, processedData []byte, err error) {
		log := d.log.With().Str("SysEx", "programRequestMessage").Logger()
		if len(bytes) != 66-2 {
			log.Error().Msgf("Scene dump request response has a bad length: %d", len(bytes))
			return []byte{}, []byte{}, fmt.Errorf("scene dump request response has a bad length: %d", len(bytes))
		}

		programData := bytes[7:]
		return bytes, programData, nil
	}
	return device.NewSysExMessage(request, responseHandler)
}

func (d *AkaiLpd8) OnStart(c chan []byte, out drivers.Out) {
	// response, _, _ := d.identityMessage(0).Send(c, out, d.log)
	_, activeProgram, _ := d.activeProgramRequestMessage().Send(c, out, d.log)
	d.log.Info().Msgf("Active program % X", activeProgram)

	response, program, _ := d.programRequestMessage(activeProgram[0]).Send(c, out, d.log)
	d.log.Info().Msgf("Response % X, program % X", response, program)

	// _, sceneData, err := d.sceneDumpRequestMessage(0).Send(c, out, d.log)
	// if err != nil {
	// 	panic(err)
	// }
	// d.log.Info().Msgf("Processed scene data % X", sceneData)

	// LED mode to 0x1
	// sceneData[2] = 0x1
	// d.log.Info().Msgf("Updated scene data % X", sceneData)

	// response := d.sceneDumpMessage(0, dataToMidiData(scene)).Send(c, out, d.log)
	// d.log.Info().Msgf("Dump response % X", response)

	// response = d.sceneWriteMessage(0).Send(c, out, d.log)
	// d.log.Info().Msgf("Write response % X", response)
}

func (d *AkaiLpd8) UpdateRules(
	rules []configuration.Rule,
	c chan []byte,
	out drivers.Out,
) (updatedRules []configuration.Rule) {
	// Fetch program data from device
	_, activeProgram, err := d.activeProgramRequestMessage().Send(c, out, d.log)
	if err != nil {
		panic(err)
	}
	d.log.Debug().Msgf("Active program % X", activeProgram)

	_, programData, err := d.programRequestMessage(activeProgram[0]).Send(c, out, d.log)
	if err != nil {
		panic(err)
	}
	d.log.Debug().Msgf("Program % X", programData)

	// Get global MIDI channel from program data
	globalMidiChannel := programData[0]
	// Update rules with scene data
	for _, rule := range rules {
		if rule.MidiMessage.DeviceName != d.DeviceName {
			continue
		}
		if rule.MidiMessage.DeviceControlPath != "" {
			//
			padRe := regexp.MustCompile("^Pad([1-8])$")
			knobRe := regexp.MustCompile("^Knob([1-8])$")

			//
			if padRe.MatchString(rule.MidiMessage.DeviceControlPath) {
				matches := padRe.FindStringSubmatch(rule.MidiMessage.DeviceControlPath)
				padNumber64, _ := strconv.ParseUint(matches[1], 10, 8)
				padNumber := uint8(padNumber64)
				padIndex := 1 + (padNumber-1)*4
				// Update rule
				rule.MidiMessage.Channel = globalMidiChannel
				rule.MidiMessage.Type = configuration.Note
				rule.MidiMessage.Controller = programData[padIndex]
				// rule.MidiMessage.Type = configuration.ControlChange
				// rule.MidiMessage.Controller = programData[padIndex+2]
				// rule.MidiMessage.Type = configuration.ProgramChange
				// rule.MidiMessage.Controller = programData[padIndex+1]
				rule.MidiMessage.MinValue = 0x0
				rule.MidiMessage.MaxValue = 0x7f
				updatedRules = append(updatedRules, rule)
			} else if knobRe.MatchString(rule.MidiMessage.DeviceControlPath) {
				matches := knobRe.FindStringSubmatch(rule.MidiMessage.DeviceControlPath)
				knobNumber64, _ := strconv.ParseUint(matches[1], 10, 8)
				knobNumber := uint8(knobNumber64)
				knobIndex := 33 + (knobNumber-1)*3
				// Update rule
				rule.MidiMessage.Channel = globalMidiChannel
				rule.MidiMessage.Type = configuration.ControlChange
				rule.MidiMessage.Controller = programData[knobIndex]
				rule.MidiMessage.MinValue = programData[knobIndex+1]
				rule.MidiMessage.MaxValue = programData[knobIndex+2]
				updatedRules = append(updatedRules, rule)
			} else {
				log.Warn().Msgf("Unknown device control path %s", rule.MidiMessage.DeviceControlPath)
			}
		} else {
			updatedRules = append(updatedRules, rule)
		}
	}
	return updatedRules
}
