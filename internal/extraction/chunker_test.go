// Package extraction provides LLM-powered knowledge extraction
package extraction

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextChunker(t *testing.T) {
	t.Run("returns single chunk for small text", func(t *testing.T) {
		chunker := NewTextChunker(4000)
		text := "Short text."

		chunks := chunker.Chunk(text, 4000)

		assert.Len(t, chunks, 1)
		assert.Equal(t, text, chunks[0])
	})

	t.Run("splits by paragraphs", func(t *testing.T) {
		chunker := NewTextChunker(100) // Low limit to force split
		text := "Paragraph one with content.\n\nParagraph two with more content."

		chunks := chunker.Chunk(text, 100)

		assert.Greater(t, len(chunks), 0)
	})

	t.Run("handles large paragraphs by sentence", func(t *testing.T) {
		chunker := NewTextChunker(50) // Very low limit
		text := "Sentence one. Sentence two. Sentence three. Sentence four."

		chunks := chunker.Chunk(text, 50)

		assert.Greater(t, len(chunks), 0)
	})

	t.Run("handles empty text", func(t *testing.T) {
		chunker := NewTextChunker(4000)
		text := ""

		chunks := chunker.Chunk(text, 4000)

		assert.Len(t, chunks, 1)
		assert.Equal(t, "", chunks[0])
	})

	t.Run("uses default when maxTokens is 0", func(t *testing.T) {
		chunker := NewTextChunker(0)
		text := "Some text."

		chunks := chunker.Chunk(text, 0)

		assert.Len(t, chunks, 1)
	})

	t.Run("handles very large document", func(t *testing.T) {
		chunker := NewTextChunker(4000)
		// Create ~40K character text
		text := strings.Repeat("The system must validate input. ", 1000)

		chunks := chunker.Chunk(text, 4000)

		assert.Greater(t, len(chunks), 1)
		// Each chunk should be reasonable size
		for _, chunk := range chunks {
			assert.LessOrEqual(t, len(chunk), 16000) // 4000 tokens * 4 chars
		}
	})
}

func TestSplitBySentences(t *testing.T) {
	t.Run("splits by period", func(t *testing.T) {
		text := "First sentence. Second sentence."
		sentences := splitBySentences(text)

		assert.Len(t, sentences, 2)
	})

	t.Run("splits by question mark", func(t *testing.T) {
		text := "Is this a question? Yes it is."
		sentences := splitBySentences(text)

		assert.Len(t, sentences, 2)
	})

	t.Run("splits by exclamation", func(t *testing.T) {
		text := "Wow! That is amazing."
		sentences := splitBySentences(text)

		assert.Len(t, sentences, 2)
	})
}
