package mpeg2ts

import (
	"fmt"
	"strconv"
)

type Mpeg2TS struct {
	Packets Mpeg2TSPackets
}

type Mpeg2TSPackets []Mpeg2TSPacket

type Mpeg2TSPacket struct {
	Data                       []byte
	SyncByte                   byte
	PID                        uint16 // 中身は13bit
	TransportScrambleControl   byte
	AdaptationFieldControl     byte
	TransportErrorIndicator    bool
	PayloadUnitStartIndicator  bool
	TransportPriorityIndicator bool
	ContinuityCheckIndex       byte
	AdaptationField            *AdaptationField
}

type AdaptationField struct {
	Size                     byte
	DiscontinuityIndicator   bool
	RandomAccessIndicator    bool
	ESPriorityIndicator      bool
	PCRFlag                  bool
	OPCRFlag                 bool
	SplicingPointFlag        bool
	TransportPrivateDataFlag bool
	ExtensionFlag            bool
}

func (m Mpeg2TS) New(count int64) Mpeg2TS {
	m.Packets = make(Mpeg2TSPackets, count)
	return m
}

func (m Mpeg2TS) CheckStream() error {
	ci := map[uint16]byte{}
	dc := 0

	for i := uint16(0); i < 0x2000; i++ {
		ci[i] = byte(16)
	}

	for i, p := range m.Packets {
		if p.PID == PID_NullPacket {
			continue
		}
		if ci[p.PID] == 16 {
			fmt.Printf("PID: %d ci: nil != pci: %d\r\n", p.PID, p.ContinuityCheckIndex)
		} else {
			fmt.Printf("PID: %d ci: %d != pci: %d\r\n", p.PID, (ci[p.PID]+1)%16, p.ContinuityCheckIndex)
		}
		if ci[p.PID] == 16 {
			// 初期値
			if p.AdaptationFieldControl != 0 && p.AdaptationFieldControl != 2 {
				ci[p.PID] = p.ContinuityCheckIndex
			} else {
				ci[p.PID] = 1
				fmt.Println("skip")
			}
		} else if (ci[p.PID]+1)%16 != p.ContinuityCheckIndex {
			if p.AdaptationFieldControl != 0 && p.AdaptationFieldControl != 2 {
				dc++
				ci[p.PID] = p.ContinuityCheckIndex
				fmt.Println("Continuity check error: index " + strconv.Itoa(i))
			} else {
				fmt.Println("skip")
			}
			// fmt.Printf("PID: %d ci: %d != pci: %d\r\n", p.PID, (ci[p.PID]+1)%16, p.ContinuityCheckIndex)
			// return errors.New("Continuity check error: index " + strconv.Itoa(i))
		} else {
			if p.AdaptationFieldControl != 0 && p.AdaptationFieldControl != 2 {
				ci[p.PID] = p.ContinuityCheckIndex
			} else {
				fmt.Println("skip")
			}
		}
	}
	fmt.Println("Drop frame:", dc)
	return nil
}
