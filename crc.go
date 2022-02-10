package mpeg2ts

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
