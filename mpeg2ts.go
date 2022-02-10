package mpeg2ts

import (
	"fmt"
	"os"
)

func New(chunkSize int) *MPEG2TS {
	m := MPEG2TS{}
	m.PacketList, _ = NewPacketList(chunkSize)
	return &m
}

func NewWithPacketCount(packetCount int64, chunkSize int) *MPEG2TS {
	m := New(chunkSize)
	m.PacketList.packets = make([]Packet, 0, packetCount)
	return m
}

func LoadStandardTS(fname string) (*MPEG2TS, error) {
	return loadFile(fname, PacketSizeDefault)
}
func LoadStandardTSWithFEC(fname string) (*MPEG2TS, error) {
	return loadFile(fname, PacketSizeWithFEC)
}

func loadFile(fname string, packetLength int) (*MPEG2TS, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var fsize int64

	if fi, err := file.Stat(); err == nil {
		fsize = fi.Size()
	} else {
		return nil, err
	}

	if fsize < PacketSizeDefault {
		return nil, fmt.Errorf("filesize (%d) is smaller than the minimum (%d)", fsize, PacketSizeDefault)
	}

	m := NewWithPacketCount(fsize/int64(packetLength), packetLength)

	packetBuffer := make([]byte, packetLength)
	i := 0
	for {
		// fmt.Println(i)
		n, err := file.Read(packetBuffer)
		if err != nil {
			// Readエラー処理
			if err.Error() == "EOF" {
				break
			}
			return nil, err

		}
		if n < packetLength {
			return nil, fmt.Errorf("sirikire %d", n)
		}

		err = m.PacketList.AddBytes(packetBuffer, packetLength)
		if err != nil {
			return nil, err
		}

		i++
	}
	return m, nil
}

func (m MPEG2TS) CheckStream() StreamCheckResult {
	cr := StreamCheckResult{}
	ci := map[uint16]byte{}
	dc := 0

	for i := uint16(0); i < 0x2000; i++ {
		ci[i] = byte(16)
	}

	for i, p := range m.PacketList.All() {
		if p.PID == PID_NullPacket {
			continue
		}
		// if ci[p.PID] == 16 {
		// 	fmt.Printf("PID: %d ci: nil != pci: %d\r\n", p.PID, p.ContinuityCheckIndex)
		// } else {
		// 	fmt.Printf("PID: %d ci: %d != pci: %d\r\n", p.PID, (ci[p.PID]+1)%16, p.ContinuityCheckIndex)
		// }
		if ci[p.PID] == 16 {
			// 初期値
			if p.AdaptationFieldControl != 0 && p.AdaptationFieldControl != 2 {
				ci[p.PID] = p.ContinuityCheckIndex
			} else {
				ci[p.PID] = 1
				// fmt.Println("skip")
			}
		} else if (ci[p.PID]+1)%16 != p.ContinuityCheckIndex {
			if p.AdaptationFieldControl != 0 && p.AdaptationFieldControl != 2 {
				dc++
				ci[p.PID] = p.ContinuityCheckIndex
				cr.DropList = append(cr.DropList, struct {
					Description string
					Index       int
				}{"frame drop detected", i})
				// fmt.Println("Continuity check error: index " + strconv.Itoa(i))
			}
			// fmt.Printf("PID: %d ci: %d != pci: %d\r\n", p.PID, (ci[p.PID]+1)%16, p.ContinuityCheckIndex)
			// return errors.New("Continuity check error: index " + strconv.Itoa(i))
		} else {
			if p.AdaptationFieldControl != 0 && p.AdaptationFieldControl != 2 {
				ci[p.PID] = p.ContinuityCheckIndex
			} else {
				// fmt.Println("skip")
			}
		}
	}
	cr.DropCount = dc
	// fmt.Println("Drop frame:", dc)
	return cr
}

func (m *MPEG2TS) FilterByPIDs(pids ...uint16) *MPEG2TS {
	mx := New(m.chunkSize)
	for _, p := range m.PacketList.All() {
		for _, id := range pids {
			if p.PID == id {
				// fmt.Println(p.Index)
				// fmt.Printf("%#v\r\n", p.Data)
				mx.AddPacket(p)
				break
			}
		}
	}
	return mx
}
