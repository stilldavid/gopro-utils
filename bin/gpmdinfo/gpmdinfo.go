package main

import (
	_ "encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/stilldavid/gopro-utils/telemetry"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func main() {
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

	inName := flag.String("i", "", "Required: telemetry file to read")
	flag.Parse()

	if *inName == "" {
		flag.Usage()
		return
	}

	telemFile, err := os.Open(*inName)
	if err != nil {
		fmt.Println("Cannot access telemetry file %s.\n", *inName)
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Cannot close file", file.Name(), err)
		}
	}(telemFile)

	label := make([]byte, 4, 4) // 4 byte ascii label of data
	desc := make([]byte, 4, 4)  // 4 byte description of length of data

	// keep a copy of the scale to apply to subsequent sentences
	s := telemetry.SCAL{}

	// keep a copy of the full telemetry
	t := &telemetry.TELEM{}
	t_prev := &telemetry.TELEM{}

	fmt.Println(`insert into flight_point (utc, acft_alt, acft_hdg, speed, acft, video_id) values`)

	for {
		// pick out the label
		read, err := telemFile.Read(label)
		if err == io.EOF || read == 0 {
			break
		}

		if !stringInSlice(string(label), labels) {
			fmt.Printf("Could not find label in list: %s (%x)\n", label, label)
			break
		}

		// pick out the label description
		read, err = telemFile.Read(desc)
		if err == io.EOF || read == 0 {
			break
		}

		// first byte is zero, there is no length
		if 0x0 == desc[0] {
			/*
				if "STRM" == string(label) {
					thing := binary.BigEndian.Uint16(desc[1:4])
					fmt.Printf("%s: %v\n", label, thing)
				} else if "DEVC" == string(label) {
					thing := binary.BigEndian.Uint16(desc[1:4])
					fmt.Printf("%s: %v\n", label, thing)
				} else {
					fmt.Printf("%s: %x\n", label, desc)
				}
			*/
			continue
		}

		// skip empty packets
		if "EMPT" == string(label) {
			_, err = telemFile.Seek(4, 1)
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

		fmt.Printf("%s (%c): %v entries of length %v\n", label, desc[0], num_values, val_size)

		if "SCAL" == string(label) {
			value := make([]byte, val_size*num_values, val_size*num_values)
			read, err = telemFile.Read(value)
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
				read, err := telemFile.Read(value)
				if err == io.EOF || read == 0 {
					fmt.Printf("error reading file\n")
					break
				}

				label_string := string(label)

				// I think DVID is the payload boundary.
				if "DVID" == label_string {
					// first full, guess it's about a second
					if t_prev.IsZero() {
						*t_prev = *t
						t.Clear()
						fmt.Println("zero! moving on...")
						continue
					}

					// process until t.Time
					t_prev.Process(t.Time.Time)

					*t_prev = *t
					t = &telemetry.TELEM{}

				} else if "GPS5" == label_string {
					g := telemetry.GPS5{}
					g.Parse(value, &s)
					t.Gps = append(t.Gps, g)
					//fmt.Printf("appending %v\n", g)

				} else if "GPSU" == label_string {
					gu := telemetry.GPSU{}
					err := gu.Parse(value)
					if err != nil {
						fmt.Println(err)
						break
					}
					t.Time = gu

					fmt.Printf("\ttime: %v\n", gu)
				} else if "ACCL" == label_string {
					a := telemetry.ACCL{}
					err := a.Parse(value, &s)
					if err != nil {
						fmt.Println(err)
						break
					}
					t.Accl = append(t.Accl, a)
				} else if "TMPC" == label_string {
					tmp := telemetry.TMPC{}
					tmp.Parse(value)
					t.Temp = tmp
				} else if "TSMP" == label_string {
					tsmp := telemetry.TSMP{}
					tsmp.Parse(value, &s)
				} else if "GYRO" == label_string {
					g := telemetry.GYRO{}
					err := g.Parse(value, &s)
					if err != nil {
						fmt.Println(err)
						break
					}
					t.Gyro = append(t.Gyro, g)
				} else if "GPSP" == label_string {
					g := telemetry.GPSP{}
					err := g.Parse(value)
					if err != nil {
						fmt.Println(err)
						break
					}
					t.GpsAccuracy = g
				} else if "GPSF" == label_string {
					g := telemetry.GPSF{}
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
					fmt.Printf("\tvalue is %v\n", value)
				}
			}
		}

		// pack into 4 bytes
		mod := length % 4
		if mod != 0 {
			seek := 4 - mod
			_, err = telemFile.Seek(seek, 1)
			if err != nil {
				break
			}
		}
	}
}
