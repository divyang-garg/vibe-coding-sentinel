// Package services - Document extractors tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"archive/zip"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestExtractPDFText tests PDF text extraction
func TestExtractPDFText(t *testing.T) {
	// Check if pdftotext is available
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not available, skipping PDF extraction test")
	}

	tests := []struct {
		name      string
		setupFile func(t *testing.T) string
		wantErr   bool
	}{
		{
			name: "file not found",
			setupFile: func(t *testing.T) string {
				return "/nonexistent/file.pdf"
			},
			wantErr: true,
		},
		{
			name: "invalid PDF format",
			setupFile: func(t *testing.T) string {
				// Create a file that's not a valid PDF
				tmpFile, err := os.CreateTemp("", "test-*.txt")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.WriteString("This is not a PDF")
				tmpFile.Close()
				return tmpFile.Name()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile(t)
			defer func() {
				if strings.HasPrefix(filePath, os.TempDir()) {
					os.Remove(filePath)
				}
			}()

			_, err := extractPDFText(filePath)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestExtractDOCXText tests DOCX text extraction
func TestExtractDOCXText(t *testing.T) {
	tests := []struct {
		name      string
		setupFile func(t *testing.T) string
		wantErr   bool
	}{
		{
			name: "file not found",
			setupFile: func(t *testing.T) string {
				return "/nonexistent/file.docx"
			},
			wantErr: true,
		},
		{
			name: "invalid DOCX format - not a zip",
			setupFile: func(t *testing.T) string {
				tmpFile, err := os.CreateTemp("", "test-*.txt")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.WriteString("This is not a DOCX")
				tmpFile.Close()
				return tmpFile.Name()
			},
			wantErr: true,
		},
		{
			name: "missing document.xml",
			setupFile: func(t *testing.T) string {
				// Create a valid zip but without document.xml
				tmpFile, err := os.CreateTemp("", "test-*.docx")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()

				// Reopen file for writing
				zipFile, err := os.Create(tmpFile.Name())
				if err != nil {
					t.Fatalf("Failed to create zip file: %v", err)
				}
				zw := zip.NewWriter(zipFile)
				// Add a file that's not document.xml
				w, _ := zw.Create("other.xml")
				w.Write([]byte("<other>content</other>"))
				zw.Close()
				zipFile.Close()

				return tmpFile.Name()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile(t)
			defer func() {
				if strings.HasPrefix(filePath, os.TempDir()) {
					os.Remove(filePath)
				}
			}()

			_, err := extractDOCXText(filePath)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestParseDocumentXML tests XML parsing for DOCX
func TestParseDocumentXML(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		wantErr bool
		want    string
	}{
		{
			name: "valid XML with text",
			xml: `<?xml version="1.0"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:body>
		<w:p>
			<w:r>
				<w:t>Hello World</w:t>
			</w:r>
		</w:p>
	</w:body>
</w:document>`,
			wantErr: false,
			want:    "Hello World",
		},
		{
			name: "empty document",
			xml: `<?xml version="1.0"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:body>
	</w:body>
</w:document>`,
			wantErr: false,
			want:    "",
		},
		{
			name: "multiple paragraphs",
			xml: `<?xml version="1.0"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:body>
		<w:p>
			<w:r>
				<w:t>First paragraph</w:t>
			</w:r>
		</w:p>
		<w:p>
			<w:r>
				<w:t>Second paragraph</w:t>
			</w:r>
		</w:p>
	</w:body>
</w:document>`,
			wantErr: false,
			want:    "First paragraph\nSecond paragraph",
		},
		{
			name:    "malformed XML",
			xml:     "<invalid><unclosed>",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.xml)
			result, err := parseDocumentXML(reader)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !strings.Contains(result, tt.want) {
					t.Errorf("Expected result to contain %q, got %q", tt.want, result)
				}
			}
		})
	}
}

// TestExtractImageText tests image OCR extraction
func TestExtractImageText(t *testing.T) {
	// Check if tesseract is available
	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Skip("tesseract not available, skipping image OCR test")
	}

	tests := []struct {
		name      string
		setupFile func(t *testing.T) string
		wantErr   bool
	}{
		{
			name: "file not found",
			setupFile: func(t *testing.T) string {
				return "/nonexistent/file.png"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile(t)
			defer func() {
				if strings.HasPrefix(filePath, os.TempDir()) {
					os.Remove(filePath)
				}
			}()

			_, err := extractImageText(filePath)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
