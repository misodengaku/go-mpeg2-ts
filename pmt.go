package mpeg2ts

import (
	"fmt"
)

const (
	// StreamTypeReserved       = 0x00 // ITU-T | ISO/IEC Reserved
	// StreamTypeISO11172_2_Video       = 0x01 // ISO/IEC 11172-2 Video
	// StreamTypeISO13818_2_Video       = 0x02 // Rec. ITU-T H.262 | ISO/IEC 13818-2 Video or ISO/IEC 11172-2 constrained parameter video stream
	// StreamTypeISO11172_3_Audio       = 0x03 // ISO/IEC 11172-3 Audio
	// StreamTypeISO13818_3_Audio       = 0x04 // ISO/IEC 13818-3 Audio
	// StreamTypeISO13818_1_PrivateSections       = 0x05 // Rec. ITU-T H.222.0 | ISO/IEC 13818-1 private_sections
	// StreamTypeISO13818_1_PES       = 0x06 // Rec. ITU-T H.222.0 | ISO/IEC 13818-1 PES packets containing private data
	// StreamTypeISO13522_MHEG       = 0x07 // ISO/IEC 13522 MHEG
	// StreamTypeISO13818_AnnexA       = 0x08 // Rec. ITU-T H.222.0 | ISO/IEC 13818-1 Annex A DSM-CC
	// StreamTypeH222_1       = 0x09 // Rec. ITU-T H.222.1
	// StreamTypeISO13818_6_TypeA       = 0x0A // ISO/IEC 13818-6 type A
	// StreamTypeISO13818_6_TypeB       = 0x0B // ISO/IEC 13818-6 type B
	// StreamTypeISO13818_6_TypeC       = 0x0C // ISO/IEC 13818-6 type C
	// StreamTypeISO13818_6_TypeD       = 0x0D // ISO/IEC 13818-6 type D
	// StreamTypeISO13818_1_Aux       = 0x0E // Rec. ITU-T H.222.0 | ISO/IEC 13818-1 auxiliary
	// StreamTypeISO13818_7_Audio       = 0x0F // ISO/IEC 13818-7 Audio with ADTS transport syntax
	// StreamTypeISO14496_2_Visual       = 0x10 // ISO/IEC 14496-2 Visual
	// StreamTypeISO14496_3_Audio       = 0x11 // ISO/IEC 14496-3 Audio with the LATM transport syntax as defined in ISO/IEC 14496-3
	// StreamTypeISO14496_1_PES       = 0x12 // ISO/IEC 14496-1 SL-packetized stream or FlexMux stream carried in PES packets
	// StreamTypeISO14496_1_Sections       = 0x13 // ISO/IEC 14496-1 SL-packetized stream or FlexMux stream carried in ISO/IEC 14496_sections
	// StreamTypeISO13818_6_SDP       = 0x14 // ISO/IEC 13818-6 Synchronized Download Protocol
	// StreamTypeMetadataInPES       = 0x15 // Metadata carried in PES packets
	// StreamTypeMetadataInSections       = 0x16 // Metadata carried in metadata_sections
	// StreamTypeMetadataInDataCarousel       = 0x17 // Metadata carried in ISO/IEC 13818-6 Data Carousel
	// StreamTypeMetadataInObjectCarousel       = 0x18 // Metadata carried in ISO/IEC 13818-6 Object Carousel
	// StreamTypeMetadataInSDP       = 0x19 // Metadata carried in ISO/IEC 13818-6 Synchronized Download Protocol
	// StreamTypeIPMP       = 0x1A // IPMP stream (defined in ISO/IEC 13818-11, MPEG-2 IPMP)
	StreamTypeAVC = 0x1B // AVC video stream conforming to one or more profiles defined in Annex A of Rec. ITU-T H.264 |
	// StreamTypeISO14496_3_Audio       = 0x1C // ISO/IEC 14496-3 Audio, without using any additional transport syntax, such as DST, ALS and SLS
	// StreamType       = 0x1D // ISO/IEC 14496-17 Text
	// StreamType       = 0x1E // Auxiliary video stream as defined in ISO/IEC 23002-3
	// StreamType       = 0x1F // SVC video sub-bitstream of an AVC video stream conforming to one or more profiles defined in Annex G of Rec. ITU-T H.264 | ISO/IEC 14496-10
	// StreamType       = 0x20 // MVC video sub-bitstream of an AVC video stream conforming to one or more profiles defined in Annex H of Rec. ITU-T H.264 | ISO/IEC 14496-10
	// StreamType       = 0x21 // Video stream conforming to one or more profiles as defined in Rec. ITU-T T.800 | ISO/IEC 15444-1
	// StreamType       = 0x22 // Additional view Rec. ITU-T H.262 | ISO/IEC 13818-2 video stream for service-compatible stereoscopic 3D services (see Notes 3 and 4)
	// StreamType       = 0x23 // Additional view Rec. ITU-T H.264 | ISO/IEC 14496-10 video stream conforming to one or more profiles defined in Annex A for service-compatible stereoscopic 3D services (see Notes 3 and 4)
	// StreamType       = 0x24 // Rec. ITU-T H.265 | ISO/IEC 23008-2 video stream or an HEVC temporal video sub-bitstream
	// StreamType       = 0x25 // HEVC temporal video subset of an HEVC video stream conforming to one or more profiles defined in Annex A of Rec. ITU-T H.265 | ISO/IEC 23008-2
	// StreamType       = 0x26 // MVCD video sub-bitstream of an AVC video stream conforming to one or more profiles defined in Annex I of Rec. ITU-T H.264 | ISO/IEC 14496-10
	// StreamType       = 0x27 // Timeline and External Media Information Stream (see Annex U)
	// StreamType       = 0x28 // HEVC enhancement sub-partition which includes TemporalId 0 of an HEVC video stream where all NALs units contained in the stream conform to one or more profiles defined in Annex G of Rec. ITU-T H.265 | ISO/IEC 23008-2
	// StreamType       = 0x29 // HEVC temporal enhancement sub-partition of an HEVC video stream where all NAL units contained in the stream conform to one or more profiles defined in Annex G of Rec. ITU-T H.265 | ISO/IEC 23008-2
	// StreamType       = 0x2A // HEVC enhancement sub-partition which includes TemporalId 0 of an HEVC video stream where all NAL units contained in the stream conform to one or more profiles defined in Annex H of Rec. ITU-T H.265 | ISO/IEC 23008-2
	// StreamType       = 0x2B // HEVC temporal enhancement sub-partition of an HEVC video stream where all NAL units contained in the stream conform to one or more profiles defined in Annex H of Rec. ITU-T H.265 | ISO/IEC 23008-2
	// StreamType       = 0x2C // Green access units carried in MPEG-2 sections
	// StreamType       = 0x2D // ISO/IEC 23008-3 Audio with MHAS transport syntax – main stream
	// StreamType       = 0x2E // ISO/IEC 23008-3 Audio with MHAS transport syntax – auxiliary stream
	// StreamType       = 0x2F // Quality access units carried in sections
	// StreamType       = 0x30 // Media Orchestration Access Units carried in sections
	// StreamType       = 0x31 // Substream of a Rec. ITU-T H.265 | ISO/IEC 23008 2 video stream that contains a Motion Constrained Tile Set, parameter sets, slice headers or a combination thereof. See 2.17.5.1.
	// StreamType       = 0x32 // JPEG XS video stream conforming to one or more profiles as defined in ISO/IEC 21122-2
	// StreamType       = 0x33 // VVC video stream or a VVC temporal video sub-bitstream conforming to one or more profiles defined in Annex A of Rec. ITU-T H.266 | ISO/IEC 23090-3
	// StreamType       = 0x34 // VVC temporal video subset of a VVC video stream conforming to one or more profiles defined in Annex A of Rec. ITU-T H.266 | ISO/IEC 23090-3
	// StreamType       = 0x35 // EVC video stream or an EVC temporal video sub-bitstream conforming to one or more profiles defined in ISO/IEC 23094-1
	// StreamType       = 0x36 // .. 0x7E Rec. ITU-T H.222.0 | ISO/IEC 13818-1 reserved
	// StreamType       = 0x7F // IPMP stream
	// StreamType       = 0x80 // .. 0xFF User Private
)

