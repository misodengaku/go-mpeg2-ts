package mpeg2ts

import (
	"hash/crc32"
	"math/bits"
)

// MSB-first CRC-32-IEEE
func calculateCRC(data []byte) uint32 {
	reverseData := []byte{}
	for _, v := range data {
		reverseData = append(reverseData, byte(bits.Reverse8(uint8(v))))
	}
	return bits.Reverse32(crc32.Checksum(reverseData, crc32.MakeTable(0xEDB88320))) ^ 0xffffffff
}
