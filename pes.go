package mpeg2ts

import (
	"fmt"
	"sync"
	"time"
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
	// DSM_trick_mode_flag == 1
}

type PESParser struct {
	packetCount int
	buffer      []PESByte
	bufferSize  int
	mutex       *sync.Mutex
	PES
}

type PESByte struct {
	Datum         byte
	StartOfPacket bool
}

func NewPESParser(bufferSize int) PESParser {
	p := PESParser{packetCount: 0, bufferSize: bufferSize}
	p.buffer = make([]PESByte, 0, p.bufferSize)
	p.mutex = &sync.Mutex{}
	return p
}

func (pp *PESParser) StartPESReadLoop() chan PES {
	pc := make(chan PES)
	go func(pesOutChan chan PES) {
		state := 0
		for {
			if state == 0 {
				pp.mutex.Lock()
				if len(pp.buffer) < 6 {
					pp.mutex.Unlock()
					time.Sleep(1 * time.Microsecond)
					continue
				}
				prefixIndex := -1
				for i := 0; i < len(pp.buffer)-3; i++ {
					if pp.buffer[i].Datum == 0 && pp.buffer[i+1].Datum == 0 && pp.buffer[i+2].Datum == 1 {
						prefixIndex = i
						break
					}
				}
				if prefixIndex == -1 {
					pp.mutex.Unlock()
					time.Sleep(1 * time.Microsecond)
					continue
				} else {
					pp.buffer = pp.buffer[prefixIndex:]
				}
				pp.PES.Prefix = uint32(pp.buffer[0].Datum)<<16 | uint32(pp.buffer[1].Datum)<<8 | uint32(pp.buffer[2].Datum)
				pp.PES.StreamID = pp.buffer[3].Datum
				pp.PES.PacketLength = uint16(pp.buffer[4].Datum)<<8 | uint16(pp.buffer[5].Datum)

				pp.buffer = pp.buffer[6:]
				state = 1
				pp.mutex.Unlock()
			}

			if state == 1 {
				pp.mutex.Lock()
				if len(pp.buffer) < 13 {
					pp.mutex.Unlock()
					time.Sleep(1 * time.Microsecond)
					continue
				}
				switch pp.PES.StreamID {
				default:
					if (pp.buffer[0].Datum>>6)&0x03 != 0x02 {
						// invalid. reset
						pp.buffer = pp.buffer[1:]
						state = 0
						pp.mutex.Unlock()
						continue
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

					pp.buffer = pp.buffer[trimIndex+1:]
					state = 2
					pp.mutex.Unlock()

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
					// if stream_id == program_stream_map || stream_id == private_stream_2 || stream_id == ECM || stream_id == EMM || stream_id == program_stream_directory || stream_id == DSMCC_stream || stream_id == "ITU-T Rec. H.222.1 type E stream"
					// for i := 0; i < PES_packet_length; i++ {
					// PES_packet_data_byte 8 bslbf
					// }
				case StreamID_PaddingStream:
					// for i < 0; i < PES_packet_length; i++ {
					// padding_byte 8 bslbf
					// }
				}
			}

			if state == 2 {
				// read payload
				pp.mutex.Lock()
				if len(pp.buffer) == 0 {
					pp.mutex.Unlock()
					time.Sleep(1 * time.Microsecond)
					continue
				}
				writtenBytes := 0
				for _, v := range pp.buffer {
					if v.StartOfPacket {
						pc <- pp.PES
						pp.PES = PES{}
						state = 0
						break
					}
					pp.PES.ElementaryStream = append(pp.PES.ElementaryStream, v.Datum)
					writtenBytes += 1
				}
				newBuffer := make([]PESByte, 0, pp.bufferSize)
				copy(newBuffer, pp.buffer[writtenBytes:])
				pp.buffer = newBuffer
				pp.mutex.Unlock()
			}
		}
	}(pc)
	return pc
}

func (pp *PESParser) WriteBytes(p []byte, sop bool) (n int, err error) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	inputBytes := len(p)
	if len(pp.buffer) == cap(pp.buffer) {
		return 0, fmt.Errorf("bytebuffer full")
	}
	if inputBytes+len(pp.buffer) > cap(pp.buffer) {
		inputBytes = cap(pp.buffer) - len(pp.buffer)
	}

	pesBytes := make([]PESByte, 0, inputBytes)
	for _, v := range p[:inputBytes] {
		b := PESByte{Datum: v}
		pesBytes = append(pesBytes, b)
	}
	pesBytes[0].StartOfPacket = sop
	pp.buffer = append(pp.buffer, pesBytes...)
	// fmt.Printf("len: %d cap: %d\n", len(pp.buffer), cap(pp.buffer))
	return inputBytes, nil
}

func (pp *PESParser) EnqueueTSPacket(tsPacket Packet) error {
	byteBuffer, err := tsPacket.GetPayload()
	if err != nil {
		return err
	}
	_, err = pp.WriteBytes(byteBuffer, tsPacket.PayloadUnitStartIndicator)
	return err
}