// Program Map Table
// Rec. ITU-T H.222.0 (06/2021) p.57
type PMT struct {
	Pointer                byte
	TableID                byte
	SectionSyntaxIndicator bool
	Reserved1              byte
	SectionLength          uint16
	ProgramNumber          uint16
	Reserved2              byte
	Version                byte
	CurrentNextIndicator   bool
	SectionNumber          byte
	LastSectionNumber      byte
	Reserved3              byte
	PCR_PID                uint16
	Reserved4              byte
	ProgramInfoLength      uint16
	Descriptors            []ProgramElementDescriptor
	Streams                []StreamInfo
	CRC32                  uint
}

type StreamInfo struct {
	Type          byte
	Reserved1     byte
	ElementaryPID uint16
	Reserved2     byte
	ESInfoLength  uint16
	Descriptors   []ProgramElementDescriptor
}

type ProgramElementDescriptor struct {
	Tag    uint8
	Length uint8

	VideoStreamDescriptor
	RegistrationDescriptor
	UserPrivateDescriptor
}

type VideoStreamDescriptor struct {
	MultipleFrameRateFlag    bool
	FrameRateCode            uint8
	MPEG1OnlyFlag            bool
	ConstrainedParameterFlag bool
	StillPictureFlag         bool

	// MPEG1 Only
	ProfileAndLevelIndication uint8
	ChromaFormat              uint8
	FrameRateExtensionFlag    bool
	Reserved                  uint8
}

