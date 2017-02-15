package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paulmach/go.geo"
)

// Represents one second of telemetry data
type TELEM struct {
	Accl        []ACCL `json:"-"`
	Gps         []GPS5 `json:"gps"`
	Gyro        []GYRO `json:"-"`
	GpsFix      GPSF   `json:"gps_fix"`
	GpsAccuracy GPSP   `json:"gps_accuracy"`
	Time        GPSU   `json:"-"`
	Temp        TMPC   `json:"temp"`
}

// the thing we want, json-wise
type TELEM_OUT struct {
	*GPS5

	GpsAccuracy uint16  `json:"gps_accuracy,omitempty"`
	GpsFix      uint32  `json:"gps_fix,omitempty"`
	Temp        float32 `json:"temp,omitempty"`
	Heading     float64 `json:"heading,omitempty"`
}

var pp *geo.Point = geo.NewPoint(10, 10)
var last_good_heading float64 = 0

// zeroes out the telem struct
func (t *TELEM) Clear() {
	t.Accl = t.Accl[:0]
	t.Gps = t.Gps[:0]
	t.Gyro = t.Gyro[:0]
	t.Time.Time = time.Time{}
}

// determines if the telem has data
func (t *TELEM) IsZero() bool {
	// hack?
	return t.Time.Time.IsZero()
}

func (t *TELEM) Process(until time.Time) error {
	//fmt.Printf("processing from %v to %v\n", t.Time.Time, until)

	len := len(t.Gps)
	diff := until.Sub(t.Time.Time)

	offset := diff.Seconds() / float64(len)

	for i, _ := range t.Gps {
		dur := time.Duration(float64(i)*offset*1000) * time.Millisecond
		ts := t.Time.Time.Add(dur)
		t.Gps[i].TS = ts.UnixNano() / 1000
		//fmt.Printf("\t%v: %v\ta: %v\ts: %v\tf: %v\n", i, ts, t.Gps[i].Altitude, t.Gps[i].Speed, t.Temp)
	}

	return nil
}

func (t *TELEM) ShitJson(first bool) (bytes.Buffer, error) {
	var buffer bytes.Buffer

	for i, tp := range t.Gps {
		if first && i == 0 {
		} else {
			buffer.Write([]byte(","))
		}

		jobj := TELEM_OUT{&tp, 0, 0, 0, 0}
		if 0 == i {
			jobj.GpsAccuracy = t.GpsAccuracy.Accuracy
			jobj.GpsFix = t.GpsFix.F
			jobj.Temp = t.Temp.Temp
		}

		p := geo.NewPoint(tp.Longitude, tp.Latitude)
		jobj.Heading = pp.BearingTo(p)
		pp = p

		if jobj.Heading < 0 {
			jobj.Heading = 360 + jobj.Heading
		}

		if tp.Speed > 1 {
			last_good_heading = jobj.Heading
		} else {
			jobj.Heading = last_good_heading
		}

		jstr, err := json.Marshal(jobj)
		if err != nil {
			fmt.Printf("error jsoning\n")
			break
		}

		buffer.Write(jstr)

	}
	return buffer, nil
}
