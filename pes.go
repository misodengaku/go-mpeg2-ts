package mpeg2ts

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

const (
	// Rec. ITU-T H.222.0 (06/2021) pp.42-43
	StreamID_ProgramStreamMap       = 0xbc
	StreamID_PrivateStream1         = 0xbd
	StreamID_PaddingStream          = 0xbe
	StreamID_PrivateStream2         = 0xbf
	StreamID_ECM                    = 0xf0
	StreamID_EMM                    = 0xf1
	StreamID_DSMCC                  = 0xf2
	StreamID_ISO13522               = 0xf3
	StreamID_H222_1_TypeA           = 0xf4
	StreamID_H222_1_TypeB           = 0xf5
	StreamID_H222_1_TypeC           = 0xf6
	StreamID_H222_1_TypeD           = 0xf7
	StreamID_H222_1_TypeE           = 0xf8
	StreamID_AncillaryStream        = 0xf9
	StreamID_SLPacketizedStream     = 0xfa
	StreamID_FlexMuxStream          = 0xfb
	StreamID_MetadataStream         = 0xfc
	StreamID_ExtendedStreamID       = 0xfd
	StreamID_ReservedDataStream     = 0xfe
	StreamID_ProgramStreamDirectory = 0xff

	ScramblingControl_NotScrambled = 0x00
	ScramblingControl_UserDefined1 = 0x01
	ScramblingControl_UserDefined2 = 0x02
	ScramblingControl_UserDefined3 = 0x03
)

var (
	ErrAlreadyClosed          = errors.New("PESParser is already closed")
	ErrByteIncomingChanClosed = errors.New("byteIncomingChan is closed")
	ErrCanceled               = errors.New("canceled")
)

const (
	StateFindPrefix = iota
	StateParseOptPESHeader
	StateReadPacket
	StateReadBytes
	StateReadPaddingBytes
)

// Packetized Elementary Stream
// Rec. ITU-T H.222.0 (06/2021) pp.39-52
type PES struct {
	Prefix       uint32
	StreamID     byte
	PacketLength uint16
	ProgramClockReference

	// stream_id != program_stream_map && stream_id != padding_stream && stream_id != private_stream_2 && stream_id != ECM && stream_id != EMM && stream_id != program_stream_directory && stream_id != DSMCC_stream && stream_id != ITU-T Rec. H.222.1 type E stream
	ScramblingControl      byte
	Priority               bool
	DataAlignment          bool
	Copyright              bool
	Original               bool
	PTSFlag                bool
	DTSFlag                bool
	ESCRFlag               bool
	ESRateFlag             bool
	DSMTrickModeFlag       bool
	AdditionalCopyInfoFlag bool
	CRCFlag                bool
	ExtensionFlag          bool
	HeaderDataLength       byte
	rawPTS                 uint32
	rawDTS                 uint32
	PTS                    float64
	DTS                    float64
	ESCRBase               uint32
	ESCRExtension          uint16
	ESRate                 uint32
	ElementaryStream       []byte
	PacketDataStream       []byte
	Padding                []byte
	// DSM_trick_mode_flag == 1
}

type PESParser struct {
	packetCount      int
	buffer           []PESByte
	bufferSize       int
	byteIncomingChan chan []PESByte
	mutex            *sync.Mutex
	isClosed         bool
	statusMutex      *sync.Mutex
	ctx              context.Context
	PES
}

type PESByte struct {
	Datum         byte
	StartOfPacket bool
	EndOfStream   bool
}

func NewPESParser(ctx context.Context, bufferSize int) PESParser {
	pp := PESParser{packetCount: 0, bufferSize: bufferSize}
	pp.buffer = make([]PESByte, 0, pp.bufferSize)
	pp.byteIncomingChan = make(chan []PESByte, 128*1024)

	pp.mutex = &sync.Mutex{}
	pp.statusMutex = &sync.Mutex{}
	pp.isClosed = false
	pp.ctx = ctx
	return pp
}

func (pp *PESParser) receiveBytes() ([]PESByte, error) {
	var in []PESByte
	select {
	case <-pp.ctx.Done():
		return nil, ErrCanceled
	case w, ok := <-pp.byteIncomingChan:
		if !ok {
			return nil, ErrByteIncomingChanClosed
		}
		// enqueue bytes to parser queue
		in = make([]PESByte, len(w))
		copy(in, w)
		return in, nil
	}
}

