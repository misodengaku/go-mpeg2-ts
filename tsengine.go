package mpeg2ts

import (
	"context"
	"sync"
	"time"
)

type TransportStreamEngine struct {
	buffer     []byte
	bufferSize int
	packets    PacketList
	chunkSize  int
	mutex      *sync.Mutex
}

func InitTSEngine(chunkSize, bufferSize int) (TransportStreamEngine, error) {
	tse := TransportStreamEngine{}
	tse.bufferSize = bufferSize
	tse.buffer = make([]byte, 0, tse.bufferSize)
	tse.chunkSize = chunkSize
	tse.packets, _ = NewPacketList(chunkSize)
	tse.mutex = &sync.Mutex{}
	return tse, nil
}

func (tse *TransportStreamEngine) StartPacketReadLoop(ctx context.Context) <-chan Packet {
	cp := make(chan Packet)
	go func(packetOutChan chan Packet) {
		for {
			select {
			case <-ctx.Done():
				close(packetOutChan)
				return
			default:
				// pass
			}
			if tse.getBufferLength() < tse.chunkSize {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			tse.mutex.Lock()
			for len(tse.buffer) >= tse.chunkSize {
				syncIndex := -1
				for i, v := range tse.buffer {
					if v == 0x47 {
						syncIndex = i
						break
					}
				}

				var packetData []byte
				if syncIndex == -1 {
					// tse.buffer is dirty. clear and continue
					tse.dequeueWithoutLock(len(tse.buffer))
					continue
				} else if syncIndex > 0 {
					tse.dequeueWithoutLock(syncIndex)
					if len(tse.buffer) >= tse.chunkSize {
						packetData = tse.dequeueWithoutLock(tse.chunkSize)
					} else {
						break
					}
				} else {
					packetData = tse.dequeueWithoutLock(tse.chunkSize)
				}
				if packetData == nil {
					break
				}
				packet := Packet{}
				packet.Data = make([]byte, tse.chunkSize)
				copy(packet.Data, packetData)
				err := packet.parseHeader()
				if err != nil {
					continue
				} else {
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
		return r
	}
	return nil
}
func (tse *TransportStreamEngine) dequeue(size int) []byte {
	tse.mutex.Lock()
	r := tse.dequeueWithoutLock(size)
	tse.mutex.Unlock()
	return r
}

func (tse *TransportStreamEngine) enqueueWithoutLock(in []byte) {
	v := make([]byte, len(in))
	copy(v, in)
	tse.buffer = append(tse.buffer, v...)
}

func (tse *TransportStreamEngine) enqueue(in []byte) {
	tse.mutex.Lock()
	tse.enqueueWithoutLock(in)
	tse.mutex.Unlock()
}

func (tse *TransportStreamEngine) getBufferLength() int {
	tse.mutex.Lock()
	l := len(tse.buffer)
	tse.mutex.Unlock()
	return l
}

func (tse *TransportStreamEngine) Write(p []byte) (n int, err error) {
	tse.enqueue(p)
	return len(p), nil
}
