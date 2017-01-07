package mpeg2ts

import "errors"

const (
	ScrambleControl_NotScrambled = 0
	ScrambleControl_Userdefined1 = 1
	ScrambleControl_Userdefined2 = 2
	ScrambleControl_Userdefined3 = 3

	AdaptationField_Reserved                = 0
	AdaptationField_Payloadonly             = 1
	AdaptationField_AdaptationFieldOnly     = 2
	AdaptationField_AdaptationFieldFollowed = 3
)

func (p *Mpeg2TSPacket) Load(data []byte) error {
	if len(data) != 188 {
		return errors.New("size mismatch")
	}

	p.Data = data
	return nil
}

func (p *Mpeg2TSPacket) GetHeader() []byte {
	if p.Data == nil || len(p.Data) != 188 {
		return nil
	}
	return p.Data[:4]
}

func (p *Mpeg2TSPacket) GetPayload() []byte {
	if len(p.Data) != 188 {
		return nil
	}
	return p.Data[4:]
}

func (p *Mpeg2TSPacket) hasAdaptationField() bool {
	if p.AdaptationFieldControl == AdaptationField_AdaptationFieldOnly || p.AdaptationFieldControl == AdaptationField_AdaptationFieldFollowed {
		return true
	}
	return false
}

func (p *Mpeg2TSPacket) ParseHeader() error {
	data := p.GetHeader()
	payload := p.GetPayload()
	if data == nil {
		return errors.New("not loaded")
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
	if p.hasAdaptationField() {
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

		p.AdaptationField = &af
	}

	if data[0] != 0x47 {
		return errors.New("magic number is broken")
	}

	return nil
}
