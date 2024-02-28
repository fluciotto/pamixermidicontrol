package korgNanokontrol2

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"github.com/fluciotto/pamixermidicontrol/src/configuration"
	"github.com/fluciotto/pamixermidicontrol/src/device"
	"github.com/fluciotto/pamixermidicontrol/src/device/korg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/sysex"
)

type KorgNanoKontrol2 struct {
	log        zerolog.Logger
	DeviceName string
}

func New(name string) *KorgNanoKontrol2 {
	return &KorgNanoKontrol2{
		log:        log.With().Str("device", "Korg nanoKontrol2").Logger(),
		DeviceName: name,
	}
}

func (d *KorgNanoKontrol2) identityMessage(channel byte) *device.SysExMessage {
	request := sysex.IdentityRequest(channel)
	responseHandler := func(bytes []byte) (rawData []byte, processedData []byte, err error) {
		log := d.log.With().Str("SysEx", "IdentityMessage").Logger()
		if len(bytes) != 15-2 {
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

func (d *KorgNanoKontrol2) searchDeviceMessage(echoBackID byte) *device.SysExMessage {
	request := []byte{
		0xf0,
		0x42, // Korg
		0x50, 0x00, echoBackID, 0xf7,
	}
	responseHandler := func(bytes []byte) (rawData []byte, processedData []byte, err error) {
		log := d.log.With().Str("SysEx", "searchDeviceMessage").Logger()
		if len(bytes) != 15-2 {
			log.Error().Msgf("Search device response has a bad length: %d", len(bytes))
			return []byte{}, []byte{}, fmt.Errorf("search device response has a bad length: %d", len(bytes))
		}
		manufacturerId := bytes[0]
		globalMidiChannel := bytes[3]
		echoBackId := bytes[4]
		familyId := binary.LittleEndian.Uint16([]byte{bytes[5], bytes[6]})
		memberId := binary.LittleEndian.Uint16([]byte{bytes[7], bytes[8]})
		version := fmt.Sprintf("%d.%d.%d", bytes[9], bytes[10], bytes[11])
		log.Info().Msgf("Global MIDI channel 0x%X", globalMidiChannel)
		log.Info().Msgf("Manufacturer ID 0x%X", manufacturerId)
		log.Info().Msgf("Family ID 0x%X", familyId)
		log.Info().Msgf("Member ID 0x%X", memberId)
		log.Info().Msgf("Version %s", version)
		log.Info().Msgf("Echo back ID 0x%X", echoBackId)
		return bytes, nil, nil
	}
	return device.NewSysExMessage(request, responseHandler)
}

func (d *KorgNanoKontrol2) modeMessage(channel byte) *device.SysExMessage {
	request := []byte{
		0xf0,
		0x42, // Korg
		0x40 + channel&0x0f,
		0x00, 0x01, 0x13, 0x00, // nanoKontrol2 ID
		0x1f, // Data dump request
		0x12, // Mode request
		0x00,
		0xf7,
	}
	responseHandler := func(bytes []byte) (rawData []byte, processedData []byte, err error) {
		log := d.log.With().Str("SysEx", "modeMessage").Logger()
		if len(bytes) != 11-2 {
			log.Error().Msgf("Mode response has a bad length: %d", len(bytes))
			return []byte{}, []byte{}, fmt.Errorf("mode response has a bad length: %d", len(bytes))
		}
		manufacturerId := bytes[0]
		globalMidiChannel := bytes[1] & 0xf
		mode := bytes[8]
		log.Info().Msgf("Global MIDI channel 0x%X", globalMidiChannel)
		log.Info().Msgf("Manufacturer ID 0x%X", manufacturerId)
		log.Info().Msgf("Mode 0x%X", mode)
		return bytes, nil, nil
	}
	return device.NewSysExMessage(request, responseHandler)
}

func (d *KorgNanoKontrol2) sceneDumpRequestMessage(channel byte) *device.SysExMessage {
	request := []byte{
		0xf0,
		0x42, // Korg
		0x40 + channel&0x0f,
		0x00, 0x01, 0x13, 0x00, // nanoKontrol2 ID
		0x1f, // Data dump request
		0x10, // Current scene data dump request
		0x00,
		0xf7,
	}
	responseHandler := func(bytes []byte) (rawData []byte, processedData []byte, err error) {
		log := d.log.With().Str("SysEx", "sceneDumpMessage").Logger()
		if len(bytes) != 402-2 {
			log.Error().Msgf("Scene dump request response has a bad length: %d", len(bytes))
			return []byte{}, []byte{}, fmt.Errorf("scene dump request response has a bad length: %d", len(bytes))
		}
		sceneMidiData := bytes[12:]
		sceneData := korg.MidiDataToData(sceneMidiData)
		return bytes, sceneData, nil
	}
	return device.NewSysExMessage(request, responseHandler)
}

func (d *KorgNanoKontrol2) sceneDumpMessage(channel byte, sceneData []byte) *device.SysExMessage {
	if len(sceneData) != 339 {
		panic(fmt.Sprintf("Scene data has a bad length %d", len(sceneData)))
	}
	request := slices.Concat(
		[]byte{
			0xf0,
			0x42, // Korg
			0x40 + channel&0x0f,
			0x00, 0x01, 0x13, 0x00, // nanoKontrol2 ID
			0x7f, // Data dump commad
			0x7f, //
			0x02,
			0x03,
			0x05,
			0x40, // Current scene data dump
		},
		sceneData,
		[]byte{0xf7},
	)
	responseHandler := func(bytes []byte) (rawData []byte, processedData []byte, err error) {
		log := d.log.With().Str("SysEx", "sceneSaveMessage").Logger()
		if len(bytes) != 11-2 {
			log.Error().Msgf("Scene dump response has a bad length: %d", len(bytes))
			return []byte{}, []byte{}, fmt.Errorf("scene dump response has a bad length: %d", len(bytes))
		}
		result := bytes[7] // 0x23 OK, 0x24 Error
		log.Info().Msgf("Scene dump result 0x%X", result)
		return bytes, nil, nil
	}
	return device.NewSysExMessage(request, responseHandler)
}

func (d *KorgNanoKontrol2) sceneWriteMessage(channel byte) *device.SysExMessage {
	request := []byte{
		0xf0,
		0x42, // Korg
		0x40 + channel&0x0f,
		0x00, 0x01, 0x13, 0x00, // nanoKontrol2 ID
		0x1f, // Data dump commad
		0x11, // Scene write request
		0x00,
		0xf7,
	}
	responseHandler := func(bytes []byte) (rawData []byte, processedData []byte, err error) {
		log := d.log.With().Str("SysEx", "sceneSaveMessage").Logger()
		if len(bytes) != 11-2 {
			log.Error().Msgf("Scene dump response has a bad length: %d", len(bytes))
			return []byte{}, []byte{}, fmt.Errorf("scene dump response has a bad length: %d", len(bytes))
		}
		result := bytes[7] // 0x21 OK, 0x22 Error
		log.Info().Msgf("Scene write result 0x%X", result)
		return bytes, nil, nil
	}
	return device.NewSysExMessage(request, responseHandler)
}

func (d *KorgNanoKontrol2) UpdateRules(
	rules []configuration.Rule,
	c chan []byte,
	out drivers.Out,
) (updatedRules []configuration.Rule) {
	// Fetch scene data from device
	_, sceneData, err := d.sceneDumpRequestMessage(0).Send(c, out, d.log)
	if err != nil {
		panic(err)
	}
	var assignTypeToMidiMessageType = func(assignType byte) configuration.MidiMessageType {
		if assignType == 1 {
			return configuration.ControlChange
		} else if assignType == 2 {
			return configuration.Note
		}
		return configuration.None
	}
	// Get global MIDI channel from scene data
	globalMidiChannel := sceneData[0]
	// Update rules with scene data
	for _, rule := range rules {
		if rule.MidiMessage.DeviceName != d.DeviceName {
			continue
		}
		if rule.MidiMessage.DeviceControlPath != "" {
			//
			groupRe := regexp.MustCompile("^Group([1-8])/(Slider|Knob|Solo|Mute|Record)$")
			transportRe := regexp.MustCompile("^Transport/.*")
			transportTrackRe := regexp.MustCompile("^Transport/Track/(Prev|Next)$")
			transportCycleRe := regexp.MustCompile("^Transport/Cycle$")
			transportMarkerRe := regexp.MustCompile("^Transport/Marker/(Set|Prev|Next)$")
			transportBottomRe := regexp.MustCompile("^Transport/(Rewind|FastForward|Stop|Play|Rec)$")

			//
			if groupRe.MatchString(rule.MidiMessage.DeviceControlPath) {
				matches := groupRe.FindStringSubmatch(rule.MidiMessage.DeviceControlPath)
				groupNumber64, _ := strconv.ParseUint(matches[1], 10, 8)
				groupNumber := uint8(groupNumber64)
				control := matches[2]
				sceneDataGroupIndex := 3 + (groupNumber-1)*31
				// Update rule
				rule.MidiMessage.Type = configuration.ControlChange
				rule.MidiMessage.Channel = sceneData[sceneDataGroupIndex]
				if rule.MidiMessage.Channel == 16 {
					rule.MidiMessage.Channel = globalMidiChannel
				}
				if control == "Slider" {
					rule.MidiMessage.Controller = sceneData[sceneDataGroupIndex+3]
					rule.MidiMessage.MinValue = sceneData[sceneDataGroupIndex+4]
					rule.MidiMessage.MaxValue = sceneData[sceneDataGroupIndex+5]
				} else if control == "Knob" {
					rule.MidiMessage.Controller = sceneData[sceneDataGroupIndex+9]
					rule.MidiMessage.MinValue = sceneData[sceneDataGroupIndex+10]
					rule.MidiMessage.MaxValue = sceneData[sceneDataGroupIndex+11]
				} else if control == "Solo" {
					rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataGroupIndex+13])
					rule.MidiMessage.Controller = sceneData[sceneDataGroupIndex+15]
					rule.MidiMessage.MinValue = sceneData[sceneDataGroupIndex+16]
					rule.MidiMessage.MaxValue = sceneData[sceneDataGroupIndex+17]
				} else if control == "Mute" {
					rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataGroupIndex+19])
					rule.MidiMessage.Controller = sceneData[sceneDataGroupIndex+21]
					rule.MidiMessage.MinValue = sceneData[sceneDataGroupIndex+22]
					rule.MidiMessage.MaxValue = sceneData[sceneDataGroupIndex+23]
				} else if control == "Record" {
					rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataGroupIndex+25])
					rule.MidiMessage.Controller = sceneData[sceneDataGroupIndex+27]
					rule.MidiMessage.MinValue = sceneData[sceneDataGroupIndex+28]
					rule.MidiMessage.MaxValue = sceneData[sceneDataGroupIndex+29]
				}
				updatedRules = append(updatedRules, rule)
			} else if transportRe.MatchString(rule.MidiMessage.DeviceControlPath) {
				sceneDataTransportIndex := 251
				rule.MidiMessage.Channel = sceneData[sceneDataTransportIndex]
				if rule.MidiMessage.Channel == 16 {
					rule.MidiMessage.Channel = globalMidiChannel
				}
				if transportTrackRe.MatchString(rule.MidiMessage.DeviceControlPath) {
					matches := transportTrackRe.FindStringSubmatch(rule.MidiMessage.DeviceControlPath)
					control := matches[1]
					if control == "Prev" {
						sceneDataIndex := 252
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					} else if control == "Next" {
						sceneDataIndex := 258
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					}
					updatedRules = append(updatedRules, rule)
				} else if transportCycleRe.MatchString(rule.MidiMessage.DeviceControlPath) {
					sceneDataIndex := 264
					rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
					rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
					rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
					rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
					rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					updatedRules = append(updatedRules, rule)
				} else if transportMarkerRe.MatchString(rule.MidiMessage.DeviceControlPath) {
					matches := transportMarkerRe.FindStringSubmatch(rule.MidiMessage.DeviceControlPath)
					control := matches[1]
					if control == "Set" {
						sceneDataIndex := 270
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					} else if control == "Prev" {
						sceneDataIndex := 276
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					} else if control == "Next" {
						sceneDataIndex := 282
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					}
					updatedRules = append(updatedRules, rule)
				} else if transportBottomRe.MatchString(rule.MidiMessage.DeviceControlPath) {
					matches := transportBottomRe.FindStringSubmatch(rule.MidiMessage.DeviceControlPath)
					control := matches[1]
					// Update rule
					rule.MidiMessage.Channel = sceneData[sceneDataTransportIndex]
					if rule.MidiMessage.Channel == 16 {
						rule.MidiMessage.Channel = globalMidiChannel
					}
					if control == "Rewind" {
						sceneDataIndex := 288
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					} else if control == "FastForward" {
						sceneDataIndex := 294
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					} else if control == "Stop" {
						sceneDataIndex := 300
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					} else if control == "Play" {
						sceneDataIndex := 306
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					} else if control == "Rec" {
						sceneDataIndex := 312
						rule.MidiMessage.Type = assignTypeToMidiMessageType(sceneData[sceneDataIndex+1])
						rule.MidiMessage.Note = sceneData[sceneDataIndex+2]
						rule.MidiMessage.Controller = sceneData[sceneDataIndex+2]
						rule.MidiMessage.MinValue = sceneData[sceneDataIndex+3]
						rule.MidiMessage.MaxValue = sceneData[sceneDataIndex+4]
					}
					updatedRules = append(updatedRules, rule)
				} else {
					log.Warn().Msgf("Unknown device control path %s", rule.MidiMessage.DeviceControlPath)
				}
			} else {
				log.Warn().Msgf("Unknown device control path %s", rule.MidiMessage.DeviceControlPath)
			}
		} else {
			updatedRules = append(updatedRules, rule)
		}
	}
	return updatedRules
}
