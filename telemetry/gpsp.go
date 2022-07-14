package telemetry

import (
        "encoding/binary"
        "errors"
)

// GPS position accuracy in cm
type GPSP struct {
        Accuracy float64
}

func (gpsp *GPSP) Parse(bytes []byte) error {
        if 2 != len(bytes) {
                return errors.New("Invalid length GPSP packet")
        }

        gpsp.Accuracy = float64(binary.BigEndian.Uint16(bytes[0:2]))

        return nil
}
