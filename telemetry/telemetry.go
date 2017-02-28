package telemetry

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/paulmach/go.geo"
)

// Represents one second of telemetry data
type TELEM struct {
	Accl        []ACCL
	Gps         []GPS5
	Gyro        []GYRO
	GpsFix      GPSF
	GpsAccuracy GPSP
	Time        GPSU
	Temp        TMPC
}

// the thing we want, json-wise
// GPS data might have a generated timestamp and derived track
type TELEM_OUT struct {
	*GPS5

	GpsAccuracy uint16  `json:"gps_accuracy,omitempty"`
	GpsFix      uint32  `json:"gps_fix,omitempty"`
	Temp        float32 `json:"temp,omitempty"`
	Track       float64 `json:"track,omitempty"`
}

var pp *geo.Point = geo.NewPoint(10, 10)
var last_good_track float64 = 0

// zeroes out the telem struct
func (t *TELEM) Clear() {
	t.Accl = t.Accl[:0]
	t.Gps = t.Gps[:0]
	t.Gyro = t.Gyro[:0]
	t.Time.Time = time.Time{}
}

// determines if the telem has data
func (t *TELEM) IsZero() bool {
	// hack.
	return t.Time.Time.IsZero()
}

// try to populate a timestamp for every GPS row. probably bogus.
func (t *TELEM) Process(until time.Time) error {
	len := len(t.Gps)
	diff := until.Sub(t.Time.Time)

	offset := diff.Seconds() / float64(len)

	for i, _ := range t.Gps {
		dur := time.Duration(float64(i)*offset*1000) * time.Millisecond
		ts := t.Time.Time.Add(dur)
		t.Gps[i].TS = ts.UnixNano() / 1000
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
		jobj.Track = pp.BearingTo(p)
		pp = p

		if jobj.Track < 0 {
			jobj.Track = 360 + jobj.Track
		}

		// only set the track if speed is over 1 m/s
		// if it's slower (eg, stopped) it will drift all over with the location
		if tp.Speed > 1 {
			last_good_track = jobj.Track
		} else {
			jobj.Track = last_good_track
		}

		jstr, err := json.Marshal(jobj)
		if err != nil {
			return buffer, errors.New("error jsoning\n")
		}

		buffer.Write(jstr)

	}

	return buffer, nil
}
