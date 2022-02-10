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
	PES
	buffer []PESByte
	mutex  *sync.Mutex
}

type PESByte struct {
	Datum       byte
	EndOfPacket bool
}

func NewPESParser() PESParser {
	p := PESParser{packetCount: 0}
	p.buffer = make([]PESByte, 0, 1048576)
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
						pp.PES.rawPTS = uint32((pp.buffer[3].Datum>>1)&0x03)<<30 | uint32(pp.buffer[4].Datum)<<22 | uint32(pp.buffer[5].Datum>>1)<<15 | uint32(pp.buffer[6].Datum)<<7 | uint32(pp.buffer[7].Datum>>1)
						pp.PES.PTS = float64(pp.PES.rawPTS) / 90000 // 90kHz
						trimIndex += 5
					}
					if pp.DTSFlag {
						// if (PTS_DTS_flags == '11') {
						pp.PES.rawDTS = uint32((pp.buffer[8].Datum>>1)&0x03)<<30 | uint32(pp.buffer[9].Datum)<<22 | uint32(pp.buffer[10].Datum>>1)<<15 | uint32(pp.buffer[11].Datum)<<7 | uint32(pp.buffer[12].Datum>>1)
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
				for i, v := range pp.buffer {
					pp.PES.ElementaryStream = append(pp.PES.ElementaryStream, v.Datum)
					writtenBytes = i + 1
					if v.EndOfPacket {
						pc <- pp.PES
						pp.PES = PES{}
						state = 0
						break
					}
				}
				pp.buffer = pp.buffer[writtenBytes:]
				pp.mutex.Unlock()
			}
		}
	}(pc)
	return pc
}

func (pp *PESParser) WriteBytes(p []byte, eop bool) (n int, err error) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	inputBytes := len(p)
	if len(pp.buffer) == cap(pp.buffer) {
		return 0, fmt.Errorf("bytebuffer full")
	}
	if inputBytes+len(pp.buffer) > cap(pp.buffer) {
		inputBytes = cap(pp.buffer) - len(pp.buffer)
	}

	pesBytes := []PESByte{}
	for _, v := range p[:inputBytes] {
		b := PESByte{Datum: v}
		pesBytes = append(pesBytes, b)
	}
	pesBytes[inputBytes-1].EndOfPacket = eop
	pp.buffer = append(pp.buffer, pesBytes...)
	return inputBytes, nil
}

func (pp *PESParser) EnqueueTSPacket(tsPacket Packet, eop bool) error {
	byteBuffer, err := tsPacket.GetPayload()
	if err != nil {
		return err
	}
	_, err = pp.WriteBytes(byteBuffer, eop)
	return err
}

// func (p *Packet) ParsePES() (PES, error) {
// 	pes := PES{}
// 	pp.buffer := p.GetPayload()
// 	pes.Pointer = pp.buffer[0]
// 	pes.TableID = pp.buffer[1]
// 	pes.SectionSyntaxIndicator = ((pp.buffer[2] >> 7) & 0x01) == 1
// 	if ((pp.buffer[2] >> 6) & 0x01) == 1 {
// 		return PES{}, fmt.Errorf("invalid format")
// 	}
// 	pes.SectionLength = uint16(pp.buffer[2]&0x0F)<<8 | uint16(pp.buffer[3])
// 	pes.TransportStreamID = uint16(pp.buffer[4])<<8 | uint16(pp.buffer[5])
// 	pes.Version = (pp.buffer[6] >> 1) & 0x1F
// 	pes.CurrentNextIndicator = (pp.buffer[6] & 0x01) == 0x01
// 	pes.SectionNumber = pp.buffer[7]
// 	pes.LastSectionNumber = pp.buffer[8]

// 	// fmt.Printf("pes %d dump %x %x %t %d \r\n", p.Index, pes.Pointer, pes.TableID, pes.SectionSyntaxIndicator, pes.SectionLength)
// 	// for i := 0; i < 188; i++ {
// 	// 	fmt.Printf("%02x ", p.Data[i])
// 	// }
// 	// fmt.Println()

// 	// SectionLength - 5 - 4
// 	// header, _ := p.GetHeader()
// 	// fmt.Printf("%#v\r\n", header)
// 	pes.Programs = make([]PESProgram, (pes.SectionLength-5-4)/4)
// 	for i := uint16(0); i < (pes.SectionLength-5-4)/4; i++ {
// 		base := 9 + i*4
// 		pes.Programs[i].ProgramNumber = uint16(pp.buffer[base])<<8 | uint16(pp.buffer[base+1])
// 		if pes.Programs[i].ProgramNumber == 0x0000 {
// 			pes.Programs[i].NetworkPID = uint16(pp.buffer[base+2]&0x1f)<<8 | uint16(pp.buffer[base+3])&0x1fff
// 		} else {
// 			pes.Programs[i].ProgramMapPID = uint16(pp.buffer[base+2]&0x1f)<<8 | uint16(pp.buffer[base+3])&0x1fff
// 		}
// 		// pes.Programs[i].ProgramMapPID = uint16(pp.buffer[base+2]&0x1f)<<8 | uint16(pp.buffer[base+3])
// 	}
// 	// fmt.Printf("CRC32 dump: %02x %02x %02x %02x\r\n", uint(pp.buffer[pes.SectionLength]), uint(pp.buffer[pes.SectionLength+1]), uint(pp.buffer[pes.SectionLength+2]), uint(pp.buffer[pes.SectionLength+3]))
// 	pes.CRC32 = uint(pp.buffer[pes.SectionLength])<<24 | uint(pp.buffer[pes.SectionLength+1])<<16 | uint(pp.buffer[pes.SectionLength+2])<<8 | uint(pp.buffer[pes.SectionLength+3])
// 	// fmt.Printf("%#v\r\n", pp.buffer[1:pes.SectionLength])

// 	crc := CalculateCRC(0, pp.buffer[1:pes.SectionLength]) ^ 0xffffffff
// 	if uint32(pes.CRC32) != crc {
// 		return PES{}, fmt.Errorf("CRC32 mismatch")
// 	}

// 	// fmt.Println("CRC OK")
// 	return pes, nil
// }

// // based on isal's crc32 algo found at:
// // https://github.com/01org/isa-l/blob/master/crc/crc_base.c#L138-L155
// func CalculateCRC(seed uint32, data []byte) (crc uint32) {
// 	rem := uint64(^seed)

// 	var i, j int

// 	const (
// 		// defined in
// 		// https://github.com/01org/isa-l/blob/master/crc/crc_base.c#L33
// 		MAX_ITER = 8
// 	)

// 	for i = 0; i < len(data); i++ {
// 		rem = rem ^ (uint64(data[i]) << 24)
// 		for j = 0; j < MAX_ITER; j++ {
// 			rem = rem << 1
// 			if (rem & 0x100000000) != 0 {
// 				rem ^= uint64(0x04C11DB7)
// 			}
// 		}
// 	}

// 	crc = uint32(^rem)
// 	return
// }
