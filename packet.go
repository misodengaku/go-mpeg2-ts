package mpeg2ts

import (
	"fmt"
	"sync"
)

const (
	ScrambleControl_NotScrambled = 0
	ScrambleControl_Userdefined1 = 1
	ScrambleControl_Userdefined2 = 2
	ScrambleControl_Userdefined3 = 3

	AdaptationField_Reserved                = 0
	AdaptationField_PayloadOnly             = 1
	AdaptationField_AdaptationFieldOnly     = 2
	AdaptationField_AdaptationFieldFollowed = 3
)

func NewPacketList(chunkSize int) (PacketList, error) {
	pl := PacketList{}
	pl.mutex = &sync.Mutex{}
	pl.chunkSize = chunkSize
	pl.packets = make([]Packet, 0, 1024)
	return pl, nil
}

func (ps *PacketList) DequeuePacket() (Packet, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	if len(ps.packets) == 0 {
		return Packet{}, fmt.Errorf("PacketList is empty")
	}
	packet := ps.packets[0]
	ps.packets = append(ps.packets[:0], ps.packets[1:]...)

	return packet, nil
}

func (ps *PacketList) All() []Packet {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	p := make([]Packet, len(ps.packets))
	copy(p, ps.packets)
	return p
}
func (ps *PacketList) AddPacket(p Packet) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.packets = append(ps.packets, p)
}

func (ps *PacketList) AddBytes(packetBytes []byte, packetSize int) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if len(packetBytes) != packetSize {
		return fmt.Errorf("packetBytes length and packetSize is not match. len(packetBytes) is %d", len(packetBytes))
	}
	index := len(ps.packets)
	p := Packet{}
	p.Data = make([]byte, PacketSizeDefault)
	copy(p.Data, packetBytes)
	p.Index = index
	// fmt.Printf("index: %d\n", index)
	err := p.parseHeader()
	if err != nil {
		return err
	}
	ps.packets = append(ps.packets, p)
	// if index == 0 {

	// 	fmt.Printf("%#v\n", p)
	// 	fmt.Printf("%#v\n", *ps)
	// 	fmt.Println("---------------------------------")
	// }
	return nil
}

func (p *Packet) GetHeader() ([]byte, error) {
	if p.Data == nil || len(p.Data) != PacketSizeDefault {
		return nil, fmt.Errorf("invalid header")
	}
	return p.Data[:4], nil
}

func (p *Packet) GetPayload() ([]byte, error) {
	if !p.isHeaderParsed {
		return nil, fmt.Errorf("execute parseHeader() first")
	}

	if len(p.Data) != PacketSizeDefault {
		return nil, fmt.Errorf("invalid data size")
	}

	// fmt.Printf("payload: %#v\n", p.Data[4+p.AdaptationField.Size+1:])
	if p.AdaptationField.Length > 0 {
		return p.Data[4+p.AdaptationField.Length+1:], nil
	}
	return p.Data[4:], nil
}

func (p *Packet) HasAdaptationField() bool {
	c := p.AdaptationFieldControl
	if c == AdaptationField_AdaptationFieldOnly || c == AdaptationField_AdaptationFieldFollowed {
		return true
	}
	return false
}

