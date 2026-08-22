// Package image provides pure image-manipulation utilities built on the Go
// standard library. It has no external dependencies so it builds anywhere the
// rest of ImageForge builds.
//
// Supported decode/encode formats are the ones registered by the stdlib:
// JPEG, PNG, and GIF. Other formats (WebP, BMP, …) cannot be resized here;
// callers fall back to serving the original bytes as the thumbnail.
//
// Scaling is implemented with a manual nearest-neighbor loop because the
// stdlib "image/draw" package does not provide a scaler (those live in the
// optional golang.org/x/image/draw module, which ImageForge does not import).
package image

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
)

// MimeType detects the content type of image data from its magic bytes. It
// recognizes the common web image formats explicitly and falls back to
// http.DetectContentType for anything else.
func MimeType(data []byte) string {
	if len(data) >= 12 {
		switch {
		case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47:
			return "image/png"
		case data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
			return "image/jpeg"
		case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
			return "image/gif"
		case data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
			data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50:
			return "image/webp"
		case data[0] == 0x42 && data[1] == 0x4D:
			return "image/bmp"
		}
	}
	return http.DetectContentType(data)
}

// Resize decodes image data, scales it to fit within width×height while
// preserving aspect ratio (never upscaling), and re-encodes it in the same
// format as the input. The returned bytes are suitable for storage.
//
// If the data cannot be decoded (unsupported format or corruption) the
// original bytes are returned unchanged together with a non-nil error so the
// caller can fall back to serving the original as the thumbnail.
func Resize(data []byte, width, height int) ([]byte, error) {
	if width <= 0 || height <= 0 {
		return data, nil
	}

	src, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, err
	}

	bounds := src.Bounds()
	origW, origH := bounds.Dx(), bounds.Dy()
	if origW <= 0 || origH <= 0 {
		return data, nil
	}

	// Already fits: return as-is to avoid a pointless re-encode.
	if origW <= width && origH <= height {
		return data, nil
	}

	scale := min(float64(width)/float64(origW), float64(height)/float64(origH))
	newW := max(1, int(float64(origW)*scale))
	newH := max(1, int(float64(origH)*scale))

	// Nearest-neighbor scale into a new RGBA image. Using src.At makes this
	// work for whatever concrete type the decoder produced (RGBA, Gray,
	// Paletted, …) without per-format cases.
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := range newH {
		sy := bounds.Min.Y + y*origH/newH
		for x := range newW {
			sx := bounds.Min.X + x*origW/newW
			dst.Set(x, y, src.At(sx, sy))
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85})
	case "gif":
		err = gif.Encode(&buf, dst, nil)
	case "png":
		err = png.Encode(&buf, dst)
	default:
		// Unknown decoder format: re-encode as PNG so the output is always valid.
		err = png.Encode(&buf, dst)
	}
	if err != nil {
		return data, err
	}
	return buf.Bytes(), nil
}
