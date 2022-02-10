# go-mpeg2-ts
(very experimental) MPEG2-TS parser in pure Go

## example

see under [example](./example/) directory

```go
	mpeg2, err := mpeg2ts.LoadStandardTS("../files/test.ts")
	if err != nil {
		panic(err)
	}

    elementaryPID := uint16(0x0041)
	pesPackets := mpeg2.FilterByPIDs(elementaryPID)
	pesParser := mpeg2ts.NewPESParser(8 * 1048576)
	pesChan := pesParser.StartPESReadLoop()
	go func() {
		i := 0
		for {
			p := <-pesChan
			go func(index int, pes mpeg2ts.PES) {
				fmt.Printf("ES frame: %dbytes\n", len(p.ElementaryStream))
				fname := fmt.Sprintf("es_%04d.bin", i)
				os.WriteFile(fname, p.ElementaryStream, 0644)
			}(i, p)
			i++
		}
	}()
    
	for _, p := range pesPackets.PacketList.All() {
		eop := false
		if len(p.AdaptationField.Stuffing) > 0 {
			eop = true
		}
		err = pesParser.EnqueueTSPacket(p, eop)
		if err != nil {
			panic(err)
		}
	}
```

## generate test video
```bash
$ timeout -s INT 5 gst-launch-1.0 videotestsrc ! x264enc ! progressreport ! mpegtsmux ! filesink location=test.ts
Setting pipeline to PAUSED ...
Pipeline is PREROLLING ...
Redistribute latency...
Pipeline is PREROLLED ...
Setting pipeline to PLAYING ...
New clock: GstSystemClock
progressreport0 (00:00:05): 74 seconds
handling interrupt.
Interrupt: Stopping pipeline ...
Execution ended after 0:00:04.884230197
Setting pipeline to PAUSED ...
Setting pipeline to READY ...
Setting pipeline to NULL ...
Freeing pipeline ...
$ ffprobe -hide_banner test.ts 
Input #0, mpegts, from 'test.ts':
  Duration: 00:01:27.50, start: 3600.000000, bitrate: 1576 kb/s
  Program 1 
    Stream #0:0[0x41]: Video: h264 (High 4:4:4 Predictive) (HDMV / 0x564D4448), yuv444p(tv, bt470bg/smpte170m/bt709, progressive), 320x240 [SAR 1:1 DAR 4:3], 30 fps, 30 tbr, 90k tbn, 60 tbc
```