func (pp *PESParser) findPrefix() bool {
	if pp.getBufferLength() < 6 {
		// not enough buffer
		return false
	}

	// find prefix bytes
	prefixIndex := -1
	for i := 0; i < pp.getBufferLength()-3; i++ {
		if pp.buffer[i].Datum == 0 && pp.buffer[i+1].Datum == 0 && pp.buffer[i+2].Datum == 1 {
			// does not match prefix pattern. skip 1byte
			prefixIndex = i
			pp.dequeue(prefixIndex)
			break
		}
	}
	if prefixIndex == -1 {
		// ran out of buffers. but no prefix pattern found
		pp.dequeue(pp.getBufferLength())
		return false
	}

	pp.PES.Prefix = uint32(pp.buffer[0].Datum)<<16 | uint32(pp.buffer[1].Datum)<<8 | uint32(pp.buffer[2].Datum)
	pp.PES.StreamID = pp.buffer[3].Datum
	pp.PES.PacketLength = uint16(pp.buffer[4].Datum)<<8 | uint16(pp.buffer[5].Datum)

	pp.dequeue(6)
	return true
}

func (pp *PESParser) readPaddingBytes() bool {
	if pp.getBufferLength() > int(pp.PES.PacketLength) {
		// not enough buffer
		return false
	}
	pp.PES.Padding = make([]byte, pp.PES.PacketLength)
	pp.mutex.Lock()
	for i, v := range pp.buffer[:pp.PES.PacketLength] {
		pp.PES.Padding[i] = v.Datum
	}
	pp.mutex.Unlock()
	pp.dequeue(int(pp.PES.PacketLength))
	return true
}

func (pp *PESParser) parseOptionalPESHeaders() {
	pp.PES.ScramblingControl = (pp.buffer[0].Datum >> 4) & 0x03
	pp.PES.Priority = (pp.buffer[0].Datum>>3)&0x01 == 1
	pp.PES.DataAlignment = (pp.buffer[0].Datum>>2)&0x01 == 1
	pp.PES.Copyright = (pp.buffer[0].Datum>>1)&0x01 == 1
	pp.PES.Original = (pp.buffer[0].Datum)&0x01 == 1
	pp.PES.PTSFlag = (pp.buffer[1].Datum>>7)&0x01 == 1
	pp.PES.DTSFlag = (pp.buffer[1].Datum>>6)&0x01 == 1
	pp.PES.ESCRFlag = (pp.buffer[1].Datum>>5)&0x01 == 1
	pp.PES.ESRateFlag = (pp.buffer[1].Datum>>4)&0x01 == 1
	pp.PES.DSMTrickModeFlag = (pp.buffer[1].Datum>>3)&0x01 == 1
	pp.PES.AdditionalCopyInfoFlag = (pp.buffer[1].Datum>>2)&0x01 == 1
	pp.PES.CRCFlag = (pp.buffer[1].Datum>>1)&0x01 == 1
	pp.PES.ExtensionFlag = (pp.buffer[1].Datum)&0x01 == 1
	pp.PES.HeaderDataLength = pp.buffer[2].Datum

	// PES header
	trimIndex := 2
	if pp.PTSFlag {
		// if (PTS_DTS_flags == '10') {
		pp.PES.rawPTS = uint32((pp.buffer[3].Datum>>1)&0x07)<<30 | uint32(pp.buffer[4].Datum)<<22 | uint32(pp.buffer[5].Datum>>1)<<15 | uint32(pp.buffer[6].Datum)<<7 | uint32(pp.buffer[7].Datum>>1)
		pp.PES.PTS = float64(pp.PES.rawPTS) / 90000 // 90kHz
		trimIndex += 5
	}
	if pp.DTSFlag {
		// if (PTS_DTS_flags == '11') {
		pp.PES.rawDTS = uint32((pp.buffer[8].Datum>>1)&0x07)<<30 | uint32(pp.buffer[9].Datum)<<22 | uint32(pp.buffer[10].Datum>>1)<<15 | uint32(pp.buffer[11].Datum)<<7 | uint32(pp.buffer[12].Datum>>1)
		pp.PES.DTS = float64(pp.PES.rawDTS) / 90000 // 90kHz
		trimIndex += 5
	}
	if pp.ESCRFlag {
		// pp.ESCRBase = pp.buffer
	}
	// ESRate
	// DSMtrick
	// additionalCopyInfo
	// CRC
	// PES extension

	pp.dequeue(trimIndex + 1)
}

