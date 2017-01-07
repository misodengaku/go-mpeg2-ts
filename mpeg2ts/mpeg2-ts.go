package mpeg2ts

type Mpeg2TS struct {
	Packets Mpeg2TSPackets
}

type Mpeg2TSPackets []Mpeg2TSPacket

type Mpeg2TSPacket struct {
	Data                       []byte
	SyncByte                   byte
	PID                        uint16 // 中身は13bit
	TransportScrambleControl   byte
	AdaptationFieldControl     byte
	TransportErrorIndicator    bool
	PayloadUnitStartIndicator  bool
	TransportPriorityIndicator bool
	ContinuityCheckIndex       byte
	AdaptationField            *AdaptationField
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

func (m Mpeg2TS) New(count int64) Mpeg2TS {
	m.Packets = make(Mpeg2TSPackets, count)
	return m
}
