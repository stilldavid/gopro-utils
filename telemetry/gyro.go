package telemetry

import (
	"encoding/binary"
	"errors"
)

type GYRO struct {
	X int16
	Y int16
	Z int16
}

func (gyro *GYRO) Parse(bytes []byte) error {
	if 6 != len(bytes) {
		return errors.New("Invalid length GYRO packet")
	}

	gyro.X = int16(binary.BigEndian.Uint16(bytes[0:2]))
	gyro.Y = int16(binary.BigEndian.Uint16(bytes[2:4]))
	gyro.Z = int16(binary.BigEndian.Uint16(bytes[4:6]))

	return nil
}
