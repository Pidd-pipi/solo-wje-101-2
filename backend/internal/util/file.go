package util

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var allowedExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// SaveUpload persists an uploaded file under uploadDir and returns its public path.
func SaveUpload(uploadDir string, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		return "", fmt.Errorf("unsupported file extension %q", ext)
	}
	if file.Size > 5<<20 {
		return "", fmt.Errorf("file too large: %d bytes", file.Size)
	}
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), randomName(8), ext)
	dst := filepath.Join(uploadDir, name)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomName(n int) string {
	b := make([]byte, n)
	seed := time.Now().UnixNano()
	for i := range b {
		b[i] = letters[(seed>>(uint(i)*3))%int64(len(letters))]
	}
	return string(b)
}
