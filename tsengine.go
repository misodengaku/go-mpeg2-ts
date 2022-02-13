package mpeg2ts

import (
	"fmt"
)

type TransportStreamEngine struct {
	buf              []byte
	bufferSize       int
	packets          PacketList
	chunkSize        int
	byteIncomingChan chan []byte
}

func InitTSEngine(chunkSize, bufferSize int) (TransportStreamEngine, error) {
	tse := TransportStreamEngine{}
	tse.bufferSize = bufferSize
	tse.buf = make([]byte, 0, tse.bufferSize)
	tse.chunkSize = chunkSize
	tse.packets, _ = NewPacketList(chunkSize)
	tse.byteIncomingChan = make(chan []byte)
	return tse, nil
}

func (tse *TransportStreamEngine) StartPacketReadLoop() chan Packet {
	cp := make(chan Packet)
	go func(packetOutChan chan Packet) {
		for in := range tse.byteIncomingChan {
			tse.buf = append(tse.buf, in...)
			fmt.Println("wait", len(tse.buf))
			for len(tse.buf) >= tse.chunkSize {
				syncIndex := -1
				for i, v := range tse.buf {
					if v == 0x47 {
						syncIndex = i
						break
					}
				}
				if syncIndex == -1 {
					// tse.buf is dirty. clear and continue
					tse.buf = tse.buf[:0]
					continue
				} else if syncIndex > 0 {
					dirty := tse.buf[:syncIndex]
					fmt.Printf("trim %#v\n", dirty)
					tse.buf = tse.buf[syncIndex:]
				}

				err := tse.packets.AddBytes(tse.buf[:tse.chunkSize], tse.chunkSize)
				if err != nil {
					continue
				}
				packet, err := tse.packets.DequeuePacket()
				if err != nil {
					continue
				}
				packetOutChan <- packet
				if len(tse.buf) > tse.chunkSize {
					tse.buf = append(tse.buf[:0], tse.buf[tse.chunkSize:]...)
				}
			}
		}
	}(cp)
	return cp
}

func (tse *TransportStreamEngine) Write(p []byte) (n int, err error) {
	// inputBytes := len(p)
	// tse.mutex.Lock()
	// defer tse.mutex.Unlock()
	// if len(tse.buf) == cap(tse.buf) {
	// 	return 0, fmt.Errorf("bytebuffer full")
	// }
	// if inputBytes+len(tse.buf) > cap(tse.buf) {
	// 	inputBytes = cap(tse.buf) - len(tse.buf)
	// }

	// tse.buf = append(tse.buf, p[:inputBytes]...)
	tse.byteIncomingChan <- p //[:inputBytes]
	// fmt.Println("written", len(tse.buf))

	return len(p), nil
}
