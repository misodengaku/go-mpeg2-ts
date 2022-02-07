package mpeg2ts

const (
	// based on ISO/IEC 13818-1
	PID_PAT        = 0x0000
	PID_CAT        = 0x0001
	PID_TSDT       = 0x0002
	PID_Reserved1  = 0x0003
	PID_Reserved2  = 0x0004
	PID_Reserved3  = 0x0005
	PID_Reserved4  = 0x0006
	PID_Reserved5  = 0x0007
	PID_Reserved6  = 0x0008
	PID_Reserved7  = 0x0009
	PID_Reserved8  = 0x000a
	PID_Reserved9  = 0x000b
	PID_Reserved10 = 0x000c
	PID_Reserved11 = 0x000d
	PID_Reserved12 = 0x000e
	PID_Reserved13 = 0x000f
	PID_NullPacket = 0x1fff

	// based on ARIB STD-B10
	PID_PMT                            = PID_PAT
	PID_ECM                            = PID_PMT
	PID_ECM_S                          = PID_PMT
	PID_EMM                            = PID_CAT
	PID_EMM_S                          = PID_CAT
	PID_NIT                            = 0x0010
	PID_SDT                            = 0x0011
	PID_BAT                            = 0x0011
	PID_EIT                            = 0x0012
	PID_EIT_Terrestrial1               = 0x0012
	PID_EIT_Terrestrial2               = 0x0026
	PID_EIT_Terrestrial3               = 0x0027
	PID_RST                            = 0x0013
	PID_TDT                            = 0x0014
	PID_TOT                            = 0x0014
	PID_DCT                            = 0x0017
	PID_DLT                            = PID_DCT
	PID_DIT                            = 0x001e
	PID_SIT                            = 0x001f
	PID_LIT1                           = PID_PMT
	PID_LIT2                           = 0x0020
	PID_ERT1                           = PID_PMT
	PID_ERT2                           = 0x0021
	PID_ITT                            = PID_PMT
	PID_PCAT                           = 0x0022
	PID_SDTT                           = 0x0023
	PID_SDTT_Terrestrial1              = 0x0023
	PID_SDTT_Terrestrial2              = 0x0028
	PID_BIT                            = 0x0024
	PID_NBIT                           = 0x0025
	PID_LDT                            = 0x0025
	PID_CDT                            = 0x0029
	PID_MultipleFrameHeaderInformation = 0x002f
	PID_DSM_CC                         = PID_PMT
	PID_AIT                            = PID_PMT
	// PID_STExclude                      = 0x0000
)
