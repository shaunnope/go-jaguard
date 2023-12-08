package utils

import (
	"encoding/binary"
	"os"
)

func UnmarshalInt(data []byte) int {
	return int(binary.BigEndian.Uint64(data))
}

func MarshalInt(i int) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(i))
	return data
}

func WriteBytes(data []byte, dir string, name string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(dir+name, data, 0644)
}
