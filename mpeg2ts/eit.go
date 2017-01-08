package mpeg2ts

type EITTable struct {
	TableID                  byte
	SectionSyntaxIndicator   bool
	SectionLength            uint16
	ServiceID                uint16
	Version                  byte
	CurrentNextIndicator     bool
	SectionNumber            byte
	LastSectionNumber        byte
	SegmentLastSectionNumber byte
	TransportStreamID        byte
	OriginalNetworkID        byte
	LastTableID              byte
	CRC32                    uint
	Events                   []EITEvent
}

type EITEvent struct {
	EventID           int
	StartTime         int
	Duration          int
	RunningStatus     int
	FreeCAMode        int
	DescriptorsLength int
	Descriptors       []EITDescriptor
}

type EITDescriptor struct {
}
