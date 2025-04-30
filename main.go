package main

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// UUIDToBinReordered преобразует UUID string в BINARY(16) с перестановкой байт (MySQL UUID_TO_BIN(uuid, 1))
func UUIDToBinReordered(uuid string) ([]byte, error) {
	uuid = strings.ReplaceAll(uuid, "-", "")
	if len(uuid) != 32 {
		return nil, fmt.Errorf("invalid UUID length")
	}
	raw, err := hex.DecodeString(uuid)
	if err != nil {
		return nil, err
	}
	if len(raw) != 16 {
		return nil, fmt.Errorf("decoded UUID must be 16 bytes")
	}

	reordered := make([]byte, 16)
	copy(reordered[0:2], raw[6:8])   // time_hi
	copy(reordered[2:4], raw[4:6])   // time_mid
	copy(reordered[4:8], raw[0:4])   // time_low
	copy(reordered[8:16], raw[8:16]) // rest

	return reordered, nil
}

func main() {

	uuidStr := "9704b689-4eb9-424a-b076-d41b6fa41f8b"
	uuidBin, err := UUIDToBinReordered(uuidStr)
	if err != nil {
		panic(err) // или log.Fatal(err)
	}

	fmt.Printf("%v\n", uuidBin)

}
