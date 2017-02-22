package main

import (
	"flag"
	"fmt"
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
		fmt.Println("Cannot access telemetry file %s.\n", *inName)
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Cannot close file", file.Name(), err)
			os.Exit(1)
		}
	}(telemFile)

	// currently processing sentence
	t := &telemetry.TELEM{}
	// previous sentence - keep it around mostly for the timestamp
	t_prev := &telemetry.TELEM{}

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

		// process the previous timestamp until current known timestamp
		t_prev.Process(t.Time.Time)

		// this is pretty useless: change it to pick a field you want
		fmt.Println(t_prev.Time)

		*t_prev = *t
		t = &telemetry.TELEM{}
	}
}
