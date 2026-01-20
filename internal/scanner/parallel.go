// Package scanner provides parallel scanning functionality
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// ScanParallel performs parallel scanning using goroutines
func ScanParallel(opts ScanOptions) (*Result, error) {
	result := &Result{
		Success:   true,
		Findings:  []Finding{},
		Summary:   make(map[string]int),
		Timestamp: getCurrentTimestamp(),
	}

	patterns := GetSecurityPatterns()

	// Determine scan directory
	scanDir := opts.CodebasePath
	if scanDir == "" {
		scanDir = "."
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(scanDir)
	if err != nil {
		absPath = scanDir
	}

	// Collect all files first
	files, err := collectFiles(scanDir)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return result, nil
	}

	// Use worker pool pattern with configurable concurrency
	maxWorkers := runtime.NumCPU()
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	// Channel to collect findings
	findingsChan := make(chan []Finding, len(files))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)

	// Process files in parallel
	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			findings := scanFile(f, patterns, absPath)
			findingsChan <- findings
		}(file)
	}

	// Close channel when all workers done
	go func() {
		wg.Wait()
		close(findingsChan)
	}()

	// Aggregate results
	fileContents := make(map[string]string)
	for fileFindings := range findingsChan {
		for _, finding := range fileFindings {
			// Store file content for filtering
			if _, exists := fileContents[finding.File]; !exists {
				content, _ := os.ReadFile(filepath.Join(absPath, finding.File))
				fileContents[finding.File] = string(content)
			}

			result.Findings = append(result.Findings, finding)
			result.Summary[finding.Type]++

			// Mark as failed if critical finding
			if finding.Severity == SeverityCritical {
				result.Success = false
			}
		}
	}

	// Apply false positive filtering
	for file, content := range fileContents {
		fileFindings := make([]Finding, 0)
		for _, f := range result.Findings {
			if f.File == file {
				fileFindings = append(fileFindings, f)
			}
		}
		filtered := FilterFalsePositives(fileFindings, content)
		// Update result with filtered findings
		newFindings := make([]Finding, 0)
		for _, f := range result.Findings {
			if f.File != file {
				newFindings = append(newFindings, f)
			}
		}
		result.Findings = append(newFindings, filtered...)
	}

	// Recalculate summary after filtering
	result.Summary = make(map[string]int)
	for _, f := range result.Findings {
		result.Summary[f.Type]++
	}

	// Apply baseline filtering
	result = filterBaselineParallel(result)

	return result, nil
}
