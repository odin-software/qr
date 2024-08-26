package main

import (
	"fmt"
	"log"
)

type EncodingMode uint8
type CodeVersion int8

const (
	NUMERIC_MODE EncodingMode = 1 << iota
	ALPHA_MODE
	BYTE_MODE
)

const (
	CORRECTION_L = iota
	CORRECTION_M
	CORRECTION_Q
	CORRECTION_H
)

var VERSIONS = []int{19, 34, 55, 80, 108, 136, 156, 194, 232, 274}

func GetEncodingMode(text string) EncodingMode {
	var r = 0b0111
	for _, t := range text {
		if t > 57 || t < 48 {
			r = r &^ int(NUMERIC_MODE)
		}
		if t >= 'a' && t <= 'z' {
			r = r &^ int(ALPHA_MODE)
		}
	}
	return EncodingMode(r & -r)
}

func GetCodeVersionAndLength(text string) (int, int) {
	size := len([]byte(text))
	for i, v := range VERSIONS {
		if v > size {
			return i + 1, v
		}
	}
	lastIdx := len(VERSIONS) - 1
	return lastIdx + 1, VERSIONS[lastIdx]
}

func CreateDataSegment(text string) {
	final := ""
	mode := GetEncodingMode(text)
	dataBytes := []byte(text) // an array of the text as bytes
	count := len(dataBytes)   // how many bytes in the data
	dataLength := 0
	version, versionLength := GetCodeVersionAndLength(text)
	final += fmt.Sprintf("%04b", mode)
	final += fmt.Sprintf("%08b", len(text))
	for _, c := range dataBytes {
		final += fmt.Sprintf("%08b", c)
		dataLength += len(fmt.Sprintf("%08b", c))
	}
	// terminator

	log.Println("Stats for %s:", text)
	log.Printf("mode: %04b", mode)
	log.Printf("count: %d", count)
	log.Printf("data: %d", dataLength)
	log.Printf("version picked: %d, length of version in bytes: %d", version, versionLength)
	log.Printf("string at this point: %s", final)
}
