package telemetry

import (
	"encoding/binary"
	"errors"
)

// Total number of samples
type TSMP struct {
	Samples uint32
}

func (t *TSMP) Parse(bytes []byte, scale *SCAL) error {
	if 4 != len(bytes) {
		return errors.New("Invalid length TSMP packet")
	}

	t.Samples = binary.BigEndian.Uint32(bytes[0:4])

	return nil
}
