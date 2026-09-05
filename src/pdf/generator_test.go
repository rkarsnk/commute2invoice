package pdf

import (
	"context"
	"testing"

	"github.com/rkarsnk/commute2invoice/src/models"
)

func TestImageDataURI(t *testing.T) {
	uri, err := imageDataURI(models.EvidenceImageInfo{
		MimeType:    "image/png",
		ImageBase64: "aW1hZ2U=",
	})
	if err != nil {
		t.Fatalf("create data URI: %v", err)
	}
	if string(uri) != "data:image/png;base64,aW1hZ2U=" {
		t.Fatalf("unexpected data URI: %q", uri)
	}

	if _, err := imageDataURI(models.EvidenceImageInfo{MimeType: "text/html"}); err == nil {
		t.Fatal("expected unsupported MIME type to be rejected")
	}
}

func TestGenerateRejectsNilPreviewBeforeInvokingRenderer(t *testing.T) {
	if _, err := Generate(context.Background(), nil); err == nil {
		t.Fatal("expected nil preview to be rejected")
	}
}
