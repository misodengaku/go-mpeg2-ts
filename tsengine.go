package mpeg2ts

import (
	"fmt"
	"sync"
	"time"
)

type TransportStreamEngine struct {
	buffer           []byte
	bufferSize       int
	packets          PacketList
	chunkSize        int
	byteIncomingChan chan struct{}
	mutex            *sync.Mutex
	// f                *os.File
}

func InitTSEngine(chunkSize, bufferSize int) (TransportStreamEngine, error) {
	tse := TransportStreamEngine{}
	tse.bufferSize = bufferSize
	tse.buffer = make([]byte, 0, tse.bufferSize)
	tse.chunkSize = chunkSize
	tse.packets, _ = NewPacketList(chunkSize)
	tse.byteIncomingChan = make(chan struct{})
	tse.mutex = &sync.Mutex{}
	// tse.f, _ = os.OpenFile("dump_tse.ts", os.O_RDWR|os.O_CREATE, 0755)
	return tse, nil
}

func (tse *TransportStreamEngine) StartPacketReadLoop() chan Packet {
	cp := make(chan Packet)
	go func(packetOutChan chan Packet) {
		for {
			if tse.getBufferLength() < tse.chunkSize {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			tse.mutex.Lock()
			for len(tse.buffer) >= tse.chunkSize {
				// fmt.Println("tsloop", len(tse.buffer))
				syncIndex := -1
				for i, v := range tse.buffer {
					if v == 0x47 {
						syncIndex = i
						break
					}
				}
				if syncIndex == -1 {
					// tse.buffer is dirty. clear and continue
					fmt.Println("sync byte is not found in buffer", len(tse.buffer))
					// fmt.Printf("buffer: %#v\n", tse.buffer)
					tse.dequeueWithoutLock(len(tse.buffer))
					continue
				} else if syncIndex > 0 {
					fmt.Printf("synced. drop %dbytes\n", syncIndex)
				}
				packetData := tse.dequeueWithoutLock(tse.chunkSize)
				if packetData == nil {
					break
				}
				packet := Packet{}
				packet.Data = make([]byte, tse.chunkSize)
				copy(packet.Data, packetData)
				err := packet.parseHeader()
				if err != nil {
					fmt.Printf("[ERROR] %s\n", err)
				} else {
					if len(packet.Data) != 188 {
						fmt.Println("packetout size=", len(packet.Data))

					}
					packetOutChan <- packet
				}
			}
			tse.mutex.Unlock()
		}
	}(cp)
	return cp
}

func (tse *TransportStreamEngine) dequeueWithoutLock(size int) []byte {
	var r []byte
	if size > 0 && len(tse.buffer) >= size {
		r = make([]byte, size)
		copy(r, tse.buffer)
		tse.buffer = append(tse.buffer[:0], tse.buffer[size:]...)
		// fmt.Printf("deq: %p, len: %d, cap: %d\n", tse.buffer, len(tse.buffer), cap(tse.buffer))
		return r
	}
	return nil
}
func (tse *TransportStreamEngine) dequeue(size int) []byte {
	// fmt.Printf("deqgate\n")
	tse.mutex.Lock()
	r := tse.dequeueWithoutLock(size)
	tse.mutex.Unlock()
	return r
}

func (tse *TransportStreamEngine) enqueueWithoutLock(in []byte) {
	// fmt.Printf("in: %#v\n", in)
	tse.buffer = append(tse.buffer, in...)
	// fmt.Printf("enq: %p, len: %d, cap: %d\n", tse.buffer, len(tse.buffer), cap(tse.buffer))
}

func (tse *TransportStreamEngine) enqueue(in []byte) {
	tse.mutex.Lock()
	tse.enqueueWithoutLock(in)
	tse.mutex.Unlock()
}

func (tse *TransportStreamEngine) getBufferLength() int {
	//fmt.Println("ts len lock")
	tse.mutex.Lock()
	l := len(tse.buffer)
	// //fmt.Println("len ", l)
	tse.mutex.Unlock()
	//fmt.Println("ts len unlock")
	return l
}

func (tse *TransportStreamEngine) Write(p []byte) (n int, err error) {
	// w := make([]byte, len(p))
	// copy(w, p)
	// fmt.Println("write")
	tse.enqueue(p)
	// tse.byteIncomingChan <- struct{}{}

	// fmt.Println("writeok")
	return len(p), nil
}