func (p *Packet) parseHeader() error {
	if p.Data[0] != 0x47 {
		return fmt.Errorf("invalid magic number %02X", p.Data[0])
	}

	p.SyncByte = p.Data[0]
	p.TransportErrorIndicator = ((p.Data[1] >> 7) & 0x01) == 1
	p.PayloadUnitStartIndicator = ((p.Data[1] >> 6) & 0x01) == 1
	p.TransportPriorityIndicator = ((p.Data[1] >> 5) & 0x01) == 1
	p.PID = PID((uint16(p.Data[1])&0x1F)<<8 | uint16(p.Data[2]))
	p.TransportScrambleControl = (p.Data[3] >> 6) & 0x03
	p.AdaptationFieldControl = (p.Data[3] >> 4) & 0x03
	p.ContinuityCheckIndex = (p.Data[3] & 0x0F)

	// adaptation field
	if p.HasAdaptationField() {
		af := AdaptationField{}
		af.Length = p.Data[4]
		if p.AdaptationFieldControl == AdaptationField_AdaptationFieldFollowed {
			if af.Length > 183 {
				return fmt.Errorf("AdaptationField.Length should not exceed 182bytes")
			}
		} else if p.AdaptationFieldControl == AdaptationField_AdaptationFieldOnly {
			if af.Length != 183 {
				return fmt.Errorf("AdaptationField.Length must be 182bytes")
			}
		}

		if p.AdaptationField.Length == 0 {
			p.AdaptationField = af
			p.isHeaderParsed = true
			return nil
		}

		af.DiscontinuityIndicator = ((p.Data[5] >> 7) & 0x01) == 1
		af.RandomAccessIndicator = ((p.Data[5] >> 6) & 0x01) == 1
		af.ESPriorityIndicator = ((p.Data[5] >> 5) & 0x01) == 1
		af.PCRFlag = ((p.Data[5] >> 4) & 0x01) == 1
		af.OPCRFlag = ((p.Data[5] >> 3) & 0x01) == 1
		af.SplicingPointFlag = ((p.Data[5] >> 2) & 0x01) == 1
		af.TransportPrivateDataFlag = ((p.Data[5] >> 1) & 0x01) == 1
		af.ExtensionFlag = (p.Data[5] & 0x01) == 1
		// fmt.Printf("af: %#v\n", af)
		// fmt.Printf("bytes: %#v\n", p.Data)

		fieldIndex := 6
		if af.PCRFlag {
			// program_clock_reference_base 33 uimsbf
			af.ProgramClockReference.Base = uint64(p.Data[fieldIndex])<<25 | uint64(p.Data[fieldIndex+1])<<17 | uint64(p.Data[fieldIndex+2])<<9 | uint64(p.Data[fieldIndex+3])<<1 | uint64(p.Data[fieldIndex+4])>>7&0x01
			// reserved 6 bslbf
			// program_clock_reference_extension 9 uimsbf
			af.ProgramClockReference.Extension = uint16(p.Data[fieldIndex+4]&0x01)<<8 | uint16(p.Data[fieldIndex+5])
			fieldIndex += 6
			// fmt.Printf("PCR(%d/%d): %x %x\n", fieldIndex-5, af.Size, af.ProgramClockReference.Base, af.ProgramClockReference.Extension)
		}
		if af.OPCRFlag {
			// original_program_clock_reference_base 33 uimsbf
			af.OriginalProgramClockReference.Base = uint64(p.Data[fieldIndex])<<25 | uint64(p.Data[fieldIndex+1])<<17 | uint64(p.Data[fieldIndex+2])<<9 | uint64(p.Data[fieldIndex+3])<<1 | uint64(p.Data[fieldIndex+4])>>7&0x01
			// reserved 6 bslbf
			// original_program_clock_reference_extension 9 uimsbf
			af.OriginalProgramClockReference.Extension = uint16(p.Data[fieldIndex+4]&0x01)<<8 | uint16(p.Data[fieldIndex+5])
			fieldIndex += 6
			// fmt.Println("[BUG] OPCR parsing is not implemented")
		}
		if af.SplicingPointFlag {
			// splice_countdown 8 tcimsbf
			af.SpliceCountdown = p.Data[fieldIndex]
			fieldIndex++
		}
		if af.TransportPrivateDataFlag {
			// transport_private_data_length 8 uimsbf
			// for (i = 0; i < transport_private_data_length; i++) {
			// 	private_data_byte 8 bslbf
			// }
			af.TransportPrivateData.Length = p.Data[fieldIndex]
			af.TransportPrivateData.Data = p.Data[fieldIndex+1 : fieldIndex+1+int(af.TransportPrivateData.Length)]
			fieldIndex += 1 + int(af.TransportPrivateData.Length)
		}
		if af.ExtensionFlag {
			fmt.Println("[BUG] AdaptationFieldExtension parsing is not implemented")
			fmt.Printf("af: %#v\n", af)
			// adaptation_field_extension_length 8 uimsbf
			// ltw_flag 1 bslbf
			// piecewise_rate_flag 1 bslbf
			// seamless_splice_flag 1 bslbf
			// af_descriptor_not_present_flag 1 bslbf
			// reserved 4 bslbf
			// if (ltw_flag = = '1') {
			// 	ltw_valid_flag 1 bslbf
			// 	ltw_offset 15 uimsbf
			// }
			// if (piecewise_rate_flag = = '1') {
			// 	reserved 2 bslbf
			// 	piecewise_rate 22 uimsbf
			// }
			// if (seamless_splice_flag = = '1') {
			// 	Splice_type 4 bslbf
			// 	DTS_next_AU[32..30] 3 bslbf
			// 	marker_bit 1 bslbf
			// 	DTS_next_AU[29..15] 15 bslbf
			// 	marker_bit 1 bslbf
			// 	DTS_next_AU[14..0] 15 bslbf
			// 	marker_bit 1 bslbf
			// }
			// if (af_descriptor_not_present_flag = = '0') {
			// 	for (i = 0; i  N1; i++) {
			// 		af_descriptor()
			// 	}
			// }
			// else {
			// 	for (i = 0; i < N2; i++) {
			// 		reserved 8 bslbf
			// 	}
			// }
		}

		// TODO: nokori
		readedAFBytes := fieldIndex - 6 + 2
		if int(af.Length) > 0 && readedAFBytes < int(af.Length) {
			af.Stuffing = p.Data[6+readedAFBytes : 5+int(af.Length)]
			for i, v := range af.Stuffing {
				if v != 0xff {
					return fmt.Errorf("[BUG] stuffing bytes contains non-0xff byte. data:0x%02x index:%d", v, i)
				}
			}
		}

		p.AdaptationField = af
	}

	p.isHeaderParsed = true

	return nil
}
