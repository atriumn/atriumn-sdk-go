// Package main provides a test audit tool for Go repositories
// to identify skipped, commented, or disabled tests.
package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TestIssue represents a potential test quality issue
type TestIssue struct {
	File        string
	Line        int
	Type        string
	Description string
	Severity    string
	Code        string
}

// AuditResult contains the results of a test audit
type AuditResult struct {
	TotalFiles    int
	TotalLines    int
	Issues        []TestIssue
	CleanFiles    []string
	TestFunctions int
}

// TestAuditor performs comprehensive test suite audits
type TestAuditor struct {
	patterns map[string]*regexp.Regexp
}

// NewTestAuditor creates a new test auditor with predefined patterns
func NewTestAuditor() *TestAuditor {
	patterns := map[string]*regexp.Regexp{
		// Skipped tests
		"skip_call":        regexp.MustCompile(`\.Skip\s*\(`),
		"skip_annotation":  regexp.MustCompile(`@skip|@Skip|@SKIP`),
		
		// Commented tests
		"commented_func":   regexp.MustCompile(`^[\s]*//.*func\s+Test\w+`),
		"block_comment":    regexp.MustCompile(`^[\s]*/\*.*func\s+Test\w+`),
		
		// Build exclusions
		"build_ignore":     regexp.MustCompile(`^[\s]*//\s*\+build\s+ignore`),
		"go_build_ignore":  regexp.MustCompile(`^[\s]*//go:build\s+ignore`),
		
		// Quality issues
		"empty_test":       regexp.MustCompile(`func\s+Test\w+\([^)]*\)\s*{\s*}`),
		"immediate_return": regexp.MustCompile(`func\s+Test\w+\([^)]*\)\s*{\s*return\s*}`),
		
		// TODO/FIXME patterns
		"todo_comment":     regexp.MustCompile(`//.*(?i)(todo|fixme|pending|hack).*test`),
		"todo_test":        regexp.MustCompile(`func\s+Test\w*(?i)(todo|fixme|pending)\w*`),
		
		// Test function counter
		"test_function":    regexp.MustCompile(`func\s+Test\w+`),
	}
	
	return &TestAuditor{patterns: patterns}
}

// AuditDirectory performs a comprehensive audit of all test files in a directory
func (ta *TestAuditor) AuditDirectory(dir string) (*AuditResult, error) {
	result := &AuditResult{
		Issues:     make([]TestIssue, 0),
		CleanFiles: make([]string, 0),
	}
	
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		
		fileResult, err := ta.auditFile(path)
		if err != nil {
			return fmt.Errorf("error auditing file %s: %w", path, err)
		}
		
		result.TotalFiles++
		result.TotalLines += fileResult.TotalLines
		result.TestFunctions += fileResult.TestFunctions
		
		if len(fileResult.Issues) == 0 {
			result.CleanFiles = append(result.CleanFiles, path)
		} else {
			result.Issues = append(result.Issues, fileResult.Issues...)
		}
		
		return nil
	})
	
	return result, err
}

// fileAuditResult represents the audit result for a single file
type fileAuditResult struct {
	Issues        []TestIssue
	TotalLines    int
	TestFunctions int
}

