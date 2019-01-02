package main

// Credit: @Kajuna || https://community.gopro.com/t5/Hero5-Metadata-Visualisation/Extracting-the-metadata-in-a-useful-format/gpm-p/40293
import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/stilldavid/gopro-utils/telemetry"

	//////used for csv
	"strconv"
    "log"
    "encoding/csv"

)


func main() {
	///////////////////////////////////////////////////////////////////////////////////////////csv
	////////////////////accelerometer
	var acclCsv = [][]string{{"Milliseconds","AcclX","AcclY","AcclZ"}}
	acclFile, err := os.Create("accl.csv")
    checkError("Cannot create accl.csv file", err)
    defer acclFile.Close()
    acclWriter := csv.NewWriter(acclFile)
    /////////////////////gyroscope
    var gyroCsv = [][]string{{"Milliseconds","GyroX","GyroY","GyroZ"}}
	gyroFile, err := os.Create("gyro.csv")
    checkError("Cannot create gyro.csv file", err)
    defer gyroFile.Close()
    gyroWriter := csv.NewWriter(gyroFile)
    //////////////////////temperature
    var tempCsv = [][]string{{"Milliseconds","Temp"}}
	tempFile, err := os.Create("temp.csv")
    checkError("Cannot create temp.csv file", err)
    defer tempFile.Close()
    tempWriter := csv.NewWriter(tempFile)
    ///////////////////////Uncomment for Gps
    
    var gpsCsv = [][]string{{"Latitude","Longitude","Altitude","Speed","Speed3D","TS"}}
	gpsFile, err := os.Create("gps.csv")
    checkError("Cannot create gps.csv file", err)
    defer gpsFile.Close()
    gpsWriter := csv.NewWriter(gpsFile)
   
    //////////////////////////////////////////////////////////////////////////////////////////////

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

	seconds := -1


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

		// this is pretty useless and info overload: change it to pick a field you want
		// or mangle it to your wishes into JSON/CSV/format of choice
		//fmt.Println(t)

		///////////////////////////////////////////////////////////////////Modified to save CSV
		/////////////////////Accelerometer
	    for i, _ := range t.Accl {
	    	milliseconds := float64(seconds*1000)+float64(((float64(1000)/float64(len(t.Accl)))*float64(i)))
			acclCsv = append(acclCsv, []string{floattostr(milliseconds),floattostr(t.Accl[i].X),floattostr(t.Accl[i].Y),floattostr(t.Accl[i].Z)})
		}
		/////////////////////Gyroscope
	    for i, _ := range t.Gyro {
	    	milliseconds := float64(seconds*1000)+float64(((float64(1000)/float64(len(t.Gyro)))*float64(i)))
			gyroCsv = append(gyroCsv, []string{floattostr(milliseconds),floattostr(t.Gyro[i].X),floattostr(t.Gyro[i].Y),floattostr(t.Gyro[i].Z)})
		}
		////////////////////Temperature
		milliseconds := seconds*1000
		tempCsv = append(tempCsv, []string{strconv.Itoa(milliseconds),floattostr(float64(t.Temp.Temp))})
		////////////////////Uncomment for Gps
		
		for i, _ := range t.Gps {
			gpsCsv = append(gpsCsv, []string{floattostr(t.Gps[i].Latitude),floattostr(t.Gps[i].Longitude),floattostr(t.Gps[i].Altitude),floattostr(t.Gps[i].Speed),floattostr(t.Gps[i].Speed3D),int64tostr(t.Gps[i].TS)})
		}
		
	    //////////////////////////////////////////////////////////////////////////////////

		t = &telemetry.TELEM{}
		seconds++
	}
	/////////////////////////////////////////////////////////////////////////////////////for csv
	///////////////accelerometer
	for _, value := range acclCsv {
        err := acclWriter.Write(value)
        checkError("Cannot write to accl.csv file", err)
    }
    defer acclWriter.Flush()
    ///////////////gyroscope
    for _, value := range gyroCsv {
        err := gyroWriter.Write(value)
        checkError("Cannot write to gyro.csv file", err)
    }
    defer gyroWriter.Flush()
    /////////////temperature
    for _, value := range tempCsv {
        err := tempWriter.Write(value)
        checkError("Cannot write to temp.csv file", err)
    }
    defer tempWriter.Flush()
    /////////////Uncomment for Gps
    
    for _, value := range gpsCsv {
        err := gpsWriter.Write(value)
        checkError("Cannot write to gps.csv file", err)
    }
    defer gpsWriter.Flush()
    
    /////////////////////////////////////////////////////////////////////////////////////
}


///////////for csv

func floattostr(input_num float64) string {

        // to convert a float number to a string
    return strconv.FormatFloat(input_num, 'f', -1, 64)
}



func int64tostr(input_num int64) string {

        // to convert a float number to a string
    return strconv.FormatInt(input_num, 10)
}

 func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}

