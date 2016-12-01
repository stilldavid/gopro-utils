package main

import (
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

	fileName := flag.String("f", "", "Required: telemetry file to read")
	flag.Parse()

	if *fileName == "" {
		flag.Usage()
		return
	}

	telemFile, err := os.Open(*fileName)
	if err != nil {
		fmt.Println("Cannot access telemetry file %s.\n", fileName)
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

	// blobs always start with a label
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
			//fmt.Printf("%s: %x\n", label, desc)
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
			// fmt.Printf("\tvals: %v\n", s.Values)
		} else {
			value := make([]byte, val_size, val_size)
			for i := int64(0); i < num_values; i++ {
				read, err := telemFile.Read(value)
				if err == io.EOF || read == 0 {
					fmt.Printf("error reading file\n")
					break
				}
				label_string := string(label)
				if "GPS5" == label_string {
					g := telemetry.GPS5{}
					g.Parse(value, &s)
					fmt.Printf("\t%.6f, %.6f, %.2f, %.2f, %.2f\n", g.Latitude, g.Longitude, g.Altitude, g.Speed, g.Speed3D)
				} else if "GPSU" == label_string {
					g := telemetry.GPSU{}
					err := g.Parse(value)
					if err != nil {
						fmt.Println(err)
						break
					}
					fmt.Printf("\t%v\n", g.Time)
				} else if "ACCL" == label_string {
					a := telemetry.ACCL{}
					err := a.Parse(value, &s)
					if err != nil {
						fmt.Println(err)
						break
					}
					//fmt.Printf("\t%v\n", value)
					fmt.Printf("\t%v\t%v\t%v\n", a.X, a.Y, a.Z)
				} else if "TMPC" == label_string {
					t := telemetry.TMPC{}
					t.Parse(value, &s)
					fmt.Printf("\t%.6f\n", t.Temp)
				} else if "TSMP" == label_string {
					t := telemetry.TSMP{}
					t.Parse(value, &s)
					//fmt.Printf("\t%.6f\n", t.Temp)
				} else if "GYRO" == label_string {
					g := telemetry.GYRO{}
					err := g.Parse(value)
					if err != nil {
						fmt.Println(err)
						break
					}
					//fmt.Printf("\t%v\n", value)
					fmt.Printf("\t%v\t%v\t%v\n", g.X, g.Y, g.Z)
				} else if "UNIT" == label_string {
					//fmt.Printf("\tvals: %s\n", value)
				} else if "GPSP" == label_string {
					g := telemetry.GPSP{}
					err := g.Parse(value)
					if err != nil {
						fmt.Println(err)
						break
					}
					//fmt.Printf("\t%v\n", g.Heading)
				} else if "GPSF" == label_string {
					g := telemetry.GPSF{}
					err := g.Parse(value)
					if err != nil {
						fmt.Println(err)
						break
					}
					fmt.Printf("\t%v\n", g.F)
				} else if "UNIT" == label_string {
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
			_, err = telemFile.Seek(seek, 1)
			if err != nil {
				break
			}
		}
	}
}
