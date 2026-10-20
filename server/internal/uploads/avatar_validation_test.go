package uploads

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"testing"
)

func TestValidateAvatarDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   error
	}{
		{name: "square avatar at pixel limit", width: 4_000, height: 4_000},
		{name: "zero width", width: 0, height: 100, want: ErrInvalidImage},
		{name: "dimension limit exceeded", width: maxAvatarDimension + 1, height: 100, want: ErrImageDimensions},
		{name: "pixel limit exceeded", width: 4_001, height: 4_001, want: ErrImageTooManyPixels},
		{name: "landscape aspect ratio exceeded", width: 3_001, height: 1_000, want: ErrImageAspectRatio},
		{name: "portrait aspect ratio exceeded", width: 1_000, height: 3_001, want: ErrImageAspectRatio},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAvatarDimensions(tt.width, tt.height)
			if !errors.Is(err, tt.want) {
				t.Fatalf("validateAvatarDimensions(%d, %d) error = %v, want %v", tt.width, tt.height, err, tt.want)
			}
		})
	}
}

func TestValidateAvatarImage(t *testing.T) {
	var pngData bytes.Buffer
	if err := png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatalf("encode test image: %v", err)
	}

	tests := []struct {
		name        string
		data        []byte
		contentType string
		want        error
	}{
		{name: "valid PNG", data: pngData.Bytes(), contentType: "image/png"},
		{name: "content type does not match image bytes", data: pngData.Bytes(), contentType: "image/jpeg", want: ErrUnsupportedFileType},
		{name: "invalid image bytes", data: []byte("not an image"), contentType: "image/png", want: ErrInvalidImage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAvatarImage(bytes.NewReader(tt.data), tt.contentType)
			if !errors.Is(err, tt.want) {
				t.Fatalf("validateAvatarImage() error = %v, want %v", err, tt.want)
			}
		})
	}
}