type RegistrationDescriptor struct {
	FormatIdentifier []byte
}
type UserPrivateDescriptor struct {
	Data []byte
}

func (p *Packet) ParsePMT() (PMT, error) {
	pmt := PMT{}
	payload, err := p.GetPayload()
	if err != nil {
		return PMT{}, err
	}
	// fmt.Printf("raw pmt dump %#v\r\n", payload)
	pmt.Pointer = payload[0]
	pmt.TableID = payload[1]                                     // 8
	pmt.SectionSyntaxIndicator = ((payload[2] >> 7) & 0x01) == 1 // 1
	if ((payload[2] >> 6) & 0x01) == 1 {                         // 1
		return PMT{}, fmt.Errorf("invalid format")
	}
	pmt.Reserved1 = (payload[2] >> 4) & 0x03                            // 2
	pmt.SectionLength = uint16(payload[2]&0x0F)<<8 | uint16(payload[3]) // 12
	pmt.ProgramNumber = uint16(payload[4])<<8 | uint16(payload[5])      // 16
	pmt.Reserved2 = (payload[6] >> 6) & 0x03                            // 2
	pmt.Version = (payload[6] >> 1) & 0x1F                              // 5
	pmt.CurrentNextIndicator = (payload[6] & 0x01) == 0x01              // 1
	pmt.SectionNumber = payload[7]                                      // 8
	pmt.LastSectionNumber = payload[8]                                  // 8
	pmt.Reserved3 = (payload[9] >> 5) & 0x07                            // 3
	pmt.PCR_PID = uint16(payload[9]&0x1f)<<8 | uint16(payload[10])      // 13
	pmt.Reserved4 = (payload[11] >> 4) & 0x0f                           // 4
	pmt.ProgramInfoLength = uint16(payload[11]&0x0F)<<8 | uint16(payload[12])
	index := 13

	// fmt.Printf("pmt %d dump %x table:%x synind:%t len:%d pid:%d pil:%d\r\n", p.Index, pmt.Pointer, pmt.TableID, pmt.SectionSyntaxIndicator, pmt.SectionLength, pmt.PCR_PID, pmt.ProgramInfoLength)

	// fmt.Println("len", (13 + int(pmt.ProgramInfoLength)))
	var diff int
	pmt.Descriptors, diff = readDescriptor(payload, 13, int(pmt.ProgramInfoLength))
	index += diff

	// Stream Descriptor
	for index < int(pmt.SectionLength) {
		si := StreamInfo{}
		si.Type = payload[index]                                                       // 8
		si.Reserved1 = (payload[index+1] >> 5) & 0x07                                  // 3
		si.ElementaryPID = uint16(payload[index+1]&0x1f)<<8 | uint16(payload[index+2]) // 13
		si.Reserved2 = (payload[index+3] >> 4) & 0x0f                                  // 4
		si.ESInfoLength = uint16(payload[index+3]&0x0f)<<8 | uint16(payload[index+4])  // 12
		// fmt.Printf("si: %#v\n", si)
		index += 5
		si.Descriptors, diff = readDescriptor(payload, index, int(si.ESInfoLength))
		pmt.Streams = append(pmt.Streams, si)
		index += diff
	}
	pmt.CRC32 = uint(payload[index])<<24 | uint(payload[index+1])<<16 | uint(payload[index+2])<<8 | uint(payload[index+3])
	// fmt.Printf("crc: %08x\n", pmt.CRC32)

	crc := calculateCRC(payload[1:pmt.SectionLength])

	if uint32(pmt.CRC32) != crc {
		// fmt.Printf("CRC error\n")
		return PMT{}, fmt.Errorf("CRC32 mismatch")
	}
	// fmt.Printf("CRC OK\n")
	return pmt, nil
}

