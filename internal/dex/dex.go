package dex

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	headerSize     = 0x70
	endianConstant = 0x12345678
)

var le = binary.LittleEndian

func CollectPrefixHits(data []byte, prefixes []string, hits map[string]string) error {
	if len(data) < headerSize {
		return errors.New("file is too small to be a DEX file")
	}
	if !bytes.HasPrefix(data, []byte("dex\n")) {
		return errors.New("missing DEX magic bytes")
	}
	if le.Uint32(data[0x28:]) != endianConstant {
		return errors.New("unsupported byte order")
	}

	stringIDsSize := uint64(le.Uint32(data[0x38:]))
	stringIDsOff := uint64(le.Uint32(data[0x3C:]))
	typeIDsSize := uint64(le.Uint32(data[0x40:]))
	typeIDsOff := uint64(le.Uint32(data[0x44:]))

	n := uint64(len(data))
	if stringIDsOff+stringIDsSize*4 > n {
		return errors.New("string_ids table is out of bounds")
	}
	if typeIDsOff+typeIDsSize*4 > n {
		return errors.New("type_ids table is out of bounds")
	}

	prefixBytes := make([][]byte, len(prefixes))
	pending := 0
	for i, p := range prefixes {
		prefixBytes[i] = []byte(p)
		if _, done := hits[p]; !done {
			pending++
		}
	}

	for i := uint64(0); i < typeIDsSize && pending > 0; i++ {
		descIdx := uint64(le.Uint32(data[typeIDsOff+i*4:]))
		if descIdx >= stringIDsSize {
			continue
		}
		strOff := uint64(le.Uint32(data[stringIDsOff+descIdx*4:]))
		s, ok := stringAt(data, strOff)
		if !ok {
			continue
		}
		for j, p := range prefixes {
			if _, done := hits[p]; done {
				continue
			}
			if bytes.HasPrefix(s, prefixBytes[j]) {
				hits[p] = string(s)
				pending--
			}
		}
	}
	return nil
}


func stringAt(data []byte, off uint64) ([]byte, bool) {
	n := uint64(len(data))
	i := off
	terminated := false
	for k := 0; k < 5 && i < n; k++ {
		b := data[i]
		i++
		if b&0x80 == 0 {
			terminated = true
			break
		}
	}
	if !terminated || i >= n {
		return nil, false
	}
	end := bytes.IndexByte(data[i:], 0)
	if end < 0 {
		return nil, false
	}
	return data[i : i+uint64(end)], true
}
