package mpeg2ts

const PacketSizeDefault = 188
const PacketSizeWithFEC = 204

type MPEG2TS struct {
	Packets  Packets
	IsUseFEC bool
}

type Packets []Packet

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
	Size                     byte
	DiscontinuityIndicator   bool
	RandomAccessIndicator    bool
	ESPriorityIndicator      bool
	PCRFlag                  bool
	OPCRFlag                 bool
	SplicingPointFlag        bool
	TransportPrivateDataFlag bool
	ExtensionFlag            bool
}

type StreamCheckResult struct {
	DropCount int
	DropList  []struct {
		Description string
		Index       int
	}
}