// auditFile examines a single test file for issues
func (ta *TestAuditor) auditFile(filename string) (*fileAuditResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	result := &fileAuditResult{
		Issues: make([]TestIssue, 0),
	}
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		result.TotalLines = lineNum
		
		// Count test functions
		if ta.patterns["test_function"].MatchString(line) {
			result.TestFunctions++
		}
		
		// Check for skipped tests
		if ta.patterns["skip_call"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "skipped_test",
				Description: "Test contains t.Skip() call",
				Severity:    "medium",
				Code:        strings.TrimSpace(line),
			})
		}
		
		if ta.patterns["skip_annotation"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "skipped_test",
				Description: "Test has skip annotation",
				Severity:    "medium",
				Code:        strings.TrimSpace(line),
			})
		}
		
		// Check for commented tests
		if ta.patterns["commented_func"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "commented_test",
				Description: "Commented out test function",
				Severity:    "high",
				Code:        strings.TrimSpace(line),
			})
		}
		
		if ta.patterns["block_comment"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "commented_test",
				Description: "Block commented test function",
				Severity:    "high",
				Code:        strings.TrimSpace(line),
			})
		}
		
		// Check for build exclusions
		if ta.patterns["build_ignore"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "disabled_test",
				Description: "File has build ignore directive",
				Severity:    "high",
				Code:        strings.TrimSpace(line),
			})
		}
		
		if ta.patterns["go_build_ignore"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "disabled_test",
				Description: "File has go:build ignore directive",
				Severity:    "high",
				Code:        strings.TrimSpace(line),
			})
		}
		
		// Check for quality issues
		if ta.patterns["empty_test"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "empty_test",
				Description: "Empty test function with no implementation",
				Severity:    "medium",
				Code:        strings.TrimSpace(line),
			})
		}
		
		if ta.patterns["immediate_return"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "empty_test",
				Description: "Test function returns immediately",
				Severity:    "medium",
				Code:        strings.TrimSpace(line),
			})
		}
		
		// Check for TODO/FIXME patterns
		if ta.patterns["todo_comment"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "incomplete_test",
				Description: "Test has TODO/FIXME comment",
				Severity:    "low",
				Code:        strings.TrimSpace(line),
			})
		}
		
		if ta.patterns["todo_test"].MatchString(line) {
			result.Issues = append(result.Issues, TestIssue{
				File:        filename,
				Line:        lineNum,
				Type:        "incomplete_test",
				Description: "Test function name suggests TODO/FIXME",
				Severity:    "low",
				Code:        strings.TrimSpace(line),
			})
		}
	}
	
	return result, scanner.Err()
}

// PrintReport generates a human-readable audit report
func (ta *TestAuditor) PrintReport(result *AuditResult) {
	fmt.Println("=== TEST AUDIT REPORT ===")
	fmt.Printf("Total test files examined: %d\n", result.TotalFiles)
	fmt.Printf("Total lines of test code: %d\n", result.TotalLines)
	fmt.Printf("Total test functions: %d\n", result.TestFunctions)
	fmt.Printf("Issues found: %d\n", len(result.Issues))
	fmt.Printf("Clean files: %d\n", len(result.CleanFiles))
	fmt.Println()
	
	if len(result.Issues) == 0 {
		fmt.Println("✅ NO ISSUES FOUND - Test suite is clean!")
		fmt.Println()
		fmt.Println("Clean files:")
		for _, file := range result.CleanFiles {
			fmt.Printf("  ✅ %s\n", file)
		}
		return
	}
	
	// Group issues by severity
	severityGroups := make(map[string][]TestIssue)
	for _, issue := range result.Issues {
		severityGroups[issue.Severity] = append(severityGroups[issue.Severity], issue)
	}
	
	// Print issues by severity
	severities := []string{"high", "medium", "low"}
	for _, severity := range severities {
		issues := severityGroups[severity]
		if len(issues) == 0 {
			continue
		}
		
		fmt.Printf("🚨 %s SEVERITY ISSUES (%d):\n", strings.ToUpper(severity), len(issues))
		for _, issue := range issues {
			fmt.Printf("  %s:%d - %s\n", issue.File, issue.Line, issue.Description)
			fmt.Printf("    Type: %s\n", issue.Type)
			fmt.Printf("    Code: %s\n", issue.Code)
			fmt.Println()
		}
	}
	
	// Print clean files
	if len(result.CleanFiles) > 0 {
		fmt.Println("✅ CLEAN FILES:")
		for _, file := range result.CleanFiles {
			fmt.Printf("  %s\n", file)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-audit.go <directory>")
		fmt.Println("Example: go run test-audit.go .")
		os.Exit(1)
	}
	
	directory := os.Args[1]
	
	auditor := NewTestAuditor()
	result, err := auditor.AuditDirectory(directory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error performing audit: %v\n", err)
		os.Exit(1)
	}
	
	auditor.PrintReport(result)
	
	// Exit with non-zero code if issues found
	if len(result.Issues) > 0 {
		os.Exit(1)
	}
}