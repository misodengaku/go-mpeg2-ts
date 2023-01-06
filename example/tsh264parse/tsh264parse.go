package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	h264parse "github.com/misodengaku/go-h264-parse"
	mpeg2ts "github.com/misodengaku/go-mpeg2-ts"
)

var enableESDump = false
var disableCRCcheck = false

var mpeg2 *mpeg2ts.MPEG2TS

func main() {
	var err error
	mpeg2, err = mpeg2ts.LoadStandardTS("test.ts")
	if err != nil {
		panic(err)
	}

	var elementaryPID uint16
	patPackets := mpeg2.FilterByPIDs(mpeg2ts.PID_PAT)
	for _, p := range patPackets.PacketList.All() {
		fmt.Println("PAT frame:", p.Index, p.PID)
		patTable, _ := p.ParsePAT()

		for _, program := range patTable.Programs {
			fmt.Printf("PAT found. Program %04X -> Program Map PID %04X\n", program.ProgramNumber, program.ProgramMapPID)
			if program.ProgramNumber != 0 {
				programTable := mpeg2.FilterByPIDs(program.ProgramMapPID)
				for _, pmtPacket := range programTable.PacketList.All() {
					fmt.Printf("PMT found. Stream lookup\n")
					pmt, err := pmtPacket.ParsePMT(disableCRCcheck)
					if err != nil {
						fmt.Printf("ParsePMT failed. %s\n", err.Error())
					}
					fmt.Printf("Stream %#v\n", pmt.Streams)

					if len(pmt.Streams) > 0 {
						for _, s := range pmt.Streams {
							fmt.Printf("Stream PID:%02X type:%02X\n", s.ElementaryPID, s.Type)
							if s.Type == mpeg2ts.StreamTypeAVC {
								elementaryPID = s.ElementaryPID
								break
							}
						}
						if elementaryPID != 0 {
							break
						}
					}
				}
			}
			if elementaryPID != 0 {
				break
			}
		}

		if elementaryPID != 0 {
			break
		}
	}

	ctx := context.Background()
	fmt.Printf("Video Stream PID is 0x%04X. start PES dump\n", elementaryPID)
	pesPackets := mpeg2.FilterByPIDs(elementaryPID)
	pesParser := mpeg2ts.NewPESParser(1500)

	c := pesParser.StartPESReadLoop(ctx)
	nalus := h264parse.NALUs{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		i := 0
		for p := range c {
			fmt.Printf("ES frame: %dbytes\n", len(p.ElementaryStream))
			if enableESDump {
				fname := fmt.Sprintf("output/es_%04d.bin", i)
				os.WriteFile(fname, p.ElementaryStream, 0644)
			}

			n, err := h264parse.Unmarshal(p.ElementaryStream)
			if err != nil {
				panic(err)
			}
			nalus.Units = append(nalus.Units, n.Units...)
			i++
		}
		wg.Done()
	}()

	packets := pesPackets.PacketList.All()
	for i, p := range packets {
		if i < len(packets)-1 {
			err = pesParser.EnqueueTSPacket(p)
		} else {
			err = pesParser.EnqueueLastTSPacket(p)
		}
		if err != nil {
			panic(err)
		}
	}
	wg.Wait()

	filteredNALUs := make([]h264parse.NAL, 0, len(nalus.Units))
	for n, nal := range nalus.Units {
		fmt.Printf("%d:\t%s (%d)\n", n, nal.UnitType.String(), nal.UnitType)
		if nal.UnitType == h264parse.AccessUnitDelimiter {
			continue
		}
		if nal.UnitType == h264parse.SupplementalEnhancementInformation {
			continue
		}
		filteredNALUs = append(filteredNALUs, nal)
	}
	nalus.Units = filteredNALUs
	nb, _ := h264parse.Marshal(nalus)
	os.WriteFile("dump.h264", nb, 0755)

	checkContinuity()
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
				p.AdaptationField.Length,
				p.AdaptationField.DiscontinuityIndicator,
				p.AdaptationField.RandomAccessIndicator,
				p.AdaptationField.ESPriorityIndicator,
				p.AdaptationField.PCRFlag,
				p.AdaptationField.OPCRFlag,
				p.AdaptationField.SplicingPointFlag,
				p.AdaptationField.TransportPrivateDataFlag,
				p.AdaptationField.ExtensionFlag)
		}
		if count > 0 && count-1 == i {
			break
		}

	}
}
