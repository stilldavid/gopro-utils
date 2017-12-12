package telemetry

import (
	"fmt"
	"io"
	"io/ioutil"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Read(f io.Reader) (*TELEM, error) {
	labels := []string{
		"ACCL",
		"DEVC",
		"DVID",
		"DVNM",
		"EMPT",
		"GPRO",
		"GPS5",
		"GPSF",
		"GPSP",
		"GPSU",
		"GYRO",
		"HD5.",
		"ISOG",
		"SCAL",
		"SHUT",
		"SIUN",
		"STNM",
		"STRM",
		"TICK",
		"TMPC",
		"TSMP",
		"UNIT",
		"TYPE",
		"FACE",
		"FCNM",
		"ISOE",
		"WBAL",
		"WRGB",
		"MAGN",
		"ALLD",
	}

	label := make([]byte, 4, 4) // 4 byte ascii label of data
	desc := make([]byte, 4, 4)  // 4 byte description of length of data

	// keep a copy of the scale to apply to subsequent sentences
	s := SCAL{}

	// the full telemetry for this period
	t := &TELEM{}

	for {
		// pick out the label
		read, err := f.Read(label)
		if err == io.EOF || read == 0 {
			return nil, err
		}

		label_string := string(label)

		if !stringInSlice(label_string, labels) {
			err := fmt.Errorf("Could not find label in list: %s (%x)\n", label, label)
			return nil, err
		}

		// pick out the label description
		read, err = f.Read(desc)
		if err == io.EOF || read == 0 {
			break
		}

		// first byte is zero, there is no length
		if 0x0 == desc[0] {
			continue
		}

		// skip empty packets
		if "EMPT" == label_string {
			io.CopyN(ioutil.Discard, f, 4)
			continue
		}

		// extract the size and length
		val_size := int64(desc[1])
		num_values := (int64(desc[2]) << 8) | int64(desc[3])
		length := val_size * num_values

		// uncomment to see label, type, size and length
		//fmt.Printf("%s (%c) of size %v and len %v\n", label, desc[0], val_size, length)

		if "SCAL" == label_string {
			value := make([]byte, val_size*num_values, val_size*num_values)
			read, err = f.Read(value)
			if err == io.EOF || read == 0 {
				return nil, err
			}

			// clear the scales
			s.Values = s.Values[:0]

			err := s.Parse(value, val_size)
			if err != nil {
				return nil, err
			}
		} else {
			value := make([]byte, val_size)

			for i := int64(0); i < num_values; i++ {
				read, err := f.Read(value)
				if err == io.EOF || read == 0 {
					return nil, err
				}

				// I think DVID is the payload boundary; this might be a bad assumption
				if "DVID" == label_string {

					// XXX: I think this might skip the first sentence
					return t, nil
				} else if "GPS5" == label_string {
					g := GPS5{}
					g.Parse(value, &s)
					t.Gps = append(t.Gps, g)
				} else if "GPSU" == label_string {
					g := GPSU{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.Time = g
				} else if "ACCL" == label_string {
					a := ACCL{}
					err := a.Parse(value, &s)
					if err != nil {
						return nil, err
					}
					t.Accl = append(t.Accl, a)
				} else if "TMPC" == label_string {
					tmp := TMPC{}
					tmp.Parse(value)
					t.Temp = tmp
				} else if "TSMP" == label_string {
					tsmp := TSMP{}
					tsmp.Parse(value, &s)
				} else if "GYRO" == label_string {
					g := GYRO{}
					err := g.Parse(value, &s)
					if err != nil {
						return nil, err
					}
					t.Gyro = append(t.Gyro, g)
				} else if "GPSP" == label_string {
					g := GPSP{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.GpsAccuracy = g
				} else if "GPSF" == label_string {
					g := GPSF{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.GpsFix = g
				} else if "UNIT" == label_string {
					// this is a string of units like "rad/s", not sure if it changes
					//fmt.Printf("\tvals: %s\n", value)
				} else if "SIUN" == label_string {
					// this is the SI unit - also not sure if it changes
					//fmt.Printf("\tvals: %s\n", value)
				} else if "DVNM" == label_string {
					// device name, "Camera"
					//fmt.Printf("\tvals: %s\n", value)
				} else {
					//fmt.Printf("\tvalue is %v\n", value)
				}
			}
		}

		// pack into 4 bytes
		mod := length % 4
		if mod != 0 {
			seek := 4 - mod
			io.CopyN(ioutil.Discard, f, seek)
		}
	}

	return nil, nil
}
