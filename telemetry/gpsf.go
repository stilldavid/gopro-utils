package telemetry

import (
	"encoding/binary"
	"errors"
)

// GPS Fix
type GPSF struct {
	F uint32
}

func (gpsf *GPSF) Parse(bytes []byte) error {
	if 4 != len(bytes) {
		return errors.New("Invalid length GPSF packet")
	}

	gpsf.F = binary.BigEndian.Uint32(bytes[0:4])

	return nil
}
