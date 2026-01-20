// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pdf "github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

// DocumentParser interface for parsing different document types
type DocumentParser interface {
	Parse(filePath string) (string, error)
	Supports(filePath string) bool
}

// NewDocumentParser creates a document parser for the given file
func NewDocumentParser(filePath string) (DocumentParser, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".md", ".markdown":
		return &markdownParser{}, nil
	case ".txt":
		return &textParser{}, nil
	case ".docx":
		return &docxParser{}, nil
	case ".pdf":
		return &pdfParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// markdownParser handles Markdown files
type markdownParser struct{}

func (p *markdownParser) Supports(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".md" || ext == ".markdown"
}

func (p *markdownParser) Parse(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read markdown file: %w", err)
	}
	return string(content), nil
}

// textParser handles plain text files
type textParser struct{}

func (p *textParser) Supports(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".txt"
}

func (p *textParser) Parse(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read text file: %w", err)
	}
	return string(content), nil
}

// docxParser handles DOCX files
type docxParser struct{}

func (p *docxParser) Supports(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".docx"
}

func (p *docxParser) Parse(filePath string) (string, error) {
	doc, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX file: %w", err)
	}
	defer doc.Close()

	return doc.Editable().GetContent(), nil
}

// pdfParser handles PDF files
type pdfParser struct{}

func (p *pdfParser) Supports(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".pdf"
}

func (p *pdfParser) Parse(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	reader, err := pdf.NewReader(f, info.Size())
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var text strings.Builder
	fontMap := make(map[string]*pdf.Font)
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		pageText, _ := page.GetPlainText(fontMap)
		text.WriteString(pageText)
		if i < reader.NumPage() {
			text.WriteString("\n")
		}
	}

	return text.String(), nil
}
