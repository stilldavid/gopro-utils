package telemetry

import (
	"encoding/binary"
	"errors"
)

type GPSP struct {
	Heading int16
}

func (gpsp *GPSP) Parse(bytes []byte) error {
	if 2 != len(bytes) {
		return errors.New("Invalid length GPSP packet")
	}

	gpsp.Heading = int16(binary.BigEndian.Uint16(bytes[0:2]))

	return nil
}
