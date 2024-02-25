package telemetry

import (
	"encoding/binary"
	"errors"
)

// GPS position accuracy in cm
type GPSP struct {
	Accuracy uint16
}

func (gpsp *GPSP) Parse(bytes []byte) error {
	if len(bytes) != 2 {
		return errors.New("invalid length GPSP packet")
	}

	gpsp.Accuracy = binary.BigEndian.Uint16(bytes[0:2])

	return nil
}
