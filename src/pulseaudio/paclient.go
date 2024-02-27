package pulseaudio

import (
	"slices"

	"github.com/fluciotto/pamixermidicontrol/src/configuration"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/the-jonsey/pulseaudio"
)

type Stream struct {
	name     string
	fullName string
	paStream interface{}
}

type PAClient struct {
	log             zerolog.Logger
	context         *pulseaudio.Client
	outputs         []Stream
	playbackStreams []Stream
	inputs          []Stream
	recordStreams   []Stream
}

func NewPAClient() *PAClient {
	context, err := pulseaudio.NewClient()
	if err != nil {
		panic(err)
	}
	client := &PAClient{
		log:             log.With().Str("module", "PulseAudio").Logger(),
		context:         context,
		outputs:         []Stream{},
		playbackStreams: []Stream{},
		inputs:          []Stream{},
		recordStreams:   []Stream{},
	}
	return client
}

func (client *PAClient) List() {
	if server, err := client.context.ServerInfo(); err == nil {
		client.log.Info().Msgf("PulseAudio server\t\tHostname=%s", server.Hostname)
		client.log.Info().Msgf("\t\t\t\tUser=%s", server.User)
		client.log.Info().Msgf("\t\t\t\t%s v%s", server.PackageName, server.PackageVersion)
		client.log.Info().Msgf("\t\t\t\tChannels=%d", server.SampleSpec.Channels)
		client.log.Info().Msgf("\t\t\t\tFormat=%d", server.SampleSpec.Format)
		client.log.Info().Msgf("\t\t\t\tRate=%d", server.SampleSpec.Rate)
		client.log.Info().Msgf("\t\t\t\tDefault input=%s", server.DefaultSource)
		client.log.Info().Msgf("\t\t\t\tDefault output=%s", server.DefaultSink)
	}
	client.refreshStreams()
	// List sinks
	lo.ForEach(client.outputs, func(stream Stream, i int) {
		client.log.Info().Msgf("Found output device:\t%s", stream.name)
	})
	// List sources
	lo.ForEach(client.inputs, func(stream Stream, i int) {
		client.log.Info().Msgf("Found input device:\t%s", stream.name)
	})
	// List sinks inputs
	lo.ForEach(client.playbackStreams, func(stream Stream, i int) {
		client.log.Info().Msgf("Found playback stream:\t%s", stream.name)
	})
	// List sources
	lo.ForEach(client.recordStreams, func(stream Stream, i int) {
		client.log.Info().Msgf("Found record stream:\t%s", stream.name)
	})
}

func (client *PAClient) refreshStreams() error {
	// Sinks
	sinks, err := client.context.Sinks()
	if err != nil {
		panic(err)
	}
	client.outputs = lo.Map(sinks, func(sink pulseaudio.Sink, i int) Stream {
		return Stream{
			name:     sink.Description,
			fullName: sink.Name,
			paStream: sink,
		}
	})
	// Sources
	sources, err := client.context.Sources()
	if err != nil {
		panic(err)
	}
	client.inputs = lo.Map(sources, func(source pulseaudio.Source, i int) Stream {
		return Stream{
			name:     source.Description,
			fullName: source.Name,
			paStream: source,
		}
	})
	// Sinks inputs
	sinksInputs, err := client.context.SinkInputs()
	if err != nil {
		panic(err)
	}
	client.playbackStreams = lo.Map(sinksInputs, func(sinkInput pulseaudio.SinkInput, i int) Stream {
		return Stream{
			name:     sinkInput.PropList["application.name"],
			fullName: sinkInput.PropList["module-stream-restore.id"],
			paStream: sinkInput,
		}
	})
	// Sources outputs
	sourcesOutputs, err := client.context.SourceOutputs()
	if err != nil {
		panic(err)
	}
	client.recordStreams = lo.Map(sourcesOutputs, func(sourceOutput pulseaudio.SourceOutput, i int) Stream {
		return Stream{
			name:     sourceOutput.PropList["application.name"],
			fullName: sourceOutput.PropList["module-stream-restore.id"],
			paStream: sourceOutput,
		}
	})
	return nil
}

