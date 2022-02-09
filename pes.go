package mpeg2ts

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
	// DSM_trick_mode_flag == 1
}

type PESParser struct {
	packetCount int
	PES
}

// type PESProgram struct {
// 	ProgramNumber uint16
// 	Reserved      int
// 	NetworkPID    uint16
// 	ProgramMapPID uint16
// }

func NewPESParser() PESParser {
	p := PESParser{packetCount: 0}
	return p
}

func (pp *PESParser) GetPES() (PES, error) {
	return pp.PES, nil
}

func (pp *PESParser) EnqueueTSPacket(tsPacket Packet) error {
	payload, err := tsPacket.GetPayload()
	if err != nil {
		return err
	}
	// fmt.Printf("eq: %#v\n", payload)
	if pp.packetCount == 0 {
		pp.PES.Prefix = uint32(payload[0])<<16 | uint32(payload[1])<<8 | uint32(payload[2])
		pp.PES.StreamID = payload[3]
		pp.PES.PacketLength = uint16(payload[4])<<8 | uint16(payload[5])

		pp.PES.ScramblingControl = (payload[6] >> 6) & 0x03
		pp.PES.Priority = (payload[6]>>3)&0x01 == 1
		pp.PES.DataAlignment = (payload[6]>>2)&0x01 == 1
		pp.PES.Copyright = (payload[6]>>1)&0x01 == 1
		pp.PES.Original = (payload[6])&0x01 == 1
		pp.PES.PTSFlag = (payload[7]>>7)&0x01 == 1
		pp.PES.DTSFlag = (payload[7]>>6)&0x01 == 1
		pp.PES.ESCRFlag = (payload[7]>>5)&0x01 == 1
		pp.PES.ESRateFlag = (payload[7]>>4)&0x01 == 1
		pp.PES.DSMTrickModeFlag = (payload[7]>>3)&0x01 == 1
		pp.PES.AdditionalCopyInfoFlag = (payload[7]>>2)&0x01 == 1
		pp.PES.CRCFlag = (payload[7]>>1)&0x01 == 1
		pp.PES.ExtensionFlag = (payload[7])&0x01 == 1
		pp.PES.HeaderDataLength = payload[8]

		if pp.PTSFlag {
			pp.PES.rawPTS = uint32((payload[9]>>1)&0x03)<<30 | uint32(payload[10])<<22 | uint32(payload[11]>>1)<<15 | uint32(payload[12])<<7 | uint32(payload[13]>>1)
			pp.PES.PTS = float64(pp.PES.rawPTS) / 90000 // 90kHz
		}
		if pp.DTSFlag {
			pp.PES.rawDTS = uint32((payload[14]>>1)&0x03)<<30 | uint32(payload[15])<<22 | uint32(payload[16]>>1)<<15 | uint32(payload[17])<<7 | uint32(payload[18]>>1)
		}
	}
	pp.packetCount++
	return nil
}

// func (p *Packet) ParsePES() (PES, error) {
// 	pes := PES{}
// 	payload := p.GetPayload()
// 	pes.Pointer = payload[0]
// 	pes.TableID = payload[1]
// 	pes.SectionSyntaxIndicator = ((payload[2] >> 7) & 0x01) == 1
// 	if ((payload[2] >> 6) & 0x01) == 1 {
// 		return PES{}, fmt.Errorf("invalid format")
// 	}
// 	pes.SectionLength = uint16(payload[2]&0x0F)<<8 | uint16(payload[3])
// 	pes.TransportStreamID = uint16(payload[4])<<8 | uint16(payload[5])
// 	pes.Version = (payload[6] >> 1) & 0x1F
// 	pes.CurrentNextIndicator = (payload[6] & 0x01) == 0x01
// 	pes.SectionNumber = payload[7]
// 	pes.LastSectionNumber = payload[8]

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
// 		pes.Programs[i].ProgramNumber = uint16(payload[base])<<8 | uint16(payload[base+1])
// 		if pes.Programs[i].ProgramNumber == 0x0000 {
// 			pes.Programs[i].NetworkPID = uint16(payload[base+2]&0x1f)<<8 | uint16(payload[base+3])&0x1fff
// 		} else {
// 			pes.Programs[i].ProgramMapPID = uint16(payload[base+2]&0x1f)<<8 | uint16(payload[base+3])&0x1fff
// 		}
// 		// pes.Programs[i].ProgramMapPID = uint16(payload[base+2]&0x1f)<<8 | uint16(payload[base+3])
// 	}
// 	// fmt.Printf("CRC32 dump: %02x %02x %02x %02x\r\n", uint(payload[pes.SectionLength]), uint(payload[pes.SectionLength+1]), uint(payload[pes.SectionLength+2]), uint(payload[pes.SectionLength+3]))
// 	pes.CRC32 = uint(payload[pes.SectionLength])<<24 | uint(payload[pes.SectionLength+1])<<16 | uint(payload[pes.SectionLength+2])<<8 | uint(payload[pes.SectionLength+3])
// 	// fmt.Printf("%#v\r\n", payload[1:pes.SectionLength])

// 	crc := CalculateCRC(0, payload[1:pes.SectionLength]) ^ 0xffffffff
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
