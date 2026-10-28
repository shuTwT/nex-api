package utils

import "testing"

func TestPositiveInt(t *testing.T) {
	if value, err := PositiveInt("", 10); err != nil || value != 10 {
		t.Fatalf("fallback = %d, %v", value, err)
	}
	if value, err := PositiveInt("2", 10); err != nil || value != 2 {
		t.Fatalf("parsed = %d, %v", value, err)
	}
	if _, err := PositiveInt("0", 10); err == nil {
		t.Fatal("PositiveInt accepted zero")
	}
}
