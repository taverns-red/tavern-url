package service

import (
	"testing"
)

func TestGeneratePNG(t *testing.T) {
	svc := NewQRService()

	data, err := svc.GeneratePNG("https://example.com/test", 256, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty PNG data")
	}

	// PNG magic bytes.
	if len(data) < 4 || data[0] != 0x89 || data[1] != 0x50 || data[2] != 0x4E || data[3] != 0x47 {
		t.Error("data does not have PNG magic bytes")
	}
}

func TestGeneratePNG_CustomColors(t *testing.T) {
	svc := NewQRService()

	data, err := svc.GeneratePNG("https://example.com", 256, "000000", "ffffff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PNG data")
	}
}

func TestGeneratePNG_SizeLimits(t *testing.T) {
	svc := NewQRService()

	// Too small — should default to 256.
	data, err := svc.GeneratePNG("https://example.com", 10, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PNG data")
	}

	// Too large — should cap at 2048.
	data, err = svc.GeneratePNG("https://example.com", 5000, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PNG data")
	}
}

func TestParseHexColor(t *testing.T) {
	cases := []struct {
		hex     string
		r, g, b uint8
	}{
		{"ff0000", 255, 0, 0},
		{"00ff00", 0, 255, 0},
		{"0000ff", 0, 0, 255},
		{"#6366f1", 99, 102, 241},
	}

	for _, tc := range cases {
		c, err := parseHexColor(tc.hex)
		if err != nil {
			t.Errorf("parseHexColor(%q) error: %v", tc.hex, err)
			continue
		}
		if c.R != tc.r || c.G != tc.g || c.B != tc.b {
			t.Errorf("parseHexColor(%q) = RGB(%d,%d,%d), want RGB(%d,%d,%d)",
				tc.hex, c.R, c.G, c.B, tc.r, tc.g, tc.b)
		}
	}
}
