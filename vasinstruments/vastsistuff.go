package vasinstruments

import (
	"fmt"
	"time"
)

func bytestoint32(b []byte) int32 {
	if len(b) >= 4 {
		return int32(uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24)
	}
	if len(b) >= 2 {
		return int32(uint32(b[1]) | uint32(b[0])<<8)
	}
	return 0
}

func TSIdate(b []byte) string {
	if len(b) < 6 {
		return "0000-00-00"
	}
	y := int(b[0])*256 + int(b[1])
	m := int(b[3])
	d := int(b[5])
	return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
}

func TSItime(b []byte) string {
	if len(b) < 6 {
		return "00:00:00"
	}
	t := int(b[1])
	m := int(b[3])
	s := int(b[5])
	return fmt.Sprintf("%02d:%02d:%02d", t, m, s)
}

func TSIbyteorder(b []byte) []byte {
	// Skapa en kopia så att vi inte förstör original-slicen i minnet
	res := make([]byte, len(b))
	copy(res, b)

	// Byt plats på parvis bytes säkert utan risk för index out of range
	for i := 0; i+1 < len(res); i += 2 {
		res[i], res[i+1] = res[i+1], res[i]
	}
	return res
}

func setTSItime() []uint16 {
	t := time.Now()
	return []uint16{
		uint16(t.Year()),
		uint16(t.Month()),
		uint16(t.Day()),
		uint16(t.Hour()),
		uint16(t.Minute()),
		uint16(t.Second()),
	}
}
