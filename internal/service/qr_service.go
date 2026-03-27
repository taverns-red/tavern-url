package service

import (
	"bytes"
	"fmt"
	"image/color"
	"strconv"

	qrcode "github.com/skip2/go-qrcode"
)

// QRService generates QR codes for short links.
type QRService struct{}

// NewQRService creates a new QRService.
func NewQRService() *QRService {
	return &QRService{}
}

// GeneratePNG generates a QR code as a PNG image.
func (s *QRService) GeneratePNG(url string, size int, fg, bg string) ([]byte, error) {
	if size < 100 {
		size = 256
	}
	if size > 2048 {
		size = 2048
	}

	q, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	if fg != "" {
		fgColor, err := parseHexColor(fg)
		if err == nil {
			q.ForegroundColor = fgColor
		}
	}
	if bg != "" {
		bgColor, err := parseHexColor(bg)
		if err == nil {
			q.BackgroundColor = bgColor
		}
	}

	var buf bytes.Buffer
	err = q.Write(size, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to write QR PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// parseHexColor parses a hex color string like "ff0000" into a color.RGBA.
func parseHexColor(hex string) (color.RGBA, error) {
	if len(hex) == 7 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex color: %q", hex)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}
