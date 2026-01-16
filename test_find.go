package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func findFilesRecursively(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			matched, err := filepath.Match(pattern, info.Name())
			if err != nil {
				return nil
			}
			if matched {
				matches = append(matches, path)
			}
		}
		return nil
	})
	return matches, err
}

func main() {
	tempDir := "/tmp/test_find"
	os.MkdirAll(tempDir+"/routes", 0755)
	os.WriteFile(tempDir+"/routes/userRoutes.js", []byte("test"), 0644)
	
	files, err := findFilesRecursively(tempDir, "*.js")
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Found files: %v\n", files)
}
EOF && go run test_find.go && rm test_find.go && rm -rf /tmp/test_find