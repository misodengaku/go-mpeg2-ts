package mpeg2ts

import (
	"fmt"
)

// Program Association Table
type PAT struct {
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

func (p *Packet) ParsePAT() (PAT, error) {
	pat := PAT{}
	payload, err := p.GetPayload()
	if err != nil {
		return PAT{}, err
	}
	pat.Pointer = payload[0]
	pat.TableID = payload[1]
	pat.SectionSyntaxIndicator = ((payload[2] >> 7) & 0x01) == 1
	if ((payload[2] >> 6) & 0x01) == 1 {
		return PAT{}, fmt.Errorf("invalid format")
	}
	pat.SectionLength = uint16(payload[2]&0x0F)<<8 | uint16(payload[3])
	pat.TransportStreamID = uint16(payload[4])<<8 | uint16(payload[5])
	pat.Version = (payload[6] >> 1) & 0x1F
	pat.CurrentNextIndicator = (payload[6] & 0x01) == 0x01
	pat.SectionNumber = payload[7]
	pat.LastSectionNumber = payload[8]

	fmt.Printf("PAT: %#v\r\n", pat)

	pat.Programs = make([]PATProgram, (pat.SectionLength-5-4)/4)
	for i := uint16(0); i < (pat.SectionLength-5-4)/4; i++ {
		base := 9 + i*4
		// fmt.Println("base", base, len(payload))
		pat.Programs[i].ProgramNumber = uint16(payload[base])<<8 | uint16(payload[base+1])
		if pat.Programs[i].ProgramNumber == 0x0000 {
			pat.Programs[i].NetworkPID = uint16(payload[base+2]&0x1f)<<8 | uint16(payload[base+3])&0x1fff
		} else {
			pat.Programs[i].ProgramMapPID = uint16(payload[base+2]&0x1f)<<8 | uint16(payload[base+3])&0x1fff
		}
		// pat.Programs[i].ProgramMapPID = uint16(payload[base+2]&0x1f)<<8 | uint16(payload[base+3])
	}
	// fmt.Printf("CRC32 dump: %02x %02x %02x %02x\r\n", uint(payload[pat.SectionLength]), uint(payload[pat.SectionLength+1]), uint(payload[pat.SectionLength+2]), uint(payload[pat.SectionLength+3]))
	pat.CRC32 = uint(payload[pat.SectionLength])<<24 | uint(payload[pat.SectionLength+1])<<16 | uint(payload[pat.SectionLength+2])<<8 | uint(payload[pat.SectionLength+3])
	// fmt.Printf("%#v\r\n", payload[1:pat.SectionLength])

	crc := CalculateCRC(0, payload[1:pat.SectionLength]) ^ 0xffffffff
	if uint32(pat.CRC32) != crc {
		return PAT{}, fmt.Errorf("CRC32 mismatch")
	}

	// fmt.Println("CRC OK")
	return pat, nil
}

// based on isal's crc32 algo found at:
// https://github.com/01org/isa-l/blob/master/crc/crc_base.c#L138-L155
func CalculateCRC(seed uint32, data []byte) (crc uint32) {
	rem := uint64(^seed)

	var i, j int

	const (
		// defined in
		// https://github.com/01org/isa-l/blob/master/crc/crc_base.c#L33
		MAX_ITER = 8
	)

	for i = 0; i < len(data); i++ {
		rem = rem ^ (uint64(data[i]) << 24)
		for j = 0; j < MAX_ITER; j++ {
			rem = rem << 1
			if (rem & 0x100000000) != 0 {
				rem ^= uint64(0x04C11DB7)
			}
		}
	}

	crc = uint32(^rem)
	return
}
