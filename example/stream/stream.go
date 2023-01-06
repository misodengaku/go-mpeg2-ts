package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	mpeg2ts "github.com/misodengaku/go-mpeg2-ts"
)

const fromUDP = false

var (
	udpInCount      = uint32(0)
	pesOutCount     = uint32(0)
	inBuffer        []byte
	inBufferMutex   *sync.Mutex
	disableCRCcheck = false
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	inBuffer = make([]byte, 0, 16*1048576)
	tse, _ := mpeg2ts.InitTSEngine(mpeg2ts.PacketSizeDefault, 1024)
	tsPacketChan := tse.StartPacketReadLoop(ctx)
	pesParser := mpeg2ts.NewPESParser(1500)

	inBufferMutex = &sync.Mutex{}

	if fromUDP {
		go startUDPSrc(ctx)
	} else {
		go func() {
			err := startFileSrc(ctx, "test.ts", 1000*1000)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					panic(err)
				}
				// error is EOF. cancel only
			}
			cancel()
		}()
	}

	pesChan := pesParser.StartPESReadLoop(ctx)
	go func() {
		// PES receiver
		for {
			select {
			case pes := <-pesChan:
				atomic.AddUint32(&pesOutCount, uint32(len(pes.ElementaryStream)))
			case <-ctx.Done():
				log.Println("PES receiver exit")
				return
			}
		}
	}()

	state := 0
	pmtPID := -1
	elementaryPID := uint16(0)
	frameIndex := uint64(0)
	continuityIndexes := map[uint16]byte{}
	stateChanged := true
	bufTicker := time.NewTicker(27 * time.Millisecond)
	statTicker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-bufTicker.C
			inBufferMutex.Lock()
			tse.Write(inBuffer)
			inBuffer = inBuffer[:0]
			inBufferMutex.Unlock()
		}
	}()
Loop:
	for {
		select {
		case <-statTicker.C:
			log.Println("-------------------")
			old := atomic.SwapUint32(&pesOutCount, 0)
			log.Printf("pesrate %dKbps frame:%d\n", old/1024*8, frameIndex)
			old = atomic.SwapUint32(&udpInCount, 0)
			log.Printf("incoming rate %dKbps\n", old/1024*8)
		case v, ok := <-tsPacketChan:
			if !ok {
				log.Fatal("tsPacketChan is closed!")
			}
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
						break
					}
				}
			} else if state == 1 && v.PID == uint16(pmtPID) {
				// PMT receive state

				pmt, err := v.ParsePMT(disableCRCcheck)
				if err != nil {
					log.Println("invalid PMT!", err)
					// continue
				}

				if len(pmt.Streams) > 0 {
					for _, s := range pmt.Streams {
						if s.Type == mpeg2ts.StreamTypeAVC {
							// H.264 stream found. transition to PES receive state
							elementaryPID = s.ElementaryPID
							state = 2
							stateChanged = true
							break
						} else if s.Type == mpeg2ts.StreamTypeISO13818_2_Video {
							// ISO 13818-2 Video stream found. transition to PES receive state
							elementaryPID = s.ElementaryPID
							state = 2
							stateChanged = true
							break
						}
					}
				}
			} else if state == 2 && v.PID == elementaryPID {
				if stateChanged {
					log.Printf("state %d, PID=%04X\n", state, v.PID)
					stateChanged = false
				}

				// PES receive state
				pesParser.EnqueueTSPacket(v)
			}

			// check packet continuity
			ci, ok := continuityIndexes[v.PID]
			if ok {
				if (ci+1)%16 != v.ContinuityCheckIndex {
					// log.Printf("drop frame detected! tsframe: %d PID: %02X expected: %d actual: %d\n", frameIndex, v.PID, (ci+1)%16, v.ContinuityCheckIndex)
				}
			}
			continuityIndexes[v.PID] = v.ContinuityCheckIndex
			frameIndex++
		case <-ctx.Done():
			break Loop
		}
	}
	time.Sleep(1 * time.Second)
	log.Printf("stream is end\n")
}

func startUDPSrc(ctx context.Context) {

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
				inBufferMutex.Lock()
				inBuffer = append(inBuffer, buf[:n]...)
				atomic.AddUint32(&udpInCount, uint32(n))
				inBufferMutex.Unlock()
			}
		}
	}()
	log.Println("udp server running...")
	<-ctx.Done()
}

func startFileSrc(_ctx context.Context, filename string, byterateLimit int) error {
	log.Printf("Starting file source... (max %dbyte/s)\n", byterateLimit/1000)
	ctx, cancel := context.WithCancel(_ctx)

	var err error
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(f)

	go func() {
		// receive UDP packet, and store to receive buffer
		buf := [1500]byte{}
		var n int
		for {
			n, err = reader.Read(buf[:])
			if err != nil {
				cancel()
				return
			}
			if n > 0 {
				inBufferMutex.Lock()
				inBuffer = append(inBuffer, buf[:n]...)
				atomic.AddUint32(&udpInCount, uint32(n))
				inBufferMutex.Unlock()
			}
			time.Sleep(1 * time.Microsecond)
		}
	}()
	log.Println("file src running...")
	<-ctx.Done()
	log.Println("file src exited")
	return err
}
