package mpeg2ts

import "sync"

const PacketSizeDefault = 188
const PacketSizeWithFEC = 204

type MPEG2TS struct {
	PacketList
	IsUseFEC bool
}

type PacketList struct {
	packets []Packet
	mutex   *sync.Mutex
}

// MPEG2-TS Packet
type Packet struct {
	Index                      int
	Data                       []byte
	SyncByte                   byte
	PID                        uint16 // 中身は13bit
	TransportScrambleControl   byte
	AdaptationFieldControl     byte
	TransportErrorIndicator    bool
	PayloadUnitStartIndicator  bool
	TransportPriorityIndicator bool
	ContinuityCheckIndex       byte
	AdaptationField            AdaptationField

	isHeaderParsed bool
}

type AdaptationField struct {
	Length                        byte
	DiscontinuityIndicator        bool
	RandomAccessIndicator         bool
	ESPriorityIndicator           bool
	PCRFlag                       bool
	OPCRFlag                      bool
	SplicingPointFlag             bool
	TransportPrivateDataFlag      bool
	ExtensionFlag                 bool
	ProgramClockReference         ProgramClockReference
	OriginalProgramClockReference ProgramClockReference
	Stuffing                      []byte
}

type ProgramClockReference struct {
	Base      uint64
	Extension uint16
}

type StreamCheckResult struct {
	DropCount int
	DropList  []struct {
		Description string
		Index       int
	}
}
