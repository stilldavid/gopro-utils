package gpmf

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

var fourCC = [][4]byte{
	{0x44, 0x45, 0x56, 0x43}, // "DEVC": Unique Device Source
	{0x44, 0x56, 0x49, 0x44}, // "DVID": Device ID
	{0x44, 0x56, 0x4E, 0x4D}, // "DVNM": Device Name
	{0x53, 0x54, 0x52, 0x4D}, // "STRM": Nested Signal Stream
	{0x53, 0x54, 0x4E, 0x4D}, // "STNM": Stream Name
	{0x52, 0x4D, 0x52, 0x4B}, // "RMRK": Comments For Any Stream
	{0x53, 0x43, 0x41, 0x4C}, // "SCAL": Scaling Factor
	{0x53, 0x49, 0x55, 0x4E}, // "SIUN": Standard Units
	{0x55, 0x4E, 0x49, 0x54}, // "UNIT": Display Units
	{0x54, 0x59, 0x50, 0x45}, // "TYPE": Typedef for complex structure
	{0x54, 0x53, 0x4D, 0x50}, // "TSMP": Total Samples Delivered
	{0x45, 0x4D, 0x50, 0x54}, // "EMPT": Empty Payload Count
	{0x41, 0x43, 0x43, 0x4C}, // "ACCL": Accelerometer 3 Axis
	{0x47, 0x59, 0x52, 0x4F}, // "GYRO": Gyroscope 3 Axis
	{0x47, 0x50, 0x53, 0x35}, // "GPS5": GPS Position + Speed
	{0x47, 0x50, 0x53, 0x55}, // "GPSU": GPS UTC Time
	{0x47, 0x50, 0x53, 0x46}, // "GPSF": GPS Fix
	{0x47, 0x50, 0x53, 0x50}, // "GPSP": GPS Precision
	{0x49, 0x53, 0x4F, 0x47}, // "ISOG": Image Sensor Gain
	{0x53, 0x48, 0x55, 0x54}, // "SHUT": Exposure Time
}

var formats = []byte{
	0x62, // "b": 8-bit signed integer
	0x42, // "B": 8-bit unsigned integer
	0x63, // "c": 8-bit 'c' style ASCII character
	0x73, // "s": 16-bit signed integer
	0x53, // "S": 16-bit unsigned integer
	0x6C, // "l": 32-bit signed integer
	0x4C, // "L": 32-bit unsigned integer
	0x66, // "f": 32-bit float (IEEE 754)
	0x64, // "d": 64-bit double precision (IEEE 754)
	0x46, // "F": 32-bit four character key
	0x47, // "G": 128-bit ID (like UUID)
	0x6A, // "j": 64-bit signed unsigned integer
	0x4A, // "J": 64-bit unsigned unsigned integer
	0x71, // "q": 32-bit Q Number Q15.16
	0x51, // "Q": 64-bit Q Number Q31.32
	0x55, // "U": 16-byte UTC Date and Time string
	0x3F, // "?": 32-bit unsigned integer Nested metadata
}

// KLV - Go-Pro-Metadata-Format Key-Length-Value
type KLV struct {
	FourCC []byte // Four CC Key
	Format byte   // Format of the data
	Size   uint8  // Size of the object
	Count  uint16 // Number of object
}

// SCAL - Slice of divisors for scaling data
type SCAL struct {
	Divisor []int
}

// ACCL - 3-axis accelerometer measurements (meters/sec^2)
type ACCL struct {
	X float64 `json:"accl_x"`
	Y float64 `json:"accl_y"`
	Z float64 `json:"accl_z"`
}

// GYRO - 3-axis gyroscope measurement (radians/sec)
type GYRO struct {
	X float64 `json:"gyro_x"`
	Y float64 `json:"gyro_y"`
	Z float64 `json:"gyro_z"`
}

// GPS5 - GPS data (latitude/longitude/altitude/speed/3D/time)
type GPS5 struct {
	Lat     float64 `json:"lat"`    // Latitude (Degrees)
	Lon     float64 `json:"lon"`    // Longitude (Degrees)
	Alt     float64 `json:"alt"`    // Altitude
	Speed2D float64 `json:"spd_2d"` // m/s
	Speed3D float64 `json:"spd_3d"` // m/s
}

// GPSF - GPS fix (0: No Fix, 2: 2D, 3: 3D)
type GPSF struct {
	Fix uint32 `json:"gps_fix"`
}

// GPSP - GPS position accuracy (centimeters)
type GPSP struct {
	Accuracy uint16 `json:"gps_accuracy"`
}

// GPSU - GPS acquired UTC timestamp
type GPSU struct {
	Time time.Time
}

// TMPC - Temperature (Degrees C)
type TMPC struct {
	Temp float32
}

// TSMP - Total number of samples
type TSMP struct {
	Samples uint32
}

// Parse (KLV) - Parse byte slice into KLV struct
func (klv *KLV) Parse(bytes []byte) error {
	// Check length
	if len(bytes) != 8 {
		return errors.New("KLV: Invalid packet length")
	}

	// Four CC
	klv.FourCC = bytes[0:4]
	for _, c := range klv.FourCC {
		if c < 0x41 || c > 0x5A {
			return errors.New("KLV: Invalid Four CC Character")
		}
	}

	// Format
	klv.Format = bytes[4]
	hasFormat := false
	for _, f := range formats {
		hasFormat = hasFormat || (f == klv.Format)
	}
	if !hasFormat {
		return errors.New("KLV: Invalid Format Character")
	}

	// Size & Count
	klv.Size = bytes[5]
	klv.Count = binary.BigEndian.Uint16(bytes[6:8])

	// No error
	return nil
}