func (client *PAClient) ProcessVolumeAction(action configuration.Action, volumePercent float32) error {
	var streams []Stream
	client.refreshStreams()
	switch target := action.Target.(type) {
	case *configuration.TypedTarget:
		if target.Type == configuration.OutputDevice {
			if target.Name == "Default" {
				if defaultSink, err := client.context.GetDefaultSink(); err == nil {
					streams = slices.Concat(streams, lo.Filter(client.outputs, func(stream Stream, i int) bool {
						return stream.fullName == defaultSink.Name
					}))
				}
			} else {
				streams = slices.Concat(streams, lo.Filter(client.outputs, func(stream Stream, i int) bool {
					return stream.name == target.Name
				}))
			}
		} else if target.Type == configuration.InputDevice {
			if target.Name == "Default" {
				if defaultSource, err := client.context.GetDefaultSource(); err == nil {
					streams = slices.Concat(streams, lo.Filter(client.inputs, func(stream Stream, i int) bool {
						return stream.fullName == defaultSource.Name
					}))
				}
			} else {
				streams = slices.Concat(streams, lo.Filter(client.inputs, func(stream Stream, i int) bool {
					return stream.name == target.Name
				}))
			}
		} else if target.Type == configuration.PlaybackStream {
			streams = slices.Concat(streams, lo.Filter(client.playbackStreams, func(stream Stream, i int) bool {
				return stream.name == target.Name
			}))
		} else if target.Type == configuration.RecordStream {
			streams = slices.Concat(streams, lo.Filter(client.recordStreams, func(stream Stream, i int) bool {
				return stream.name == target.Name
			}))
		}
	case *configuration.Target:
	default:
	}
	lo.ForEach(streams, func(stream Stream, index int) {
		switch st := stream.paStream.(type) {
		case pulseaudio.Sink:
			st.SetVolume(volumePercent)
			client.log.Debug().Msgf("Set %s volume to %f", stream.name, volumePercent)
		case pulseaudio.SinkInput:
			st.SetVolume(volumePercent)
			client.log.Debug().Msgf("Set %s volume to %f", stream.name, volumePercent)
		case pulseaudio.Source:
			st.SetVolume(volumePercent)
			client.log.Debug().Msgf("Set %s volume to %f", stream.name, volumePercent)
		case pulseaudio.SourceOutput:
			st.SetVolume(volumePercent)
			client.log.Debug().Msgf("Set %s volume to %f", stream.name, volumePercent)
		}
	})
	return nil
}

func (client *PAClient) ProcessToggleMute(action configuration.Action) error {
	var streams []Stream
	client.refreshStreams()
	switch target := action.Target.(type) {
	case *configuration.TypedTarget:
		if target.Type == configuration.OutputDevice {
			if target.Name == "Default" {
				if defaultSink, err := client.context.GetDefaultSink(); err == nil {
					streams = slices.Concat(streams, lo.Filter(client.outputs, func(stream Stream, i int) bool {
						return stream.fullName == defaultSink.Name
					}))
				}
			} else {
				streams = slices.Concat(streams, lo.Filter(client.outputs, func(stream Stream, i int) bool {
					return stream.name == target.Name
				}))
			}
		} else if target.Type == configuration.InputDevice {
			if target.Name == "Default" {
				if defaultSource, err := client.context.GetDefaultSource(); err == nil {
					streams = slices.Concat(streams, lo.Filter(client.inputs, func(stream Stream, i int) bool {
						return stream.fullName == defaultSource.Name
					}))
				}
			} else {
				streams = slices.Concat(streams, lo.Filter(client.inputs, func(stream Stream, i int) bool {
					return stream.name == target.Name
				}))
			}
		} else if target.Type == configuration.PlaybackStream {
			streams = slices.Concat(streams, lo.Filter(client.playbackStreams, func(stream Stream, i int) bool {
				return stream.name == target.Name
			}))
		} else if target.Type == configuration.RecordStream {
			streams = slices.Concat(streams, lo.Filter(client.recordStreams, func(stream Stream, i int) bool {
				return stream.name == target.Name
			}))
		}
	case *configuration.Target:
	default:
	}
	lo.ForEach(streams, func(stream Stream, index int) {
		switch st := stream.paStream.(type) {
		case pulseaudio.Sink:
			st.ToggleMute()
			client.log.Debug().Msgf("Toggled mute on %s", stream.name)
		case pulseaudio.SinkInput:
			st.ToggleMute()
			client.log.Debug().Msgf("Toggled mute on %s", stream.name)
		case pulseaudio.Source:
			st.ToggleMute()
			client.log.Debug().Msgf("Toggled mute on %s", stream.name)
		case pulseaudio.SourceOutput:
			st.ToggleMute()
			client.log.Debug().Msgf("Toggled mute on %s", stream.name)
		}
	})
	return nil
}

func (client *PAClient) SetDefaultOutput(action configuration.Action) error {
	client.refreshStreams()
	switch target := action.Target.(type) {
	case *configuration.Target:
		lo.ForEach(
			lo.Filter(client.outputs, func(stream Stream, i int) bool {
				return stream.name == target.Name
			}), func(stream Stream, i int) {
				client.context.SetDefaultSink(stream.fullName)
				client.log.Debug().Msgf("Set default output to %s", stream.name)
			})
	case *configuration.TypedTarget:
	default:
	}
	return nil
}
