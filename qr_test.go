package main

import "testing"

func TestGetEncodingMode(t *testing.T) {
	t.Run("it should determine correct numeric encoding mode", func(t *testing.T) {
		got := GetEncodingMode("123884")
		want := NUMERIC_MODE
		if got != want {
			t.Errorf("got %d wanted %d", got, want)
		}
	})
	t.Run("it should not determine numeric encoding mode", func(t *testing.T) {
		got := GetEncodingMode("123E84")
		want := ALPHA_MODE
		if got != want {
			t.Errorf("got %d wanted %d", got, want)
		}
	})
	t.Run("it should determine a correct alpha encoding mode", func(t *testing.T) {
		got := GetEncodingMode("E: PE8D")
		want := ALPHA_MODE
		if got != want {
			t.Errorf("got %d wanted %d", got, want)
		}
	})
	t.Run("it should determine byte encoding mode", func(t *testing.T) {
		got := GetEncodingMode("https://testing.com/222")
		want := BYTE_MODE
		if got != want {
			t.Errorf("got %d wanted %d", got, want)
		}
	})
}

func TestGetCodeVersion(t *testing.T) {
	t.Run("it should return the right qr code version", func(t *testing.T) {
		got := GetCodeVersion("Hello Test 2")
		want := VERSION_4
		if got != want {
			t.Errorf("got %d wanted %d", got, want)
		}
	})
}
