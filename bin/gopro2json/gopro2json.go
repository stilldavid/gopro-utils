package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/stilldavid/gopro-utils/telemetry"
)

func main() {
	inName := flag.String("i", "", "Required: telemetry file to read")
	outName := flag.String("o", "", "Required: json file to write")
	flag.Parse()

	if *inName == "" {
		flag.Usage()
		return
	}

	telemFile, err := os.Open(*inName)
	if err != nil {
		fmt.Println("Cannot access telemetry file %s.\n", *inName)
		os.Exit(1)
	}

	jsonFile, err := os.Create(*outName)
	if err != nil {
		fmt.Println("Cannot make output file %s.\n", *outName)
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Cannot close file", file.Name(), err)
			os.Exit(1)
		}
	}(telemFile)

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Cannot close file", file.Name(), err)
			os.Exit(1)
		}
	}(jsonFile)

	jsonFile.WriteString(`{"data":[`)

	t := &telemetry.TELEM{}
	t_prev := &telemetry.TELEM{}
	first := true

	for {
		t = telemetry.Read(telemFile)
		if t == nil {
			break
		}

		// first full, guess it's about a second
		if t_prev.IsZero() {
			*t_prev = *t
			t.Clear()
			continue
		}

		// process until t.Time
		t_prev.Process(t.Time.Time)

		buf, _ := t_prev.ShitJson(first)
		jsonFile.WriteString(buf.String())
		if first {
			first = false
		}

		*t_prev = *t
		t = &telemetry.TELEM{}
	}

	// end the bogusness
	jsonFile.WriteString(`]}`)
}
