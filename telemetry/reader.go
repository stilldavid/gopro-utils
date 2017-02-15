package telemetry

import (
	_ "encoding/binary"
	"fmt"
	"io"
	"os"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Read(f *os.File) *TELEM {
	labels := []string{
		"ACCL",
		"DEVC",
		"DVID",
		"DVNM",
		"EMPT",
		"GPS5",
		"GPSF",
		"GPSP",
		"GPSU",
		"GYRO",
		"SCAL",
		"SIUN",
		"STRM",
		"TMPC",
		"TSMP",
		"UNIT",
	}

	label := make([]byte, 4, 4) // 4 byte ascii label of data
	desc := make([]byte, 4, 4)  // 4 byte description of length of data

	// keep a copy of the scale to apply to subsequent sentences
	s := SCAL{}

	// keep a copy of the full telemetry
	t := &TELEM{}

	for {
		// pick out the label
		read, err := f.Read(label)
		if err == io.EOF || read == 0 {
			break
		}

		if !stringInSlice(string(label), labels) {
			fmt.Printf("Could not find label in list: %s (%x)\n", label, label)
			break
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
		if "EMPT" == string(label) {
			_, err = f.Seek(4, 1)
			if err != nil {
				fmt.Println(err)
				break
			}
			continue
		}

		// extract the size and length
		val_size := int64(desc[1])
		num_values := (int64(desc[2]) << 8) | int64(desc[3])
		length := val_size * num_values

		//fmt.Printf("%s (%c): %v entries of length %v\n", label, desc[0], num_values, val_size)

		if "SCAL" == string(label) {
			value := make([]byte, val_size*num_values, val_size*num_values)
			read, err = f.Read(value)
			if err == io.EOF || read == 0 {
				fmt.Printf("error reading file\n")
				break
			}

			// clear the scales
			s.Values = s.Values[:0]

			err := s.Parse(value, val_size)
			if err != nil {
				fmt.Println(err)
				break
			}
		} else {
			value := make([]byte, val_size, val_size)

			for i := int64(0); i < num_values; i++ {
				read, err := f.Read(value)
				if err == io.EOF || read == 0 {
					fmt.Printf("error reading file\n")
					break
				}

				label_string := string(label)

				// I think DVID is the payload boundary.
				if "DVID" == label_string {

					return t

				} else if "GPS5" == label_string {
					g := GPS5{}
					g.Parse(value, &s)
					t.Gps = append(t.Gps, g)
				} else if "GPSU" == label_string {
					g := GPSU{}
					err := g.Parse(value)
					if err != nil {
						fmt.Println(err)
						break
					}
					t.Time = g
				} else if "ACCL" == label_string {
					a := ACCL{}
					err := a.Parse(value, &s)
					if err != nil {
						fmt.Println(err)
						break
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
						fmt.Println(err)
						break
					}
					t.Gyro = append(t.Gyro, g)
				} else if "GPSP" == label_string {
					g := GPSP{}
					err := g.Parse(value)
					if err != nil {
						fmt.Println(err)
						break
					}
					t.GpsAccuracy = g
				} else if "GPSF" == label_string {
					g := GPSF{}
					err := g.Parse(value)
					if err != nil {
						fmt.Println(err)
						break
					}
					t.GpsFix = g
				} else if "UNIT" == label_string {
					//fmt.Printf("\tvals: %s\n", value)
				} else if "SIUN" == label_string {
					//fmt.Printf("\tvals: %s\n", value)
				} else if "DVNM" == label_string {
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
			_, err = f.Seek(seek, 1)
			if err != nil {
				break
			}
		}
	}

	return nil
}
