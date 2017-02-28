package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/stilldavid/gopro-utils/telemetry"
)

func main() {
	inName := flag.String("i", "", "Required: telemetry file to read")
	flag.Parse()

	if *inName == "" {
		flag.Usage()
		return
	}

	telemFile, err := os.Open(*inName)
	if err != nil {
		fmt.Printf("Cannot access telemetry file %s.\n", *inName)
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Cannot close file %s: %s", file.Name(), err)
			os.Exit(1)
		}
	}(telemFile)

	// currently processing sentence
	t := &telemetry.TELEM{}
	// previous sentence - keep it around mostly for the timestamp
	t_prev := &telemetry.TELEM{}

	for {
		t, err = telemetry.Read(telemFile)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			os.Exit(1)
		}

		if t == nil {
			break
		}

		// first full, guess it's about a second
		if t_prev.IsZero() {
			*t_prev = *t
			t.Clear()
			continue
		}

		// process the previous timestamp until current known timestamp
		t_prev.FillTimes(t.Time.Time)

		// this is pretty useless: change it to pick a field you want
		// or mangle it to your wishes
		fmt.Println(t_prev.Time)

		*t_prev = *t
		t = &telemetry.TELEM{}
	}
}
