package mpeg2ts

import (
	"fmt"
	"sync"
)

type TransportStreamEngine struct {
	buffer           []byte
	bufferSize       int
	packets          PacketList
	chunkSize        int
	byteIncomingChan chan struct{}
	mutex            *sync.Mutex
}

func InitTSEngine(chunkSize, bufferSize int) (TransportStreamEngine, error) {
	tse := TransportStreamEngine{}
	tse.bufferSize = bufferSize
	tse.buffer = make([]byte, 0, tse.bufferSize)
	tse.chunkSize = chunkSize
	tse.packets, _ = NewPacketList(chunkSize)
	tse.byteIncomingChan = make(chan struct{})
	tse.mutex = &sync.Mutex{}
	return tse, nil
}

func (tse *TransportStreamEngine) StartPacketReadLoop() chan Packet {
	cp := make(chan Packet)
	go func(packetOutChan chan Packet) {
		for range tse.byteIncomingChan {
			tse.mutex.Lock()
			for len(tse.buffer) >= tse.chunkSize {
				syncIndex := -1
				for i, v := range tse.buffer {
					if v == 0x47 {
						syncIndex = i
						break
					}
				}
				if syncIndex == -1 {
					// tse.buffer is dirty. clear and continue
					tse.dequeue(1)
					continue
				}
				packet := Packet{}
				packet.Data = make([]byte, tse.chunkSize)
				r := tse.dequeue(tse.chunkSize)
				if r == nil {
					break
				}
				copy(packet.Data, r)
				err := packet.parseHeader()
				if err != nil {
					fmt.Printf("[ERROR] %s\n", err)
				} else {
				packetOutChan <- packet
				}
			}
			tse.mutex.Unlock()
		}
	}(cp)
	return cp
}

func (tse *TransportStreamEngine) dequeue(size int) []byte {
	var r []byte
	if size > 0 && len(tse.buffer) > tse.chunkSize {
	// tse.mutex.Lock()
		r = make([]byte, size)
		copy(r, tse.buffer)
		tse.buffer = append(tse.buffer[:0], tse.buffer[size:]...)
		return r
	}
	return nil
}

func (tse *TransportStreamEngine) enqueue(in []byte) {
	tse.buffer = append(tse.buffer, in...)
}

func (tse *TransportStreamEngine) getBufferLength() int {
	l := len(tse.buffer)
	return l
}

func (tse *TransportStreamEngine) Write(p []byte) (n int, err error) {
	tse.mutex.Lock()
	tse.enqueue(p)
	tse.mutex.Unlock()
	tse.byteIncomingChan <- struct{}{}
	return len(p), nil
}
