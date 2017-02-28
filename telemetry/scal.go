package telemetry

import (
	"encoding/binary"
	"errors"
)

// Scale - contains slice of multipliers for subsequent data
type SCAL struct {
	Values []int
}

func (scale *SCAL) Parse(bytes []byte, size int64) error {
	s := int(size)

	if 0 != len(bytes)%s {
		return errors.New("Invalid length SCAL packet")
	}

	if s == 2 {
		for i := 0; i < len(bytes); i += s {
			scale.Values = append(scale.Values, int(binary.BigEndian.Uint16(bytes[i:i+s])))
		}
	} else if s == 4 {
		for i := 0; i < len(bytes); i += s {
			scale.Values = append(scale.Values, int(binary.BigEndian.Uint32(bytes[i:i+s])))
		}
	} else {
		return errors.New("Unknown SCAL length")
	}

	return nil
}
