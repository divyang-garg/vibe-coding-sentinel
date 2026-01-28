// Package ast provides language registry for dynamic language support
// Complies with CODING_STANDARDS.md: Registry modules max 250 lines
package ast

import (
	"fmt"
	"sync"
)

var (
	languageRegistry = make(map[string]LanguageSupport)
	registryMutex    sync.RWMutex
)

// RegisterLanguageSupport registers a language implementation
// Returns error if language is already registered or support is nil
func RegisterLanguageSupport(support LanguageSupport) error {
	if support == nil {
		return fmt.Errorf("language support cannot be nil")
	}

	lang := support.GetLanguage()
	if lang == "" {
		return fmt.Errorf("language name cannot be empty")
	}

	registryMutex.Lock()
	defer registryMutex.Unlock()

	if _, exists := languageRegistry[lang]; exists {
		return fmt.Errorf("language %s already registered", lang)
	}

	languageRegistry[lang] = support
	return nil
}

// GetLanguageSupport retrieves language support by name
// Returns nil if language is not registered
func GetLanguageSupport(language string) LanguageSupport {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	return languageRegistry[language]
}

// GetLanguageDetector retrieves detector for language
// Returns nil if language is not registered
func GetLanguageDetector(language string) LanguageDetector {
	support := GetLanguageSupport(language)
	if support == nil {
		return nil
	}
	return support.GetDetector()
}

// GetLanguageExtractor retrieves extractor for language
// Returns nil if language is not registered
func GetLanguageExtractor(language string) LanguageExtractor {
	support := GetLanguageSupport(language)
	if support == nil {
		return nil
	}
	return support.GetExtractor()
}

// GetSupportedLanguages returns list of registered languages
func GetSupportedLanguages() []string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	languages := make([]string, 0, len(languageRegistry))
	for lang := range languageRegistry {
		languages = append(languages, lang)
	}
	return languages
}

// IsLanguageSupported checks if a language is registered
func IsLanguageSupported(language string) bool {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	_, exists := languageRegistry[language]
	return exists
}
