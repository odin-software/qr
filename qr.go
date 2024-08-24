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

const (
	VERSION_1 = 16
	VERSION_2 = 28
	VERSION_3 = 44
	VERSION_4 = 64
	VERSION_5 = 86
	VERSION_6 = 108
	VERSION_7 = 124
	VERSION_8 = 154
	VERSION_9 = 182
)

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

func GetCodeVersion(text string) int {
	return VERSION_4
}

func CreateDataSegment(text string) {
	final := ""
	final1 := ""
	mode := GetEncodingMode(text)
	// log.Printf("%04b", mode)
	final += fmt.Sprintf("%04b", mode)
	final += fmt.Sprintf("%08b", len(text))
	t := []byte(text)
	for _, c := range t {
		final += fmt.Sprintf("%08b", c)
		final1 += fmt.Sprintf("%08b", c)
	}
	// terminatorBits := "0000"
	// buffer.WriteString(fmt.Sprintf("%08b", len(text)))
	// buffer.WriteString(fmt.Sprintf("%08b", []byte(text)))
	// padding with zeroes

	log.Printf("%s", final)
	log.Printf("%s", t)
	log.Printf("%d", len(final1))
	log.Printf("%d", len(final))
}
