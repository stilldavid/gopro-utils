// Extracts stats from GoPro Metadata
/* Essential to have:
	- Mean (! (vf-vi)/(tf-ti) ) acceleration on x, y and z DONE
	- Distance travelled DONE
	- Mean speed DONE
	- Mean altitude DONE
	- Peak altitude DONE
	- Peak speed DONE
*/

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
    "log"
	"math"
	"strings"
	"time"
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
	seconds := -1
	//Arrays
	
	Gyroscope_X := []float64{}
	Gyroscope_Y := []float64{}
	Gyroscope_Z := []float64{}
	Speed := []float64{}
	Altitude := []float64{}
	Latitude := []float64{}
	Longitude := []float64{}
	
	Accel_X := []float64{}
	Accel_Y := []float64{}
	Accel_Z := []float64{}
	Epochtime := []string{}
	BiggestSpeed, BiggestAltitude := 0.0, 0.0
	BiggestSpeedTime, BiggestAltitudeTime := "", ""
	Milliseconds := 0.0
	TotalDistance := 0.0
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

		for i, _ := range t.Accl {
	    	//milliseconds := float64(seconds*1000)+float64(((float64(1000)/float64(len(t.Accl)))*float64(i)))
			Accel_X = append(Accel_X, t.Accl[i].X)
			Accel_Y = append(Accel_Y, t.Accl[i].Y)
			Accel_Z = append(Accel_Z, t.Accl[i].Z)
			Milliseconds = (float64(seconds*1000)+float64(((float64(1000)/float64(len(t.Accl)))*float64(i))))/1000
		}
		
	    for i, _ := range t.Gyro {
			Gyroscope_X = append(Gyroscope_X, t.Gyro[i].X)
			Gyroscope_Y = append(Gyroscope_Y, t.Gyro[i].Y)
			Gyroscope_Z = append(Gyroscope_Z, t.Gyro[i].Z)
		}	
		t.FillTimes(t.Time.Time)
		
		for i, _ := range t.Gps {
			Latitude = append(Latitude, t.Gps[i].Latitude)
			Longitude = append(Longitude, t.Gps[i].Longitude)
			Speed = append(Speed, t.Gps[i].Speed)
			if t.Gps[i].Speed > BiggestSpeed{
				BiggestSpeed = t.Gps[i].Speed
				BiggestSpeedTime = int64tostr(t.Gps[i].TS)
			}
			Altitude = append(Altitude, t.Gps[i].Altitude)
			if t.Gps[i].Altitude > BiggestAltitude{
				BiggestAltitude = t.Gps[i].Altitude
				BiggestAltitudeTime = int64tostr(t.Gps[i].TS)
			}
			Epochtime = append(Epochtime, int64tostr(t.Gps[i].TS))
			
		}
		t = &telemetry.TELEM{}
		seconds++
	}
	for i, _ := range Latitude{
		if i < len(Latitude) - 1{
			TotalDistance = TotalDistance + Distance(Latitude[i], Longitude[i], Latitude[i+1], Longitude[i+1])
		}
	}
	
			 
	//FirstTimeRecord = Epochtime[0]
	var total_accel_x, total_accel_y, total_accel_z, speed, altitude float64 = 0, 0, 0, 0, 0
	for _, value:= range Accel_X {
		total_accel_x += value
	}
	for _, value:= range Accel_Y {
		total_accel_y += value
	}
	for _, value:= range Accel_Z {
		total_accel_z += value
	}
	for _, value:= range Speed {
		speed += value
	}
	for _, value:= range Altitude {
		altitude += value
	}
	fmt.Println("Data from " + *inName)
	fmt.Println("\nAcceleration:\n\tMean on X axis: " + floattostr(Round(total_accel_x/float64(len(Accel_X)), .5, 3)))
	fmt.Println("\tMean on Y axis: " + floattostr(Round(total_accel_y/float64(len(Accel_Y)), .5, 3)))
	fmt.Println("\tMean on Z axis: " + floattostr(Round(total_accel_z/float64(len(Accel_Z)), .5, 3)))
	fmt.Println("\nSpeed:\n\tAverage speed: " + floattostr(Round(speed/float64(len(Speed)), .5, 3)) + " m/s")
	fmt.Println("\tPeak speed: " + floattostr(getBiggestFromSlice(Speed)) + " m/s" + "\n\t\tat " + getUTCTimeFromUnix(BiggestSpeedTime) + "\n\t\tin video: " + getDifferenceBetweenDates(getUTCTimeFromUnix(Epochtime[0]), getUTCTimeFromUnix(BiggestSpeedTime)))
	fmt.Println("\nAltitude:\n\tAverage altitude: " + floattostr(Round(altitude/float64(len(Altitude)), .5, 3)) + " meters")
	fmt.Println("\tPeak altitude: " + floattostr(getBiggestFromSlice(Altitude)) + " meters" + "\n\t\tat: " + getUTCTimeFromUnix(BiggestAltitudeTime) + "\n\t\tin video: " + getDifferenceBetweenDates(getUTCTimeFromUnix(Epochtime[0]), getUTCTimeFromUnix(BiggestAltitudeTime)))
	fmt.Println("\nDistance travelled: " + "\n\ts * t (for short distances): " + floattostr( Round( Milliseconds * (Round(speed/float64(len(Speed)), .5, 3)), .5, 3)) + " meters")
	fmt.Println("\tUsing the Haversine formula: " + floattostr( Round(TotalDistance, .5, 3 )) + " meters")
}

