package main

type EncodingMode uint8

const (
	NUMERIC_MODE EncodingMode = 1 << iota
	ALPHA_MODE
	BYTE_MODE
)

func GetEncodingMode(text string) EncodingMode {
	var r = 0b111
	for _, t := range text {
		if t > 57 || t < 48 {
			r = r &^ int(NUMERIC_MODE)
		}
		if t >= 'a' && t <= 'z' {
			r = r &^ int(ALPHA_MODE)
		}
	}
	return EncodingMode(r)
}

func CreateDataSegment(text string) {
}
