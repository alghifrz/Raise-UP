package gallery_test

import (
	"bytes"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diuk/raiseup/internal/gallery"
	"github.com/gin-gonic/gin"
)

func TestUploadHandlerAcceptsMultipartImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newMemoryStore()
	service := gallery.NewService(store)
	service.SetObjectStorage(&memoryObjectStorage{}, "gallery")
	handler := gallery.NewHandler(service, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), 1024)

	body, contentType := multipartUploadBody(t, "2")
	request := httptest.NewRequest(http.MethodPost, "/gallery/upload", body)
	request.Header.Set("Content-Type", contentType)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request

	handler.Upload(context)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestUploadHandlerRejectsInvalidSortOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := gallery.NewService(newMemoryStore())
	service.SetObjectStorage(&memoryObjectStorage{}, "gallery")
	handler := gallery.NewHandler(service, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), 1024)

	body, contentType := multipartUploadBody(t, "-1")
	request := httptest.NewRequest(http.MethodPost, "/gallery/upload", body)
	request.Header.Set("Content-Type", contentType)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request

	handler.Upload(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func multipartUploadBody(t *testing.T, sortOrder string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	file, err := writer.CreateFormFile("image", "photo.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField("caption", "Dokumentasi"); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField("sort_order", sortOrder); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}
