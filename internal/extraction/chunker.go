// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
	"strings"
)

// Chunker splits text into manageable chunks for processing
type Chunker interface {
	Chunk(text string, maxTokens int) []string
}

// textChunker implements Chunker for text documents
type textChunker struct {
	maxTokens int
}

// NewTextChunker creates a new text chunker
func NewTextChunker(maxTokens int) Chunker {
	return &textChunker{maxTokens: maxTokens}
}

// Chunk splits text into chunks that don't exceed maxTokens
// Uses paragraph boundaries first, then sentence boundaries
func (c *textChunker) Chunk(text string, maxTokens int) []string {
	if maxTokens <= 0 {
		maxTokens = c.maxTokens
	}
	if maxTokens == 0 {
		maxTokens = 4000 // Default: ~1000 tokens
	}

	// Estimate: 1 token ≈ 4 characters
	maxChars := maxTokens * 4

	// Split by paragraphs first
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var currentChunk strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// If paragraph alone exceeds limit, split by sentences
		if len(para) > maxChars {
			// Add current chunk if it has content
			if currentChunk.Len() > 0 {
				chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
				currentChunk.Reset()
			}

			// Split paragraph by sentences
			sentences := splitBySentences(para)
			for _, sent := range sentences {
				sent = strings.TrimSpace(sent)
				if sent == "" {
					continue
				}

				if currentChunk.Len()+len(sent)+2 > maxChars {
					if currentChunk.Len() > 0 {
						chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
						currentChunk.Reset()
					}
				}
				if currentChunk.Len() > 0 {
					currentChunk.WriteString(" ")
				}
				currentChunk.WriteString(sent)
			}
		} else {
			// Check if adding this paragraph would exceed limit
			if currentChunk.Len() > 0 && currentChunk.Len()+len(para)+2 > maxChars {
				chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
				currentChunk.Reset()
			}
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n\n")
			}
			currentChunk.WriteString(para)
		}
	}

	// Add remaining chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	if len(chunks) == 0 {
		// Fallback: return entire text as single chunk
		chunks = []string{text}
	}

	return chunks
}

// splitBySentences splits text by sentence boundaries
func splitBySentences(text string) []string {
	// Simple sentence splitting by common punctuation
	replacer := strings.NewReplacer(
		". ", ".\n",
		"! ", "!\n",
		"? ", "?\n",
	)
	text = replacer.Replace(text)
	return strings.Split(text, "\n")
}
