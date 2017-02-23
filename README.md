GoPro Metadata Format Utilities
===============================

TLDR:

1. `ffmpeg -y -i GOPR0001.MP4 -codec copy -map 0:3 -f rawvideo GOPR0001.bin`
2. `gopro2json -i GOPR0001.bin -o GOPR0001.json`
3. There is no step 3

---

**This is bad code and you probably shouldn't use it.** I wrote it in a hurry, and I don't really know Go all that well. Consider the API *UNSTABLE* and this all to be very alpha and without tests yet.

---

I spent some time trying to reverse-engineer the GoPro Metadata Format (GPMD or GPMDF) that is stored in GoPro Hero 5 cameras if GPS is enabled. This is what I found.

Part of this code _is_ in production on [Earthscape](https://public.earthscape.com/); for an example of what you can do with the extracted data, see [this video](https://public.earthscape.com/videos/10231).


Getting The Data
----------------

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

TODO: document the format...