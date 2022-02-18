package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	mpeg2ts "github.com/misodengaku/go-mpeg2-ts"
)

func main() {
	udpConn, err := net.ListenPacket("udp", "0.0.0.0:50000")
	if err != nil {
		panic(err)
	}

	// buf := make([]byte, 1048576)
	log.Println("Starting udp server...")
	tse, _ := mpeg2ts.InitTSEngine(mpeg2ts.PacketSizeDefault, 10*1048576)
	tsPacketChan := tse.StartPacketReadLoop()
	pesParser := mpeg2ts.NewPESParser(1048576)
	pesOutCount := uint32(0)
	udpInCount := uint32(0)

	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 5000,
	}
	udpSenderConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		panic(err)
	}
	udpSenderConn.SetWriteBuffer(128 * 1048576)
	pesChan := pesParser.StartPESReadLoop()
	go func() {
		for {
			pes := <-pesChan

			atomic.AddUint32(&pesOutCount, uint32(len(pes.ElementaryStream)))
		}
	}()

	m := sync.Mutex{}
	udpBuffer := make([]byte, 0, 16*1048576)
	go func() {
		buf := [1500]byte{}

		for {
			n, _, err := udpConn.ReadFrom(buf[:])
			if err != nil {
				log.Fatalln(err)
				os.Exit(1)
			}
			if n > 0 {
				// fmt.Println("udp receive")
				m.Lock()
				udpBuffer = append(udpBuffer, buf[:n]...)

				atomic.AddUint32(&udpInCount, uint32(n))
				m.Unlock()
			}
		}
	}()

	state := 0
	pmtPID := -1
	elementaryPID := uint16(0)
	frameIndex := uint64(0)
	continuityIndexs := map[uint16]byte{}
	stateChanged := true
	bufTicker := time.NewTicker(27 * time.Millisecond)
	statTicket := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-bufTicker.C:
			m.Lock()
			tse.Write(udpBuffer)
			udpBuffer = udpBuffer[:0]
			m.Unlock()
		case <-statTicket.C:
			fmt.Println("-------------------")
			old := atomic.SwapUint32(&pesOutCount, 0)
			fmt.Printf("pesrate %dKbps\n", old/1024*8)
			old = atomic.SwapUint32(&udpInCount, 0)
			fmt.Printf("udprate %dKbps\n", old/1024*8)
		case v := <-tsPacketChan:
			if stateChanged {
				fmt.Printf("state %d, PID=%04X\n", state, v.PID)
				stateChanged = false
			}
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
						stateChanged = true
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
							stateChanged = true
							break
						}
					}
				}
			} else if state == 2 && v.PID == elementaryPID {
				// fmt.Println("PES received")
				pesParser.EnqueueTSPacket(v)
			}

			ci, ok := continuityIndexs[v.PID]
			if ok {
				if (ci+1)%16 != v.ContinuityCheckIndex {
					fmt.Printf("drop frame detected! tsframe: %d PID: %02X expected: %d actual: %d\n", frameIndex, v.PID, (ci+1)%16, v.ContinuityCheckIndex)
				}
			}
			continuityIndexs[v.PID] = v.ContinuityCheckIndex
			frameIndex++
		}
	}
}
