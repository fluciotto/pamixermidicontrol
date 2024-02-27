package pamixermidicontrol

import (
	"fmt"
	"os"
	"time"

	"github.com/DavidGamba/go-getoptions"
	"github.com/fluciotto/pamixermidicontrol/src/configuration"
	"github.com/fluciotto/pamixermidicontrol/src/midi"
	"github.com/fluciotto/pamixermidicontrol/src/pulseaudio"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func Run() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Create PulseAudio client
	paClient := pulseaudio.NewPAClient()

	// Parse command line
	opt := getoptions.New()
	opt.Self("", "Control your PulseAudio mixer with MIDI controller(s)")
	opt.HelpSynopsisArg("", "")
	opt.HelpCommand("help", opt.Alias("h"), opt.Description("Show this help"))
	opt.Bool("list", false, opt.Alias("l"), opt.Description("List MIDI ports & PulseAudio objects"))
	opt.Bool("list-midi", false, opt.Alias("m"), opt.Description("List MIDI ports"))
	opt.Bool("list-pulse", false, opt.Alias("p"), opt.Description("List PulseAudio objects"))
	opt.Parse(os.Args[1:])
	if opt.Called("help") {
		fmt.Fprint(os.Stderr, opt.Help())
		os.Exit(0)
	}
	if opt.Called("list") {
		midi.List()
		paClient.List()
		os.Exit(0)
	}
	if opt.Called("list-midi") {
		midi.List()
		os.Exit(0)
	}
	if opt.Called("list-pulse") {
		paClient.List()
		os.Exit(0)
	}

	// Configuration
	config, err := configuration.Load()
	if err != nil {
		log.Error().Msgf("Configuration error %+v", err)
		os.Exit(1)
	}
	log.Info().Msg("Loaded configuration")
	// fmt.Printf("%+v\n", config)

	// Create MIDI clients
	for _, midiDevice := range config.MidiDevices {
		deviceRules := lo.Filter(config.Rules, func(rule configuration.Rule, i int) bool {
			return rule.MidiMessage.DeviceName == midiDevice.Name
		})
		midiClient := midi.NewMidiClient(paClient, midiDevice, deviceRules)
		go midiClient.Run()
	}

	select {}
}
