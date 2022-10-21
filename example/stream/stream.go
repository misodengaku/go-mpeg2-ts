package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	mpeg2ts "github.com/misodengaku/go-mpeg2-ts"
)

func main() {
	tse, _ := mpeg2ts.InitTSEngine(mpeg2ts.PacketSizeDefault, 10*1048576)
	tsPacketChan := tse.StartPacketReadLoop()
	pesParser := mpeg2ts.NewPESParser(1048576)

	pesOutCount := uint32(0)
	udpInCount := uint32(0)

	m := sync.Mutex{}
	udpBuffer := make([]byte, 0, 16*1048576)

	log.Println("Starting udp server...")
	udpConn, err := net.ListenPacket("udp", "0.0.0.0:50000")
	if err != nil {
		panic(err)
	}

	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 5000,
	}
	udpSenderConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		panic(err)
	}
	udpSenderConn.SetWriteBuffer(8 * 1048576)
	ctx := context.Background()
	pesChan := pesParser.StartPESReadLoop(ctx)
	go func() {
		// PES receiver
		for {
			pes := <-pesChan
			atomic.AddUint32(&pesOutCount, uint32(len(pes.ElementaryStream)))
		}
	}()

	go func() {
		// receive UDP packet, and store to receive buffer
		buf := [1500]byte{}
		for {
			n, _, err := udpConn.ReadFrom(buf[:])
			if err != nil {
				log.Fatalln(err)
				os.Exit(1)
			}
			if n > 0 {
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
	statTicker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-bufTicker.C:
			m.Lock()
			tse.Write(udpBuffer)
			udpBuffer = udpBuffer[:0]
			m.Unlock()
		case <-statTicker.C:
			log.Println("-------------------")
			old := atomic.SwapUint32(&pesOutCount, 0)
			log.Printf("pesrate %dKbps\n", old/1024*8)
			old = atomic.SwapUint32(&udpInCount, 0)
			log.Printf("udprate %dKbps\n", old/1024*8)
		case v := <-tsPacketChan:
			if stateChanged {
				log.Printf("state %d, PID=%04X\n", state, v.PID)
				stateChanged = false
			}

			if state == 0 && v.PID == mpeg2ts.PID_PAT {
				// PAT receive state

				pat, err := v.ParsePAT()
				if err != nil {
					log.Fatalln(err)
					os.Exit(1)
				}
				for _, program := range pat.Programs {
					if program.ProgramNumber != 0 {
						log.Printf("ProgramMapPID found: %04X\n", program.ProgramMapPID)
						pmtPID = int(program.ProgramMapPID)
						state = 1
						stateChanged = true
					}
				}
			} else if state == 1 && v.PID == uint16(pmtPID) {
				// PMT receive state

				pmt, err := v.ParsePMT()
				if err != nil {
					log.Fatalln(err)
					os.Exit(1)
				}

				if len(pmt.Streams) > 0 {
					for _, s := range pmt.Streams {
						if s.Type == mpeg2ts.StreamTypeAVC {
							// H.264 stream found. transition to PES receive state
							elementaryPID = s.ElementaryPID
							state = 2
							stateChanged = true
							break
						}
					}
				}
			} else if state == 2 && v.PID == elementaryPID {
				// PES receive state
				pesParser.EnqueueTSPacket(v)
			}

			// check packet continuity
			ci, ok := continuityIndexs[v.PID]
			if ok {
				if (ci+1)%16 != v.ContinuityCheckIndex {
					log.Printf("drop frame detected! tsframe: %d PID: %02X expected: %d actual: %d\n", frameIndex, v.PID, (ci+1)%16, v.ContinuityCheckIndex)
				}
			}
			continuityIndexs[v.PID] = v.ContinuityCheckIndex
			frameIndex++
		}
	}
}
