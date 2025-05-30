# Scripts

This directory contains utility scripts for repository maintenance and quality assurance.

## test-audit.go

A comprehensive test audit tool that scans Go test files for common quality issues.

### What it detects

- **Skipped tests**: `t.Skip()` calls and skip annotations
- **Commented out tests**: Disabled test functions
- **Disabled test files**: Build tags that exclude tests
- **Empty tests**: Functions with no meaningful implementation
- **Incomplete tests**: TODO/FIXME markers in test code

### Usage

```bash
# Audit all test files in current directory
go run scripts/test-audit.go .

# Audit specific directory
go run scripts/test-audit.go ./auth

# Use in CI/CD pipeline
go run scripts/test-audit.go . && echo "Test audit passed"
```

### Output

The tool provides:
- Summary statistics (files, lines, test functions examined)
- Issues grouped by severity (high, medium, low)
- List of clean files with no issues
- Exit code 0 for clean suite, 1 if issues found

### Integration

Add to your CI/CD pipeline to ensure test quality:

```yaml
# GitHub Actions example
- name: Test Audit
  run: go run scripts/test-audit.go .
```

### Example Output

```
=== TEST AUDIT REPORT ===
Total test files examined: 6
Total lines of test code: 4,787
Total test functions: 150
Issues found: 0
Clean files: 6

✅ NO ISSUES FOUND - Test suite is clean!

Clean files:
  ✅ auth/client_test.go
  ✅ ai/client_test.go
  ✅ storage/client_test.go
  ✅ ingest/client_test.go
  ✅ internal/clientutil/client_test.go
  ✅ internal/apierror/apierror_test.go
```