package main

import (
	"fmt"
	"log"
	"net"
	"os"

	mpeg2ts "github.com/misodengaku/go-mpeg2-ts"
)

func main() {
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 50000,
	}
	udpConn, _ := net.ListenUDP("udp", udpAddr)

	buf := make([]byte, 1048576)
	log.Println("Starting udp server...")
	tse, _ := mpeg2ts.InitTSEngine(mpeg2ts.PacketSizeDefault)
	tsPacketChan := tse.StartPacketReadLoop()
	pesParser := mpeg2ts.NewPESParser(1048576)

	pesChan := pesParser.StartPESReadLoop()
	go func() {
		i := 0
		for {
			p := <-pesChan
			go func(index int, pes mpeg2ts.PES) {
				fmt.Printf("PES frame: %dbytes\n", len(p.ElementaryStream))
				fname := fmt.Sprintf("stream_%04d.bin", i)
				os.WriteFile(fname, p.ElementaryStream, 0644)
			}(i, p)
			i++
		}
	}()

	go func() {
		for {
			_, _, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				log.Fatalln(err)
				os.Exit(1)
			}
			// fmt.Println("udp receive")
			tse.Write(buf)
		}
	}()

	state := 0
	pmtPID := -1
	elementaryPID := uint16(0)
	for v := range tsPacketChan {
		fmt.Printf("state %d, PID=%04X\n", state, v.PID)
		if state == 0 && v.PID == mpeg2ts.PID_PAT {
			pat, err := v.ParsePAT()
			if err != nil {
				fmt.Println(err.Error())
			}
			for _, program := range pat.Programs {
				if program.ProgramNumber != 0 {
					fmt.Printf("ProgramMapPID found: %04X\n", program.ProgramMapPID)
					pmtPID = int(program.ProgramMapPID)
					state = 1
				}
			}
		} else if state == 1 && v.PID == uint16(pmtPID) {
			pmt, err := v.ParsePMT()

			if err != nil {
				fmt.Println(err.Error())
			}

			if len(pmt.Streams) > 0 {
				for _, s := range pmt.Streams {
					if s.Type == mpeg2ts.StreamTypeAVC {
						elementaryPID = s.ElementaryPID
						state = 2
						break
					}
				}
			}
		} else if state == 2 && v.PID == elementaryPID {
			fmt.Println("PES received")
			pesParser.EnqueueTSPacket(v)
		}

	}
}