func (pp *PESParser) StartPESReadLoop() <-chan PES {
	pc := make(chan PES, 16)
	go func(pesOutChan chan<- PES) {
		state := StateFindPrefix
		for {
			isLast := false

			in, err := pp.receiveBytes()
			if err != nil {
				close(pesOutChan)
				return
			}

			for i := 0; i < len(in); i++ {
				if in[i].EndOfStream {
					in = in[:i]
					isLast = true
					break
				}
			}
			pp.enqueue(in)

		ReadLoop:
			for pp.getBufferLength() > 0 {
				switch state {
				case StateFindPrefix:
					if ok := pp.findPrefix(); !ok {
						break ReadLoop
					}
					state = StateParseOptPESHeader
				case StateParseOptPESHeader:
					if pp.getBufferLength() < 13 {
						// not enough buffer
						break ReadLoop
					}
					switch pp.PES.StreamID {
					default:
						if (pp.buffer[0].Datum>>6)&0x03 != 0x02 {
							// invalid marker bits. reset
							pp.dequeue(1)
							state = StateFindPrefix
							break ReadLoop
						}

						pp.parseOptionalPESHeaders()
						state = StateReadPacket

					case StreamID_ProgramStreamMap:
						// contains only PES_packet_data_byte
						state = StateReadBytes
					case StreamID_PrivateStream2:
						// contains only PES_packet_data_byte
						state = StateReadBytes
					case StreamID_ECM:
						// contains only PES_packet_data_byte
						state = StateReadBytes
					case StreamID_EMM:
						// contains only PES_packet_data_byte
						state = StateReadBytes
					case StreamID_ProgramStreamDirectory:
						// contains only PES_packet_data_byte
						state = StateReadBytes
					case StreamID_DSMCC:
						// contains only PES_packet_data_byte
						state = StateReadBytes
					case StreamID_H222_1_TypeE:
						// contains only PES_packet_data_byte
						state = StateReadBytes
					case StreamID_PaddingStream:
						state = StateReadPaddingBytes
					}
				case StateReadPacket:
					// read payload
					if pp.getBufferLength() == 0 {
						// buffer is empty
						break ReadLoop
					}
					writtenBytes := 0

					pp.mutex.Lock()
					for _, v := range pp.buffer {
						if v.StartOfPacket {
							pr := pp.PES.DeepCopy()
							pesOutChan <- pr
							pp.PES = PES{}
							state = StateFindPrefix
							break
						}
						pp.PES.ElementaryStream = append(pp.PES.ElementaryStream, v.Datum)
						writtenBytes += 1
					}
					pp.dequeue(writtenBytes)
					pp.mutex.Unlock()

				case StateReadBytes:
					if pp.getBufferLength() < int(pp.PES.PacketLength) {
						// not enough buffer
						break ReadLoop
					}
					pp.PES.PacketDataStream = make([]byte, pp.PES.PacketLength)

					pp.mutex.Lock()
					for i, v := range pp.buffer[:pp.PES.PacketLength] {
						pp.PES.PacketDataStream[i] = v.Datum
					}
					pp.mutex.Unlock()
					fmt.Println(pp.PES.PacketLength)
					pp.dequeue(int(pp.PES.PacketLength))
					state = StateFindPrefix
				case StateReadPaddingBytes:
					if ok := pp.readPaddingBytes(); ok {
						state = StateFindPrefix
					}

				}
			}
			if isLast {
				close(pesOutChan)
				return
			}
		}
	}(pc)
	return pc
}

func (pp *PESParser) Close() {
	if pp.statusMutex == nil {
		return
	}
	pp.statusMutex.Lock()
	defer pp.statusMutex.Unlock()
	if pp.isClosed {
		return
	}

	close(pp.byteIncomingChan)
	pp.isClosed = true
}

func (pp *PESParser) dequeue(size int) []PESByte {
	r := make([]PESByte, size)
	if size > 0 {
		copy(r, pp.buffer[:size])
		pp.buffer = append(pp.buffer[:0], pp.buffer[size:]...)
	}
	return r
}

func (pp *PESParser) enqueue(in []PESByte) {
	pp.buffer = append(pp.buffer, in...)
}

func (pp *PESParser) getBufferLength() int {
	l := len(pp.buffer)
	return l
}

func (pp *PESParser) WriteBytes(p []byte, sop, eos bool) (n int, err error) {
	pp.statusMutex.Lock()
	if pp.isClosed {
		return 0, ErrAlreadyClosed
	}
	defer pp.statusMutex.Unlock()

	var b PESByte
	pesBytes := make([]PESByte, 0, len(p))
	for _, v := range p {
		b = PESByte{Datum: v}
		pesBytes = append(pesBytes, b)
	}
	pesBytes[0].StartOfPacket = sop
	pesBytes[len(p)-1].EndOfStream = eos
	pp.byteIncomingChan <- pesBytes
	return len(p), nil
}

func (pp *PESParser) EnqueueTSPacket(tsPacket Packet) error {
	byteBuffer, err := tsPacket.GetPayload()
	if err != nil {
		return err
	}
	_, err = pp.WriteBytes(byteBuffer, tsPacket.PayloadUnitStartIndicator, false)
	return err
}

func (pp *PESParser) EnqueueLastTSPacket(tsPacket Packet) error {
	byteBuffer, err := tsPacket.GetPayload()
	if err != nil {
		return err
	}
	_, err = pp.WriteBytes(byteBuffer, tsPacket.PayloadUnitStartIndicator, true)
	return err
}
