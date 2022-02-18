package mpeg2ts

import (
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
	PES
}

type PESByte struct {
	Datum         byte
	StartOfPacket bool
}

func NewPESParser(bufferSize int) PESParser {
	pp := PESParser{packetCount: 0, bufferSize: bufferSize}
	pp.buffer = make([]PESByte, 0, pp.bufferSize)
	pp.byteIncomingChan = make(chan []PESByte, 1048576)
	pp.mutex = &sync.Mutex{}
	return pp
}

func (pp *PESParser) StartPESReadLoop() chan PES {
	pc := make(chan PES)
	go func(pesOutChan chan PES) {
		state := 0
		for w := range pp.byteIncomingChan {
			in := make([]PESByte, len(w))
			copy(in, w)
			pp.enqueue(in)
			eor := false
			for pp.getBufferLength() > 0 && !eor {
				if state == 0 {
					if pp.getBufferLength() < 6 {
						// buffer is too short
						eor = true
						break
					}
					prefixIndex := -1
					for i := 0; i < pp.getBufferLength()-3; i++ {
						if pp.buffer[i].Datum == 0 && pp.buffer[i+1].Datum == 0 && pp.buffer[i+2].Datum == 1 {
							prefixIndex = i
							eor = true
							break
						}
					}
					if prefixIndex == -1 {
						// sync byte is not found
						pp.dequeue(pp.getBufferLength())
						eor = true
						break
					} else {
						pp.dequeue(prefixIndex)
					}
					pp.PES.Prefix = uint32(pp.buffer[0].Datum)<<16 | uint32(pp.buffer[1].Datum)<<8 | uint32(pp.buffer[2].Datum)
					pp.PES.StreamID = pp.buffer[3].Datum
					pp.PES.PacketLength = uint16(pp.buffer[4].Datum)<<8 | uint16(pp.buffer[5].Datum)

					pp.dequeue(6)
					state = 1
				}

				if state == 1 {
					if pp.getBufferLength() < 13 {
						// buffer is too short
						eor = true
						break
					}
					switch pp.PES.StreamID {
					default:
						if (pp.buffer[0].Datum>>6)&0x03 != 0x02 {
							// invalid. reset
							pp.dequeue(1)
							state = 0
							eor = true
							break
						}

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
						state = 2

					case StreamID_ProgramStreamMap:
						fallthrough
					case StreamID_PrivateStream2:
						fallthrough
					case StreamID_ECM:
						fallthrough
					case StreamID_EMM:
						fallthrough
					case StreamID_ProgramStreamDirectory:
						fallthrough
					case StreamID_DSMCC:
						fallthrough
					case StreamID_H222_1_TypeE:
						state = 3
					case StreamID_PaddingStream:
						state = 4
					}

				}

				if state == 2 {
					// read payload
					if pp.getBufferLength() == 0 {
						// buffer is empty
						eor = true
						break
					}
					writtenBytes := 0

					pp.mutex.Lock()
					for _, v := range pp.buffer {
						if v.StartOfPacket {
							pr := pp.PES.DeepCopy()
							pc <- pr
							pp.PES = PES{}
							state = 0
							break
						}
						pp.PES.ElementaryStream = append(pp.PES.ElementaryStream, v.Datum)
						writtenBytes += 1
					}
					pp.dequeue(writtenBytes)
					pp.mutex.Unlock()

				}

				if state == 3 {
					if pp.getBufferLength() < int(pp.PES.PacketLength) {
						// buffer is too short
						eor = true
						break
					}
					pp.PES.PacketDataStream = make([]byte, pp.PES.PacketLength)

					pp.mutex.Lock()
					for i, v := range pp.buffer[:pp.PES.PacketLength] {
						pp.PES.PacketDataStream[i] = v.Datum
					}
					pp.mutex.Unlock()
					fmt.Println(pp.PES.PacketLength)
					pp.dequeue(int(pp.PES.PacketLength))
					state = 0
				}

				if state == 4 {
					if pp.getBufferLength() > int(pp.PES.PacketLength) {
						// buffer is too short
						eor = true
						break
					}
					pp.PES.Padding = make([]byte, pp.PES.PacketLength)
					pp.mutex.Lock()
					for i, v := range pp.buffer[:pp.PES.PacketLength] {
						pp.PES.Padding[i] = v.Datum
					}
					pp.mutex.Unlock()
					pp.dequeue(int(pp.PES.PacketLength))
					state = 0
				}

			}
		}
	}(pc)
	return pc
}

func (pp *PESParser) dequeue(size int) []PESByte {
	var r []PESByte
	if size > 0 {
		r = pp.buffer[:size]
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

var i = 0

func (pp *PESParser) WriteBytes(p []byte, sop bool) (n int, err error) {
	var b PESByte
	pesBytes := make([]PESByte, 0, len(p))
	for _, v := range p {
		b = PESByte{Datum: v}
		pesBytes = append(pesBytes, b)
	}
	pesBytes[0].StartOfPacket = sop
	pp.byteIncomingChan <- pesBytes
	i++
	return len(p), nil
}

func (pp *PESParser) EnqueueTSPacket(tsPacket Packet) error {
	byteBuffer, err := tsPacket.GetPayload()
	if err != nil {
		return err
	}
	_, err = pp.WriteBytes(byteBuffer, tsPacket.PayloadUnitStartIndicator)
	return err
}
