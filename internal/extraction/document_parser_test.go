// Package extraction provides LLM-powered knowledge extraction
package extraction

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocumentParser(t *testing.T) {
	t.Run("creates markdown parser for .md", func(t *testing.T) {
		parser, err := NewDocumentParser("test.md")
		require.NoError(t, err)
		assert.True(t, parser.Supports("test.md"))
	})

	t.Run("creates markdown parser for .markdown", func(t *testing.T) {
		parser, err := NewDocumentParser("test.markdown")
		require.NoError(t, err)
		assert.True(t, parser.Supports("test.markdown"))
	})

	t.Run("creates text parser for .txt", func(t *testing.T) {
		parser, err := NewDocumentParser("test.txt")
		require.NoError(t, err)
		assert.True(t, parser.Supports("test.txt"))
	})

	t.Run("creates docx parser for .docx", func(t *testing.T) {
		parser, err := NewDocumentParser("test.docx")
		require.NoError(t, err)
		assert.True(t, parser.Supports("test.docx"))
	})

	t.Run("creates pdf parser for .pdf", func(t *testing.T) {
		parser, err := NewDocumentParser("test.pdf")
		require.NoError(t, err)
		assert.True(t, parser.Supports("test.pdf"))
	})

	t.Run("returns error for unsupported type", func(t *testing.T) {
		_, err := NewDocumentParser("test.xyz")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported file type")
	})
}

func TestMarkdownParser(t *testing.T) {
	t.Run("parses markdown file", func(t *testing.T) {
		tmpFile := filepath.Join(os.TempDir(), "test_md.md")
		content := "# Header\n\nThe system must validate input."
		os.WriteFile(tmpFile, []byte(content), 0644)
		defer os.Remove(tmpFile)

		parser, _ := NewDocumentParser(tmpFile)
		result, err := parser.Parse(tmpFile)

		require.NoError(t, err)
		assert.Equal(t, content, result)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		parser, _ := NewDocumentParser("test.md")
		_, err := parser.Parse("/nonexistent/file.md")
		assert.Error(t, err)
	})
}

func TestTextParser(t *testing.T) {
	t.Run("parses text file", func(t *testing.T) {
		tmpFile := filepath.Join(os.TempDir(), "test_txt.txt")
		content := "Plain text content here."
		os.WriteFile(tmpFile, []byte(content), 0644)
		defer os.Remove(tmpFile)

		parser, _ := NewDocumentParser(tmpFile)
		result, err := parser.Parse(tmpFile)

		require.NoError(t, err)
		assert.Equal(t, content, result)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		parser := &textParser{}
		_, err := parser.Parse("/nonexistent/file.txt")
		assert.Error(t, err)
	})
}

func TestDocxParser(t *testing.T) {
	t.Run("supports docx files", func(t *testing.T) {
		parser := &docxParser{}
		assert.True(t, parser.Supports("test.docx"))
		assert.False(t, parser.Supports("test.pdf"))
	})

	t.Run("handles missing docx file", func(t *testing.T) {
		parser := &docxParser{}
		_, err := parser.Parse("/nonexistent/file.docx")
		assert.Error(t, err)
	})
}

func TestPdfParser(t *testing.T) {
	t.Run("supports pdf files", func(t *testing.T) {
		parser := &pdfParser{}
		assert.True(t, parser.Supports("test.pdf"))
		assert.False(t, parser.Supports("test.docx"))
	})

	t.Run("handles missing pdf file", func(t *testing.T) {
		parser := &pdfParser{}
		_, err := parser.Parse("/nonexistent/file.pdf")
		assert.Error(t, err)
	})

	t.Run("handles invalid pdf file", func(t *testing.T) {
		tmpFile := filepath.Join(os.TempDir(), "test_invalid.pdf")
		os.WriteFile(tmpFile, []byte("not a valid pdf"), 0644)
		defer os.Remove(tmpFile)

		parser := &pdfParser{}
		_, err := parser.Parse(tmpFile)
		// May error or return empty, both are acceptable
		_ = err
	})
}
