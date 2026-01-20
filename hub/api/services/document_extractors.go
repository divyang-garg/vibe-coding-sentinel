// Package services - Document text extraction utilities
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// extractPDFText extracts text from PDF using pdftotext (poppler-utils)
func extractPDFText(filePath string) (string, error) {
	cmd := exec.Command("pdftotext", "-layout", filePath, "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// extractDOCXText extracts text from DOCX using archive/zip and XML parsing
func extractDOCXText(filePath string) (string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open docx: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open document.xml: %w", err)
			}
			defer rc.Close()

			// Parse XML and extract text content
			return parseDocumentXML(rc)
		}
	}
	return "", fmt.Errorf("document.xml not found in docx")
}

// parseDocumentXML parses Word document XML and extracts text from <w:t> elements
func parseDocumentXML(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var textParts []string
	var currentText strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to parse XML: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" && se.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				currentText.Reset()
			}
		case xml.CharData:
			if currentText.Len() > 0 || len(bytes.TrimSpace(se)) > 0 {
				currentText.Write(se)
			}
		case xml.EndElement:
			if se.Name.Local == "t" && se.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				if currentText.Len() > 0 {
					textParts = append(textParts, currentText.String())
					currentText.Reset()
				}
			} else if se.Name.Local == "p" && se.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				// Add paragraph break
				if len(textParts) > 0 {
					textParts = append(textParts, "\n")
				}
			}
		}
	}

	result := strings.Join(textParts, "")
	// Clean up multiple newlines
	result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	return strings.TrimSpace(result), nil
}

// extractImageText extracts text from image using Tesseract OCR
func extractImageText(filePath string) (string, error) {
	cmd := exec.Command("tesseract", filePath, "stdout", "-l", "eng")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
