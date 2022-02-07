package mpeg2ts

import (
	"errors"
	"fmt"

	"github.com/snksoft/crc"
)

type PATTable struct {
	Pointer                byte
	TableID                byte
	SectionSyntaxIndicator bool
	Reserved1              byte
	SectionLength          uint16
	TransportStreamID      uint16
	Reserved2              int
	Version                byte
	CurrentNextIndicator   bool
	SectionNumber          byte
	LastSectionNumber      byte
	CRC32                  uint
	Programs               []PATProgram
}

type PATProgram struct {
	ProgramNumber uint16
	Reserved      int
	NetworkPID    uint16
	ProgramMapPID uint16
}

func (p *Packet) ParsePAT() (*PATTable, error) {
	pat := PATTable{}
	payload := p.GetPayload()
	pat.Pointer = payload[0]
	pat.TableID = payload[1]
	pat.SectionSyntaxIndicator = ((payload[2] >> 7) & 0x01) == 1
	if ((payload[2] >> 6) & 0x01) == 1 {
		return nil, errors.New("invalid format")
	}
	pat.SectionLength = uint16(payload[2]&0x0F)<<8 | uint16(payload[3])
	pat.TransportStreamID = uint16(payload[4])<<8 | uint16(payload[5])
	pat.Version = (payload[6] >> 1) & 0x1F
	pat.CurrentNextIndicator = (payload[6] & 0x01) == 0x01
	pat.SectionNumber = payload[7]
	pat.LastSectionNumber = payload[8]

	// fmt.Printf("pat %d dump %x %x %t %d \r\n", p.Index, pat.Pointer, pat.TableID, pat.SectionSyntaxIndicator, pat.SectionLength)
	// for i := 0; i < 188; i++ {
	// 	fmt.Printf("%02x ", p.Data[i])
	// }
	// fmt.Println()

	// SectionLength - 5 - 4
	header, _ := p.GetHeader()
	fmt.Printf("%#v\r\n", header)
	pat.Programs = make([]PATProgram, (pat.SectionLength-5-4)/4)
	for i := uint16(0); i < (pat.SectionLength-5-4)/4; i++ {
		base := 9 + i*4
		pat.Programs[i].ProgramNumber = uint16(payload[base])<<8 | uint16(payload[base+1])
		// if pat.Programs[i].ProgramNumber == 0 {
		// 	pat.Programs[i].NetworkPID = uint16(payload[base+2]&0x1f)<<8 | uint16(payload[base+3])
		// } else {
		// 	pat.Programs[i].ProgramMapPID = uint16(payload[base+2]&0x1f)<<8 | uint16(payload[base+3])
		// }
		pat.Programs[i].ProgramMapPID = uint16(payload[base+2]&0x1f)<<8 | uint16(payload[base+3])
	}
	fmt.Printf("%02x %02x %02x %02x\r\n", uint(payload[pat.SectionLength]), uint(payload[pat.SectionLength+1]), uint(payload[pat.SectionLength+2]), uint(payload[pat.SectionLength+3]))
	pat.CRC32 = uint(payload[pat.SectionLength])<<24 | uint(payload[pat.SectionLength+1])<<16 | uint(payload[pat.SectionLength+2])<<8 | uint(payload[pat.SectionLength+3])
	fmt.Printf("%#v\r\n", payload[1:pat.SectionLength])
	// crc32q := crc32.MakeTable(0x82608EDB)
	poly := crc.CRC32
	poly.Polynomial = 0x04C11DB7
	poly.ReflectIn = false
	poly.ReflectOut = false
	poly.FinalXor = 0x00000000
	checksum := crc.CalculateCRC(crc.CRC32, payload[1:pat.SectionLength])
	if uint64(pat.CRC32) != checksum {
		return &pat, errors.New("checksum mismatch")
	}
	fmt.Println("CRC OK")
	return &pat, nil
}
