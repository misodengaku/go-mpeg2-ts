package main

import (
	"fmt"
	"log"
	"net"
	"os"

	mpeg2ts "github.com/misodengaku/go-mpeg2-ts"
)

func main() {
	// mpeg2, err := mpeg2ts.LoadStandardTS("test.ts")
	// mpeg2, err := mpeg2ts.LoadStandardTS("d443e813-631c-42b5-a25c-6b40558e4477_2022-02-02_055000.h264_gpac.ts")
	// if err != nil {
	// 	panic(err)
	// }

	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 50000,
	}
	udpConn, _ := net.ListenUDP("udp", udpAddr)

	buf := make([]byte, 1048576)
	log.Println("Starting udp server...")
	tse, _ := mpeg2ts.InitTSEngine()
	tsPacketChan := tse.StartPacketReadLoop()
	pesParser := mpeg2ts.NewPESParser()
	// pesChan := pesParser.StartPESReadLoop()

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
		// fmt.Printf("%#v\n", v)
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
			eop := false
			if len(v.AdaptationField.Stuffing) > 0 {
				eop = true
			}
			pesParser.EnqueueTSPacket(v, eop)
		}

	}

	// fmt.Println("Continuity check")
	// if cr := mpeg2.CheckStream(); cr.DropCount > 0 {
	// 	fmt.Println("frame drop detected!!")
	// 	for _, v := range cr.DropList {
	// 		fmt.Printf("frame index: %d\n", v.Index)
	// 	}
	// } else {
	// 	fmt.Println("OK")
	// }

	// go func() {
	// var elementaryPID uint16
	// patAll := mpeg2.FilterByPIDs(mpeg2ts.PID_PAT)
	// for _, p := range patAll.Packets {
	// 	// fmt.Println("PAT frame:", p.Index, p.PID, p.Data)
	// 	patTable, _ := p.ParsePAT()
	// 	fmt.Printf("%#v\r\n", patTable)

	// 	for _, program := range patTable.Programs {
	// 		// fmt.Printf("Program Table: %#v\n", program)
	// 		if program.ProgramNumber != 0 {
	// 			programTable := mpeg2.FilterByPIDs(program.ProgramMapPID)
	// 			fmt.Printf("Program Table: %#v\n", programTable)
	// 			for _, pmtPacket := range programTable.Packets {
	// 				pmt, _ := pmtPacket.ParsePMT()
	// 				fmt.Printf("parsed %#v\n", pmt)
	// 				if len(pmt.Streams) > 0 {
	// 					for _, s := range pmt.Streams {
	// 						if s.Type == mpeg2ts.StreamTypeAVC {
	// 							elementaryPID = s.ElementaryPID
	// 							break
	// 						}
	// 					}
	// 				}
	// 			}
	// 		}

	// 	}
	// 	// break
	// }

	// fmt.Printf("Video Stream PID is 0x%04X\n", elementaryPID)
	// ES := mpeg2.FilterByPIDs(elementaryPID)
	// pesParser := mpeg2ts.NewPESParser()
	// for _, p := range ES.Packets {
	// 	fmt.Printf("%#v\r\n", p)
	// 	pesParser.EnqueueTSPacket(p)
	// 	break
	// }
	// fmt.Printf("%#v\r\n", pesParser)

	// pmt := mpeg2.FilterByPIDs(mpeg2ts.PID_EIT)
	// for _, p := range pmt.Packets {
	// 	fmt.Printf("%#v\r\n", p)
	// }
	// 	ch <- struct{}{}
	// }()
	// <-ch

	// for streaming
	// WaitForPAT()
	// ParsePAT()
	// WaitForPMT
	// ParsePMT
	// ProgramNum X is video streaming. start forwarding
	// while(1)
	// if PID==X: forwarding()
	// elif PID==PAT: ParsePAT
}
