package telemetry

import (
	"encoding/binary"
	"errors"
	"math"
)

type TSMP struct {
	Temp float32
}

func (t *TSMP) Parse(bytes []byte, scale *SCAL) error {
	if 4 != len(bytes) {
		return errors.New("Invalid length TMPC packet")
	}

	bits := binary.BigEndian.Uint32(bytes[0:4])

	t.Temp = math.Float32frombits(bits)

	return nil
}
