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
	if p.Data == nil || len(p.Data) != 188 {
		return nil, fmt.Errorf("invalid header")
	}
	return p.Data[:4], nil
}

func (p *Packet) GetPayload() []byte {
	if len(p.Data) != PacketSizeDefault {
		return nil
	}
	return p.Data[4:]
}

func (p *Packet) HasAdaptationField() bool {
	c := p.AdaptationFieldControl
	if c == AdaptationField_AdaptationFieldOnly || c == AdaptationField_AdaptationFieldFollowed {
		return true
	}
	return false
}

func (p *Packet) parseHeader() error {
	data, err := p.GetHeader()
	if err != nil {
		return fmt.Errorf("not loaded")

	}
	payload := p.GetPayload()
	if data == nil {
		return fmt.Errorf("not loaded")
	}

	p.SyncByte = data[0]
	p.TransportErrorIndicator = ((data[1] >> 7) & 0x01) == 1
	p.PayloadUnitStartIndicator = ((data[1] >> 6) & 0x01) == 1
	p.TransportPriorityIndicator = ((data[1] >> 5) & 0x01) == 1
	p.PID = (uint16(data[1])&0x1F)<<8 | uint16(data[2])
	p.TransportScrambleControl = (data[3] >> 6) & 0x03
	p.AdaptationFieldControl = (data[3] >> 4) & 0x03
	p.ContinuityCheckIndex = (data[3] & 0x0F)

	// adaptation field
	if p.HasAdaptationField() {
		af := AdaptationField{}
		af.Size = payload[0] + 1 // サイズ書いてある分
		af.DiscontinuityIndicator = ((payload[1] >> 7) & 0x01) == 1
		af.RandomAccessIndicator = ((payload[1] >> 6) & 0x01) == 1
		af.ESPriorityIndicator = ((payload[1] >> 5) & 0x01) == 1
		af.PCRFlag = ((payload[1] >> 4) & 0x01) == 1
		af.OPCRFlag = ((payload[1] >> 3) & 0x01) == 1
		af.SplicingPointFlag = ((payload[1] >> 2) & 0x01) == 1
		af.TransportPrivateDataFlag = ((payload[1] >> 1) & 0x01) == 1
		af.ExtensionFlag = (payload[1] & 0x01) == 1

		p.AdaptationField = af
	}

	if data[0] != 0x47 {
		return fmt.Errorf("invalid magic number")
	}

	return nil
}