func readDescriptor(payload []byte, startIndex, length int) ([]ProgramElementDescriptor, int) {

	diff := 0
	peds := []ProgramElementDescriptor{}

	// fmt.Printf("piLen:%d\n", length)

	endIndex := startIndex + length
	for index := startIndex; index < endIndex; {
		// fmt.Printf("desc index:%d max:%d len:%d\n", index, (startIndex + length), endIndex)
		ped := ProgramElementDescriptor{}
		ped.Tag = payload[index]
		ped.Length = payload[index+1]

		diff += 2

		switch {
		case ped.Tag == 2: //video_stream_descriptor
			ped.VideoStreamDescriptor.MultipleFrameRateFlag = ((payload[index+2] >> 7) & 0x01) == 1    // 1
			ped.VideoStreamDescriptor.FrameRateCode = (payload[index+2] >> 3) & 0x0f                   // 4
			ped.VideoStreamDescriptor.MPEG1OnlyFlag = ((payload[index+2] >> 2) & 0x01) == 1            // 1
			ped.VideoStreamDescriptor.ConstrainedParameterFlag = ((payload[index+2] >> 1) & 0x01) == 1 // 1
			ped.VideoStreamDescriptor.StillPictureFlag = ((payload[index+2]) & 0x01) == 1              // 1
			diff += 1

			if ped.MPEG1OnlyFlag {
				ped.VideoStreamDescriptor.ProfileAndLevelIndication = payload[index+3]                   //8
				ped.VideoStreamDescriptor.ChromaFormat = (payload[index+4] >> 6) & 0x03                  // 2
				ped.VideoStreamDescriptor.FrameRateExtensionFlag = ((payload[index+4] >> 5) & 0x01) == 1 // 1
				ped.VideoStreamDescriptor.Reserved = (payload[index+4]) & 0x1f                           // 5
				diff += 2
			}

		case ped.Tag == 3: //audio_stream_descriptor
		case ped.Tag == 4: //hierarchy_descriptor
		case ped.Tag == 5: //registration_descriptor
			ped.RegistrationDescriptor.FormatIdentifier = payload[index+2 : int(ped.Length)+index+2]
			diff += int(ped.Length)
		case ped.Tag == 6: //data_stream_alignment_descriptor
		case ped.Tag == 7: //target_background_grid_descriptor
		case ped.Tag == 8: //Video_window_descriptor
		case ped.Tag == 9: //CA_descriptor
		case ped.Tag == 10: //ISO_639_language_descriptor
		case ped.Tag == 11: //System_clock_descriptor
		case ped.Tag == 12: //Multiplex_buffer_utilization_descriptor
		case ped.Tag == 13: //Copyright_descriptor
		case ped.Tag == 14: // Maximum_bitrate_descriptor
		case ped.Tag == 15: //Private_data_indicator_descriptor
		case ped.Tag == 16: //Smoothing_buffer_descriptor
		case ped.Tag == 1: // STD_descriptor
		case ped.Tag == 18: //IBP_descriptor
		case ped.Tag == 27: //MPEG-4_video_descriptor
		case ped.Tag == 28: //MPEG-4_audio_descriptor
		case ped.Tag == 29: //IOD_descriptor
		case ped.Tag == 30: // SL_descriptor
		case ped.Tag == 31: //FMC_descriptor
		case ped.Tag == 32: //External_ES_ID_descriptor
		case ped.Tag == 33: //MuxCode_descriptor
		case ped.Tag == 34: //FmxBufferSize_descriptor
		case ped.Tag == 35: // MultiplexBuffer_descriptor

		case ped.Tag == 0 || ped.Tag == 1: // reserved
		case ped.Tag >= 19 && ped.Tag <= 26: // Defined in ISO/IEC 13818-6
		case ped.Tag >= 36 && ped.Tag <= 63: // ITU-T Rec. H.222.0 | ISO/IEC 13818-1 Reserved
		case ped.Tag >= 64 && ped.Tag <= 255: //  User Private
			ped.UserPrivateDescriptor.Data = payload[index+2 : int(ped.Length)+index+2]
			diff += int(ped.Length)
		}

		// fmt.Printf("ped dump %#v\r\n", ped)
		index += diff
		peds = append(peds, ped)
	}
	return peds, diff
}
