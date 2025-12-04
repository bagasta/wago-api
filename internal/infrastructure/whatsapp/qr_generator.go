package whatsapp

import (
	"encoding/base64"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCode(code string) (string, error) {
	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	// Return raw base64 (no data URL prefix) for flexibility.
	return base64.StdEncoding.EncodeToString(png), nil
}
