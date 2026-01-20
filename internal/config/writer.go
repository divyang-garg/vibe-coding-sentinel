// Package config provides configuration file writing utilities
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package config

import (
	"os"

	"github.com/divyang-garg/sentinel-hub-api/internal/constants"
)

// WriteFile writes content to a file with default permissions
func WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), constants.DefaultFilePerm)
}

// SecureGitIgnore appends Sentinel-specific entries to .gitignore
func SecureGitIgnore() error {
	content := "\n# Sentinel Rules\n.cursor/rules/*.md\n!.cursor/rules/00-constitution.md\nsentinel\n"
	f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// CreateCI generates a GitHub Actions workflow file for Sentinel
func CreateCI() error {
	content := `name: Sentinel Gate
on: [push, pull_request]
jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Build Sentinel
        run: chmod +x synapsevibsentinel.sh && ./synapsevibsentinel.sh
      - name: Run Audit
        run: ./sentinel audit --ci
        continue-on-error: false
`
	return WriteFile(".github/workflows/sentinel.yml", content)
}