func floattostr(input_num float64) string {

        // to convert a float number to a string
    return strconv.FormatFloat(input_num, 'f', -1, 64)
}

func getDifferenceBetweenDates(date_1 string, date_2 string) string {
	var year, _ = strconv.Atoi(date_1[0:4])
	var month, _ = strconv.Atoi(date_1[5:7])
	var day, _ = strconv.Atoi(date_1[8:10])
	var hour, _ = strconv.Atoi(date_1[11:13])
	var minute, _ = strconv.Atoi(date_1[13:16])
	var second, _ = strconv.Atoi(date_1[17:19])
	var decimals, _ = strconv.Atoi(date_1[31:35])
	start := time.Date(
		year, time.Month(month), day, hour, minute, second, decimals, time.UTC)
	var year_2, _ = strconv.Atoi(date_2[0:4])
	var month_2, _ = strconv.Atoi(date_2[5:7])
	var day_2, _ = strconv.Atoi(date_2[8:10])
	var hour_2, _ = strconv.Atoi(date_2[11:13])
	var minute_2, _ = strconv.Atoi(date_2[13:16])
	var second_2, _ = strconv.Atoi(date_2[17:19])
	var decimals_2, _ = strconv.Atoi(date_2[31:35])
    big := time.Date(
        year_2, time.Month(month_2), day_2, hour_2, minute_2, second_2, decimals_2, time.UTC)
    diff := start.Sub(big)
	final := strconv.Itoa(int(diff.Hours())) + "h " + strconv.Itoa(int(diff.Minutes())) + "m " + strconv.Itoa(int(diff.Seconds())) + "s" 
	return final
}
func getUTCTimeFromUnix(timestamp string) string {
	i, err := strconv.ParseInt(timestamp[0:10], 10, 64)
    if err != nil {
        panic(err)
    }
    tm := time.Unix(i, 0).String()
	return tm + "-" + timestamp[11:16]
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
func Round(val float64, roundOn float64, places int ) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func Ms2hms(ms float64) string {
	xstr := floattostr(ms)
	i := strings.Index(xstr, ".")
	decimal := xstr[i+1:]
	if len(xstr[i+1:]) < 3{
		decimal = xstr[i+1:] + "0"
	}
	x := ms
	seconds := int(ms)
	x = x / 60
	minutes := int(math.Mod(x, 60))
	x = x / 60
	hours := int(math.Mod(x, 24))
	s := ""
	m := ""
	h := ""
	if len(strconv.Itoa(hours)) == 1{
		h = "0"
	} 
	if len(strconv.Itoa(minutes)) == 1{
		m = "0"
	}
	if len(strconv.Itoa(seconds)) == 1{
		s = "0"
	}
	return h + strconv.Itoa(hours) + ":" + m + strconv.Itoa(minutes) + ":" + s + strconv.Itoa(seconds) + "." + decimal
}
func getBiggestFromSlice(slice []float64) float64 {
	var n, biggest float64
	for _,v:=range slice {
    if v>n {
      n = v
      biggest = n
    }
	}
	return biggest
}
func getSmallestFromSlice(slice []float64) float64 {
	var n, smallest float64
	for _,v:=range slice {
    if v<n {
      n = v
      smallest = n
    }
	}
	return smallest
}
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}