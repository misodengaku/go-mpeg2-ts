package mpeg2ts

import (
	"errors"
	"fmt"
	"time"
)

// ETSI EN 300 468 V1.17.1 p.34
type TDT struct {
	TableID                byte   // 8
	SectionSyntaxIndicator byte   // 1
	ReservedFutureUse      byte   // 1
	Reserved1              byte   // 2
	SectionLength          uint16 // 12
	RAWTimestamp           uint64 // 40

	Timestamp time.Time
}

// ETSI EN 300 468 V1.17.1 p.35
type TOT struct {
	TDT
	Reserved2         byte   // 4
	DescriptorsLength uint16 // 12
	Descriptors       []struct{}
	CRC32             uint //32

	Timestamp time.Time
}

var ErrPacketIsTDT = errors.New("This packet is TDT. Use ParseTDT")

func (p *Packet) ParseTDT(acceptTOT bool) (TDT, error) {

	tdt := TDT{}
	payload, err := p.GetPayload()
	if err != nil {
		return TDT{}, err
	}
	tdt.TableID = payload[1]
	if tdt.TableID != TableID_TimeDateSection && tdt.TableID != TableID_TimeOffsetSection {
		return TDT{}, fmt.Errorf("invalid TableID. expected: 0x70, actual: 0x%02x", tdt.TableID)
	}
	if tdt.TableID == TableID_TimeOffsetSection && !acceptTOT {
		return TDT{}, errors.New("This packet is TOT. Set the acceptTOT to true or use ParseTOT")
	}
	tdt.SectionSyntaxIndicator = (payload[2] >> 7) & 0x01
	tdt.ReservedFutureUse = (payload[2] >> 6) & 0x01
	tdt.Reserved1 = (payload[2] >> 4) & 0x03
	tdt.SectionLength = uint16(payload[2]&0x0F)<<8 | uint16(payload[3])
	tdt.RAWTimestamp = uint64(payload[4])<<32 | uint64(payload[5])<<24 | uint64(payload[6])<<16 | uint64(payload[7])<<8 | uint64(payload[8])

	tdt.Timestamp = getTimestampByMJD(tdt.RAWTimestamp)
	return tdt, nil
}

func (p *Packet) ParseTOT() (TOT, error) {

	tot := TOT{}
	payload, err := p.GetPayload()
	if err != nil {
		return TOT{}, err
	}
	tot.TableID = payload[1]
	if tot.TableID != TableID_TimeOffsetSection {
		return TOT{}, ErrPacketIsTDT
	}
	tot.TDT, err = p.ParseTDT(true)
	if err != nil {
		return TOT{}, err
	}
	tot.Reserved2 = (payload[9] >> 4) & 0x0f
	tot.DescriptorsLength = uint16(payload[9]&0x0f)<<8 | uint16(payload[10])
	for i := 0; i < int(tot.DescriptorsLength); i++ {
		// FIXME: implement descriptor
	}
	tot.CRC32 = uint(payload[tot.SectionLength])<<24 | uint(payload[tot.SectionLength+1])<<16 | uint(payload[tot.SectionLength+2])<<8 | uint(payload[tot.SectionLength+3])

	tot.Timestamp = getTimestampByMJD(tot.RAWTimestamp)
	crc := calculateCRC(payload[1:tot.SectionLength])
	if uint32(tot.CRC32) != crc {
		return TOT{}, errors.New("CRC32 mismatch")
	}
	return tot, nil
}

func getTimestampByMJD(mjd uint64) time.Time {
	rawDate := mjd >> 24
	mjdOrigin := time.Date(1858, 11, 17, 0, 00, 00, 00, time.UTC)
	mjdDate := mjdOrigin.Add(time.Duration(rawDate) * time.Hour * 24)
	hour := bcdToDec(byte((mjd >> 16) & 0xff))
	min := bcdToDec(byte((mjd >> 8) & 0xff))
	sec := bcdToDec(byte((mjd) & 0xff))
	return mjdDate.Add(time.Duration(hour) * time.Hour).Add(time.Duration(min) * time.Minute).Add(time.Duration(sec) * time.Second)
}

func bcdToDec(bcd byte) byte {
	return bcd>>4*10 + bcd&0x0f
}
