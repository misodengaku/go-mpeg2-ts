package mpeg2ts

import (
	"fmt"
	"sync"
	"time"
)

type TransportStreamEngine struct {
	buf        []byte
	bufferSize int
	packets    PacketList
	mutex      *sync.Mutex
}

func InitTSEngine(chunkSize, bufferSize int) (TransportStreamEngine, error) {
	tse := TransportStreamEngine{}
	tse.bufferSize = bufferSize
	tse.buf = make([]byte, 0, tse.bufferSize)
	tse.packets, _ = NewPacketList(chunkSize)
	tse.mutex = &sync.Mutex{}
	return tse, nil
}

func (tse *TransportStreamEngine) StartPacketReadLoop() chan Packet {
	cp := make(chan Packet)
	go func(packetOutChan chan Packet) {
		for {
			// fmt.Println("wait", len(tse.buf))
			tse.mutex.Lock()
			if len(tse.buf) < PacketSizeDefault {
				tse.mutex.Unlock()
				time.Sleep(1 * time.Microsecond)
				continue
			}
			syncIndex := -1
			for i, v := range tse.buf {
				if v == 0x47 {
					syncIndex = i
					break
				}
			}
			if syncIndex == -1 {
				// tse.buf is dirty. clear and continue
				// fmt.Println("clear buf")
				tse.buf = make([]byte, 0, tse.bufferSize)
				tse.mutex.Unlock()
				continue
			} else if syncIndex > 0 {
				dirty := tse.buf[:syncIndex]
				fmt.Printf("trim %#v\n", dirty)
				tse.buf = tse.buf[syncIndex:]
			}
			buf := tse.buf[0:PacketSizeDefault]
			tse.buf = tse.buf[PacketSizeDefault:]
			err := tse.packets.AddBytes(buf, PacketSizeDefault)
			if err != nil {
				tse.mutex.Unlock()
				continue
			}
			packet, err := tse.packets.DequeuePacket()
			// fmt.Println("dequeue", err)
			if err != nil {
				tse.mutex.Unlock()
				continue
			}
			// fmt.Println("out")
			packetOutChan <- packet
			tse.mutex.Unlock()
		}
	}(cp)
	return cp
}

func (tse *TransportStreamEngine) Write(p []byte) (n int, err error) {
	inputBytes := len(p)
	tse.mutex.Lock()
	defer tse.mutex.Unlock()
	if len(tse.buf) == cap(tse.buf) {
		return 0, fmt.Errorf("bytebuffer full")
	}
	if inputBytes+len(tse.buf) > cap(tse.buf) {
		inputBytes = cap(tse.buf) - len(tse.buf)
	}

	tse.buf = append(tse.buf, p[:inputBytes-1]...)
	// fmt.Println("written", len(tse.buf))

	return inputBytes, nil
}
