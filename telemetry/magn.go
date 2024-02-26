package telemetry

import (
	"encoding/binary"
	"errors"
)

// Accelerometer in m/s for XYZ
type MAGN struct {
	X float64
	Y float64
	Z float64
}

func (magn *MAGN) Parse(bytes []byte, scale *SCAL) error {
	if len(bytes) != 6 {
		return errors.New("invalid length MAGN packet")
	}

	magn.X = float64(int16(binary.BigEndian.Uint16(bytes[0:2]))) / float64(scale.Values[0])
	magn.Y = float64(int16(binary.BigEndian.Uint16(bytes[2:4]))) / float64(scale.Values[0])
	magn.Z = float64(int16(binary.BigEndian.Uint16(bytes[4:6]))) / float64(scale.Values[0])

	return nil
}
