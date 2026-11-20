package fileresource

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreRejectsUnknownType(t *testing.T) {
	header := fileHeader(t, "notes.exe", []byte("x"))
	if _, err := Store(header, t.TempDir()); err != ErrFileType {
		t.Fatalf("got %v", err)
	}
}

func TestStoreWritesImage(t *testing.T) {
	dir := t.TempDir()
	header := fileHeader(t, "park.PNG", []byte("image"))
	path, err := Store(header, dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(path) != ".png" {
		t.Fatalf("path %s", path)
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.Base(path))); err != nil {
		t.Fatal(err)
	}
}

func fileHeader(t *testing.T, name string, body []byte) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{`form-data; name="file"; filename="` + name + `"`},
		"Content-Type":        []string{"application/octet-stream"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader := multipart.NewReader(&buf, writer.Boundary())
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = form.RemoveAll() })
	return form.File["file"][0]
}
