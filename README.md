Fork of stilldavids repo to allow for GoPro 8 files

GoPro Metadata Format Parser
============================

TLDR:

1. `ffmpeg -y -i GOPR0001.MP4 -codec copy -map 0:m:handler_name:"	GoPro MET" -f rawvideo GOPR0001.bin`
2. `gopro2json -i GOPR0001.bin -o GOPR0001.json`
3. There is no step 3

---

I spent some time trying to reverse-engineer the GoPro Metadata Format (GPMD or GPMDF) that is stored in GoPro Hero 5 cameras if GPS is enabled. This is what I found.

Extracting the Metadata File
----------------------------

The metadata stream is stored in the `.mp4` video file itself alongside the video and audio streams. We can use `ffprobe` to find it:

```
[computar][100GOPRO] ➔ ffprobe GOPR0008.MP4
ffprobe version 3.2.4 Copyright (c) 2007-2017 the FFmpeg developers
[SNIP]
    Stream #0:3(eng): Data: none (gpmd / 0x646D7067), 33 kb/s (default)
    Metadata:
      creation_time   : 2016-11-22T23:42:41.000000Z
      handler_name    : 	GoPro MET
[SNIP]
```

We can identify it by the `gpmd` in the tag string - in this case it's id 3. We can then use `ffmpeg` to extract the metadata stream into a binary file for processing:

`ffmpeg -y -i GOPR0001.MP4 -codec copy -map 0:3 -f rawvideo out-0001.bin`

This leaves us with a binary file with the data.

Data We Get
-----------

* ~400 Hz 3-axis gyro readings
* ~200 Hz 3-axis accelerometer readings
* ~18 Hz GPS position (lat/lon/alt/spd)
* 1 Hz GPS timestamps
* 1 Hz GPS accuracy (cm) and fix (2d/3d)
* 1 Hz temperature of camera

---


The Protocol
------------

Data starts with a label that describes the data following it. Values are all big endian, and floats are IEEE 754. Everything is packed to 4 bytes where applicable, padded with zeroes so it's 32-bit aligned.

 * **Labels** - human readable types of proceeding data
 * **Type** - single ascii character describing data
 * **Size** - how big is the data type
 * **Count** - how many values are we going to get
 * **Length** = size * count

Labels include:

 * `ACCL` - accelerometer reading x/y/z
 * `DEVC` - device 
 * `DVID` - device ID, possibly hard-coded to 0x1
 * `DVNM` - devicde name, string "Camera"
 * `EMPT` - empty packet
 * `GPS5` - GPS data (lat, lon, alt, speed, 3d speed)
 * `GPSF` - GPS fix (none, 2d, 3d)
 * `GPSP` - GPS positional accuracy in cm
 * `GPSU` - GPS acquired timestamp; potentially different than "camera time"
 * `GYRO` - gryroscope reading x/y/z
 * `SCAL` - scale factor, a multiplier for subsequent data
 * `SIUN` - SI units; strings (m/s², rad/s)
 * `STRM` - ¯\\\_(ツ)\_/¯
 * `TMPC` - temperature
 * `TSMP` - total number of samples
 * `UNIT` - alternative units; strings (deg, m, m/s)

Types include:

 * `c` - single char
 * `L` - unsigned long
 * `s` - signed short
 * `S` - unsigned short
 * `f` - 32 float

For implementation details, see `reader.go` and other corresponding files in `telemetry/`.
