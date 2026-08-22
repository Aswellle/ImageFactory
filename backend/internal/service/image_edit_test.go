package service

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// validPNG is a 1x1 red PNG, base64-encoded — a minimal valid source image.
const validPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

// newEditService builds an ImageEditService whose DB/batch/store/queue are nil;
// ParseImageEditRequest and validation never touch them.
func newEditService() *ImageEditService {
	return &ImageEditService{}
}

func newJSONRequest(t *testing.T, body string) *gin.Context {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/images/edits", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	return c
}

func TestParseJSONEditRequest_Valid(t *testing.T) {
	svc := newEditService()
	img := base64.StdEncoding.EncodeToString([]byte("not-used"))
	_ = img
	body := `{"prompt":"add a hat","images":["` + validPNG + `"],"model":"gpt-image-1","n":1,"size":"1024x1024","input_fidelity":"high"}`
	c := newJSONRequest(t, body)

	req, err := svc.ParseImageEditRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Prompt != "add a hat" {
		t.Errorf("prompt = %q, want %q", req.Prompt, "add a hat")
	}
	if req.Model != "gpt-image-1" {
		t.Errorf("model = %q, want %q", req.Model, "gpt-image-1")
	}
	if req.N != 1 {
		t.Errorf("n = %d, want 1", req.N)
	}
	if req.Size != "1024x1024" {
		t.Errorf("size = %q, want %q", req.Size, "1024x1024")
	}
	if req.InputFidelity != "high" {
		t.Errorf("input_fidelity = %q, want %q", req.InputFidelity, "high")
	}
	if len(req.SourceImages) != 1 {
		t.Fatalf("len(SourceImages) = %d, want 1", len(req.SourceImages))
	}
	if req.SourceImages[0].ContentType != "image/png" {
		t.Errorf("content type = %q, want image/png", req.SourceImages[0].ContentType)
	}
}

func TestParseJSONEditRequest_DataURL(t *testing.T) {
	svc := newEditService()
	body := `{"prompt":"edit","images":["data:image/png;base64,` + validPNG + `"]}`
	c := newJSONRequest(t, body)

	req, err := svc.ParseImageEditRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(req.SourceImages) != 1 || req.SourceImages[0].ContentType != "image/png" {
		t.Errorf("expected one png source image, got %#v", req.SourceImages)
	}
}

func TestParseJSONEditRequest_MultipleImages(t *testing.T) {
	svc := newEditService()
	body := `{"prompt":"merge","images":["` + validPNG + `","` + validPNG + `","` + validPNG + `"]}`
	c := newJSONRequest(t, body)

	req, err := svc.ParseImageEditRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(req.SourceImages) != 3 {
		t.Errorf("len(SourceImages) = %d, want 3", len(req.SourceImages))
	}
}

func TestValidateEditRequest_NoPrompt(t *testing.T) {
	svc := newEditService()
	body := `{"images":["` + validPNG + `"]}`
	c := newJSONRequest(t, body)

	if _, err := svc.ParseImageEditRequest(c); err == nil {
		t.Fatal("expected error for empty prompt, got nil")
	}
}

func TestValidateEditRequest_NoImages(t *testing.T) {
	svc := newEditService()
	body := `{"prompt":"add a hat"}`
	c := newJSONRequest(t, body)

	if _, err := svc.ParseImageEditRequest(c); err == nil {
		t.Fatal("expected error for no images, got nil")
	}
}

func TestValidateEditRequest_TooManyImages(t *testing.T) {
	svc := newEditService()
	body := `{"prompt":"merge","images":["` + validPNG + `","` + validPNG + `","` + validPNG + `","` + validPNG + `"]}`
	c := newJSONRequest(t, body)

	if _, err := svc.ParseImageEditRequest(c); err == nil {
		t.Fatal("expected error for >3 images, got nil")
	}
}

func TestValidateEditRequest_InvalidBase64(t *testing.T) {
	svc := newEditService()
	body := `{"prompt":"edit","images":["not-valid-base64!!!"]}`
	c := newJSONRequest(t, body)

	if _, err := svc.ParseImageEditRequest(c); err == nil {
		t.Fatal("expected error for invalid base64, got nil")
	}
}

func TestValidateEditRequest_UnsupportedType(t *testing.T) {
	svc := newEditService()
	// Valid base64 but decodes to text (a GIF header is required for image/gif;
	// plain text is detected as "text/plain; charset=utf-8").
	txt := base64.StdEncoding.EncodeToString([]byte("this is not an image"))
	body := `{"prompt":"edit","images":["` + txt + `"]}`
	c := newJSONRequest(t, body)

	if _, err := svc.ParseImageEditRequest(c); err == nil {
		t.Fatal("expected error for unsupported image type, got nil")
	}
}

func TestValidateEditRequest_BadDataURL(t *testing.T) {
	svc := newEditService()
	body := `{"prompt":"edit","images":["data:text/plain,foo"]}`
	c := newJSONRequest(t, body)

	if _, err := svc.ParseImageEditRequest(c); err == nil {
		t.Fatal("expected error for non-base64 data URL, got nil")
	}
}

func TestValidateEditRequest_Defaults(t *testing.T) {
	svc := newEditService()
	body := `{"prompt":"edit","images":["` + validPNG + `"]}`
	c := newJSONRequest(t, body)

	req, err := svc.ParseImageEditRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.N != 1 {
		t.Errorf("default n = %d, want 1", req.N)
	}
	if req.ResponseFormat != "url" {
		t.Errorf("default response_format = %q, want url", req.ResponseFormat)
	}
}

func TestParseMultipartEditRequest(t *testing.T) {
	svc := newEditService()
	pngBytes, _ := base64.StdEncoding.DecodeString(validPNG)

	// Build a multipart body manually.
	var body strings.Builder
	body.WriteString("--BOUNDARY\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"prompt\"\r\n\r\n")
	body.WriteString("add a hat\r\n")
	body.WriteString("--BOUNDARY\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"image\"; filename=\"src.png\"\r\n")
	body.WriteString("Content-Type: image/png\r\n\r\n")
	body.Write(pngBytes)
	body.WriteString("\r\n")
	body.WriteString("--BOUNDARY\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"model\"\r\n\r\n")
	body.WriteString("gpt-image-1\r\n")
	body.WriteString("--BOUNDARY--\r\n")

	req := httptest.NewRequest(http.MethodPost, "/v1/images/edits", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=BOUNDARY")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	parsed, err := svc.ParseImageEditRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Prompt != "add a hat" {
		t.Errorf("prompt = %q, want %q", parsed.Prompt, "add a hat")
	}
	if parsed.Model != "gpt-image-1" {
		t.Errorf("model = %q, want %q", parsed.Model, "gpt-image-1")
	}
	if len(parsed.SourceImages) != 1 {
		t.Fatalf("len(SourceImages) = %d, want 1", len(parsed.SourceImages))
	}
	if parsed.SourceImages[0].ContentType != "image/png" {
		t.Errorf("content type = %q, want image/png", parsed.SourceImages[0].ContentType)
	}
	if !parsed.Multipart {
		t.Errorf("multipart = false, want true")
	}
}