// Parse (SCAL) - Parse byte slice into SCAL struct
func (scale *SCAL) Parse(bytes []byte, size int64) error {
	// Number of bytes per divisor
	s := int(size)

	// No left over bytes
	if 0 != len(bytes)%s {
		return errors.New("SCAL: Invalid packet length")
	}

	// Two or Four byte scalars
	if s == 2 {
		for i := 0; i < len(bytes); i += 2 {
			scale.Divisor = append(scale.Divisor, int(binary.BigEndian.Uint16(bytes[i:i+2])))
		}
	} else if s == 4 {
		for i := 0; i < len(bytes); i += 4 {
			scale.Divisor = append(scale.Divisor, int(binary.BigEndian.Uint32(bytes[i:i+4])))
		}
	} else {
		return errors.New("SCAL: Invalid packet length")
	}

	// No error
	return nil
}

// Parse (ACCL) - Parse byte slice into ACCL struct and scale
func (accl *ACCL) Parse(bytes []byte, scale *SCAL) error {
	// Check length
	if 6 != len(bytes) {
		return errors.New("ACCL: Invalid packet length")
	}

	// Accelerometer 3D
	accl.X = float64(int16(binary.BigEndian.Uint16(bytes[0:2]))) / float64(scale.Divisor[0])
	accl.Y = float64(int16(binary.BigEndian.Uint16(bytes[2:4]))) / float64(scale.Divisor[0])
	accl.Z = float64(int16(binary.BigEndian.Uint16(bytes[4:6]))) / float64(scale.Divisor[0])

	// No error
	return nil
}

// Parse (GYRO) - Parse byte slice into GYRO struct and scale
func (gyro *GYRO) Parse(bytes []byte, scale *SCAL) error {
	// Check length
	if 6 != len(bytes) {
		return errors.New("GYRO: Invalid packet length")
	}

	// Gyroscope 3D
	gyro.X = float64(int16(binary.BigEndian.Uint16(bytes[0:2]))) / float64(scale.Divisor[0])
	gyro.Y = float64(int16(binary.BigEndian.Uint16(bytes[2:4]))) / float64(scale.Divisor[0])
	gyro.Z = float64(int16(binary.BigEndian.Uint16(bytes[4:6]))) / float64(scale.Divisor[0])

	// No error
	return nil
}

// Parse (GPS5) - Parse byte slice into GPS5 struct and scale
func (gps5 *GPS5) Parse(bytes []byte, scale *SCAL) error {
	// Check length
	if 20 != len(bytes) {
		return errors.New("GPS5: Inavlid packet length")
	}

	// Geodetic location
	gps5.Lat = float64(int32(binary.BigEndian.Uint32(bytes[0:4]))) / float64(scale.Divisor[0])
	gps5.Lon = float64(int32(binary.BigEndian.Uint32(bytes[4:8]))) / float64(scale.Divisor[1])
	gps5.Alt = float64(int32(binary.BigEndian.Uint32(bytes[8:12]))) / float64(scale.Divisor[2])

	// Speed 2D/3D
	gps5.Speed2D = float64(int32(binary.BigEndian.Uint32(bytes[12:16]))) / float64(scale.Divisor[3])
	gps5.Speed3D = float64(int32(binary.BigEndian.Uint32(bytes[16:20]))) / float64(scale.Divisor[4])

	// No error
	return nil
}

// Parse (GPSF) - Parse byte slice into GPSF struct
func (gpsf *GPSF) Parse(bytes []byte) error {
	// Check length
	if 4 != len(bytes) {
		return errors.New("GPSF: Invalid packet length")
	}

	// GPS fix
	gpsf.Fix = binary.BigEndian.Uint32(bytes[0:4])

	// No error
	return nil
}

// Parse (GPSP) - Parse byte slice into GPSP struct
func (gpsp *GPSP) Parse(bytes []byte) error {
	// Check length
	if 2 != len(bytes) {
		return errors.New("GPSP: Invalid packet length")
	}

	// GPS accuracy
	gpsp.Accuracy = binary.BigEndian.Uint16(bytes[0:2])

	// No error
	return nil
}

// Parse (GPSU) - Parse byte slice int GPSU struct
func (gpsu *GPSU) Parse(bytes []byte) error {
	// Check length
	if 16 != len(bytes) {
		return errors.New("GPSU: Invalid packet length")
	}

	// Parse timestamp ("06": Year, "01": Zero Month, "02": Zero Day, "15": Hour, "04": Zero Minute, "05": Zero Second)
	t, err := time.Parse("060102150405", string(bytes))
	if err != nil {
		return err
	}

	// GPS timestamp
	gpsu.Time = t

	// No error
	return nil
}

// Parse (TMPC) - Parse byte slice int TMPC struct
func (tmpc *TMPC) Parse(bytes []byte) error {
	// Check length
	if 4 != len(bytes) {
		return errors.New("TMPC: Invalid packet length")
	}

	// Extract bits
	bits := binary.BigEndian.Uint32(bytes[0:4])

	// Extract 32 bit float
	tmpc.Temp = math.Float32frombits(bits)

	// No error
	return nil
}

// Parse (TSMP) - Parse byte slice int TSMP struct
func (tsmp *TSMP) Parse(bytes []byte) error {
	// Check length
	if 4 != len(bytes) {
		return errors.New("Invalid length TSMP packet")
	}

	// Total number of sample
	tsmp.Samples = binary.BigEndian.Uint32(bytes[0:4])

	// No error
	return nil
}
