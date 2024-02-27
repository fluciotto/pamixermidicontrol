package device

import (
	"github.com/rs/zerolog"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

// type ResponseHandlerReturn struct {
// 	rawData       []byte
// 	processedData []byte
// 	error         error
// }

// type ResponseHandlerReturn = []byte,  []byte, error

type SysExMessage struct {
	Request         []byte
	ResponseHandler func(bytes []byte) (rawData []byte, processedData []byte, err error)
	// ResponseHandler func(bytes []byte) ResponseHandlerReturn
}

func NewSysExMessage(
	Request []byte,
	ResponseHandler func(bytes []byte) (rawData []byte, processedData []byte, err error),
) *SysExMessage {
	return &SysExMessage{
		Request:         Request,
		ResponseHandler: ResponseHandler,
	}
}

func (d *SysExMessage) Send(
	c chan []byte, out drivers.Out,
	log zerolog.Logger,
) (rawData []byte, processedData []byte, err error) {
	theLog := log.With().Str("module", "SysExMessage").Logger()
	send, err := midi.SendTo(out)
	if err != nil {
		panic(err)
	}
	request := d.Request
	theLog.Debug().Msgf("Sending SysEx message: % X", request)
	if err = send(request); err != nil {
		panic(err)
	}
	response := <-c
	return d.ResponseHandler(response)
}
