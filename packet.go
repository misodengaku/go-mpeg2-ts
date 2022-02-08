package mpeg2ts

import (
	"fmt"
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

func (ps *Packets) AddPacket(packetBytes []byte, packetSize int) error {
	if len(packetBytes) != packetSize {
		return fmt.Errorf("packetBytes length and packetSize is not match. len(packetBytes) is %d", len(packetBytes))
	}
	index := len(*ps)
	p := Packet{}
	p.Data = make([]byte, PacketSizeDefault)
	copy(p.Data, packetBytes)
	p.Index = index
	// fmt.Printf("index: %d\n", index)
	err := p.parseHeader()
	if err != nil {
		return err
	}
	*ps = append(*ps, p)
	if index == 0 {

		fmt.Printf("%#v\n", p)
		fmt.Printf("%#v\n", *ps)
		fmt.Println("---------------------------------")
	}
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

	return p.Data[4+p.AdaptationField.Size:], nil
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
		return fmt.Errorf("invalid magic number")
	}

	p.SyncByte = p.Data[0]
	p.TransportErrorIndicator = ((p.Data[1] >> 7) & 0x01) == 1
	p.PayloadUnitStartIndicator = ((p.Data[1] >> 6) & 0x01) == 1
	p.TransportPriorityIndicator = ((p.Data[1] >> 5) & 0x01) == 1
	p.PID = (uint16(p.Data[1])&0x1F)<<8 | uint16(p.Data[2])
	p.TransportScrambleControl = (p.Data[3] >> 6) & 0x03
	p.AdaptationFieldControl = (p.Data[3] >> 4) & 0x03
	p.ContinuityCheckIndex = (p.Data[3] & 0x0F)

	// adaptation field
	if p.HasAdaptationField() {
		af := AdaptationField{}
		af.Size = p.Data[4] + 1 // サイズ書いてある分
		af.DiscontinuityIndicator = ((p.Data[4] >> 7) & 0x01) == 1
		af.RandomAccessIndicator = ((p.Data[4] >> 6) & 0x01) == 1
		af.ESPriorityIndicator = ((p.Data[4] >> 5) & 0x01) == 1
		af.PCRFlag = ((p.Data[4] >> 4) & 0x01) == 1
		af.OPCRFlag = ((p.Data[4] >> 3) & 0x01) == 1
		af.SplicingPointFlag = ((p.Data[4] >> 2) & 0x01) == 1
		af.TransportPrivateDataFlag = ((p.Data[4] >> 1) & 0x01) == 1
		af.ExtensionFlag = (p.Data[4] & 0x01) == 1

		// if PCR_FLAG == '1'

		p.AdaptationField = af
	}

	p.isHeaderParsed = true

	return nil
}
