package main

import (
	"fmt"
	"os"

	mpeg2ts "github.com/misodengaku/go-mpeg2-ts"
)

var mpeg2 *mpeg2ts.MPEG2TS

func main() {
	var err error
	mpeg2, err = mpeg2ts.LoadStandardTS("test.ts")
	// mpeg2, err := mpeg2ts.LoadStandardTS("d443e813-631c-42b5-a25c-6b40558e4477_2022-02-02_055000.h264_gpac.ts")
	if err != nil {
		panic(err)
	}

	// go func() {
	var elementaryPID uint16
	patAll := mpeg2.FilterByPIDs(mpeg2ts.PID_PAT)
	for _, p := range patAll.PacketList.All() {
		fmt.Println("PAT frame:", p.Index, p.PID)
		patTable, _ := p.ParsePAT()

		for _, program := range patTable.Programs {
			fmt.Printf("PAT found. Program %04X -> Program Map PID %04X\n", program.ProgramNumber, program.ProgramMapPID)
			if program.ProgramNumber != 0 {
				programTable := mpeg2.FilterByPIDs(program.ProgramMapPID)
				for _, pmtPacket := range programTable.PacketList.All() {
					fmt.Printf("PMT found. Stream lookup\n")
					pmt, err := pmtPacket.ParsePMT()
					if err != nil {
						fmt.Printf("ParsePMT failed. %s\n", err.Error())
					}
					fmt.Printf("Stream %#v\n", pmt.Streams)

					if len(pmt.Streams) > 0 {
						for _, s := range pmt.Streams {
							fmt.Printf("Stream PID:%02X type:%02X\n", s.ElementaryPID, s.Type)
							if s.Type == mpeg2ts.StreamTypeAVC {
								elementaryPID = s.ElementaryPID
								// break
							}
						}
					}
				}
			}

		}
		// break
	}

	fmt.Printf("Video Stream PID is 0x%04X. start PES dump\n", elementaryPID)
	ES := mpeg2.FilterByPIDs(elementaryPID)
	pesParser := mpeg2ts.NewPESParser()
	pesChan := pesParser.StartPESReadLoop()
	go func() {
		i := 0
		for {
			p := <-pesChan
			fmt.Printf("PES frame: %dbytes\n", len(p.ElementaryStream))
			fname := fmt.Sprintf("es_%04d.bin", i)
			os.WriteFile(fname, p.ElementaryStream, 0644)
			i++
		}
	}()
	for _, p := range ES.PacketList.All() {
		// fmt.Printf("%#v\r\n", p)
		eop := false
		if len(p.AdaptationField.Stuffing) > 0 {
			eop = true
		}
		err = pesParser.EnqueueTSPacket(p, eop)
		if err != nil {
			panic(err)
		}
		// break
	}
	// fmt.Printf("%#v\r\n", pesParser)

	checkContinuity()

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

func checkContinuity() {
	fmt.Print("Continuity check->")
	if cr := mpeg2.CheckStream(); cr.DropCount > 0 {
		fmt.Println("frame drop detected!!")
		for _, v := range cr.DropList {
			fmt.Printf("frame index: %d\n", v.Index)
		}
	} else {
		fmt.Println("OK")
	}
}

func dumpPackets(count int) {

	for i, p := range mpeg2.PacketList.All() {
		fmt.Printf("%d sync:%x tei:%t pusi:%t tpi:%t pid:%x tsc:%d afc:%d cci:%d\r\n",
			i,
			p.SyncByte,
			p.TransportErrorIndicator,
			p.PayloadUnitStartIndicator,
			p.TransportPriorityIndicator,
			p.PID,
			p.TransportScrambleControl,
			p.AdaptationFieldControl,
			p.ContinuityCheckIndex)
		if p.HasAdaptationField() {
			fmt.Printf("\tAdaptationField dump: size:%d di:%t rai:%t espi:%t pcr:%t opcr:%t spf:%t tpdf:%t ef:%t\r\n",
				p.AdaptationField.Size,
				p.AdaptationField.DiscontinuityIndicator,
				p.AdaptationField.RandomAccessIndicator,
				p.AdaptationField.ESPriorityIndicator,
				p.AdaptationField.PCRFlag,
				p.AdaptationField.OPCRFlag,
				p.AdaptationField.SplicingPointFlag,
				p.AdaptationField.TransportPrivateDataFlag,
				p.AdaptationField.ExtensionFlag)
		}
		// }
		if count > 0 && count-1 == i {
			break
		}

	}
}
