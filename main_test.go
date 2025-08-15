package main

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestZip(files map[string]string) (string, error) {
	tmpFile, err := os.CreateTemp("", "test*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close()

	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(writer, bytes.NewReader([]byte(content)))
		if err != nil {
			return "", err
		}
	}

	return tmpFile.Name(), nil
}

func TestExtractBaseName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file.txt", "file.txt"},
		{"file_abc123.txt", "file.txt"},
		{"script_def456.js", "script.js"},
		{"image_789abc.png", "image.png"},
		{"document_a1b2c3", "document"},
		{"folder/file_xyz789.pdf", "file.pdf"},
		{"no_commit_code.doc", "no_commit_code.doc"},
	}

	for _, test := range tests {
		result := extractBaseName(filepath.Base(test.input))
		if result != test.expected {
			t.Errorf("extractBaseName(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestCompareIdenticalZips(t *testing.T) {
	// Create two identical ZIP files
	files := map[string]string{
		"file1.txt":       "content1",
		"file2_abc123.js": "content2",
		"file3.png":       "binary content",
	}

	zip1, err := createTestZip(files)
	if err != nil {
		t.Fatalf("Failed to create test ZIP 1: %v", err)
	}
	defer os.Remove(zip1)

	zip2, err := createTestZip(files)
	if err != nil {
		t.Fatalf("Failed to create test ZIP 2: %v", err)
	}
	defer os.Remove(zip2)

	result, err := compareZipFiles(zip1, zip2)
	if err != nil {
		t.Fatalf("compareZipFiles failed: %v", err)
	}

	if len(result.Identical) != 3 {
		t.Errorf("Expected 3 identical files, got %d", len(result.Identical))
	}

	if len(result.Different) != 0 {
		t.Errorf("Expected 0 different files, got %d", len(result.Different))
	}

	if len(result.OnlyInFirst) != 0 {
		t.Errorf("Expected 0 files only in first ZIP, got %d", len(result.OnlyInFirst))
	}

	if len(result.OnlyInSecond) != 0 {
		t.Errorf("Expected 0 files only in second ZIP, got %d", len(result.OnlyInSecond))
	}
}

func TestCompareZipsWithCommitCodes(t *testing.T) {
	// Create ZIP files with different commit codes but same base names
	files1 := map[string]string{
		"file_abc123.txt":  "same content",
		"script_def456.js": "same script",
	}

	files2 := map[string]string{
		"file_xyz789.txt":  "same content",
		"script_uvw012.js": "same script",
	}

	zip1, err := createTestZip(files1)
	if err != nil {
		t.Fatalf("Failed to create test ZIP 1: %v", err)
	}
	defer os.Remove(zip1)

	zip2, err := createTestZip(files2)
	if err != nil {
		t.Fatalf("Failed to create test ZIP 2: %v", err)
	}
	defer os.Remove(zip2)

	result, err := compareZipFiles(zip1, zip2)
	if err != nil {
		t.Fatalf("compareZipFiles failed: %v", err)
	}

	if len(result.Identical) != 2 {
		t.Errorf("Expected 2 identical files, got %d", len(result.Identical))
	}

	// Check that the base names are correctly identified
	expectedBaseNames := map[string]bool{"file.txt": true, "script.js": true}
	for _, fileName := range result.Identical {
		if !expectedBaseNames[fileName] {
			t.Errorf("Unexpected base name in identical files: %s", fileName)
		}
	}
}

func TestCompareDifferentZips(t *testing.T) {
	files1 := map[string]string{
		"common.txt":      "same content",
		"different.txt":   "content in zip1",
		"onlyinfirst.txt": "only in first",
	}

	files2 := map[string]string{
		"common.txt":       "same content",
		"different.txt":    "content in zip2",
		"onlyinsecond.txt": "only in second",
	}

	zip1, err := createTestZip(files1)
	if err != nil {
		t.Fatalf("Failed to create test ZIP 1: %v", err)
	}
	defer os.Remove(zip1)

	zip2, err := createTestZip(files2)
	if err != nil {
		t.Fatalf("Failed to create test ZIP 2: %v", err)
	}
	defer os.Remove(zip2)

	result, err := compareZipFiles(zip1, zip2)
	if err != nil {
		t.Fatalf("compareZipFiles failed: %v", err)
	}

	if len(result.Identical) != 1 {
		t.Errorf("Expected 1 identical file, got %d", len(result.Identical))
	}

	if len(result.Different) != 1 {
		t.Errorf("Expected 1 different file, got %d", len(result.Different))
	}

	if len(result.OnlyInFirst) != 1 {
		t.Errorf("Expected 1 file only in first ZIP, got %d", len(result.OnlyInFirst))
	}

	if len(result.OnlyInSecond) != 1 {
		t.Errorf("Expected 1 file only in second ZIP, got %d", len(result.OnlyInSecond))
	}
}

func TestXMLReportGeneration(t *testing.T) {
	// Create test files with different content
	files1 := map[string]string{
		"same.txt":      "identical content",
		"different.txt": "content in zip1\nline 2\nline 3",
		"binary.exe":    "\x00\x01\x02\x03", // Binary content
	}

	files2 := map[string]string{
		"same.txt":      "identical content",
		"different.txt": "content in zip2\nmodified line 2\nline 3",
		"binary.exe":    "\x00\x01\x02\x04", // Different binary content
	}

	zip1, err := createTestZip(files1)
	if err != nil {
		t.Fatalf("Failed to create test ZIP 1: %v", err)
	}
	defer os.Remove(zip1)

	zip2, err := createTestZip(files2)
	if err != nil {
		t.Fatalf("Failed to create test ZIP 2: %v", err)
	}
	defer os.Remove(zip2)

	result, err := compareZipFiles(zip1, zip2)
	if err != nil {
		t.Fatalf("compareZipFiles failed: %v", err)
	}

	// Generate XML report
	xmlFile := filepath.Join(os.TempDir(), "test_report.xml")
	defer os.Remove(xmlFile)

	err = generateXMLReport(result, zip1, zip2, xmlFile)
	if err != nil {
		t.Fatalf("generateXMLReport failed: %v", err)
	}

	// Read and verify XML content
	xmlContent, err := os.ReadFile(xmlFile)
	if err != nil {
		t.Fatalf("Failed to read XML file: %v", err)
	}

	xmlStr := string(xmlContent)

	// Check that XML contains expected elements
	if !strings.Contains(xmlStr, "<zipComparison") {
		t.Error("XML should contain zipComparison root element")
	}

	if !strings.Contains(xmlStr, "<identical>") {
		t.Error("XML should contain identical section")
	}

	if !strings.Contains(xmlStr, "<different>") {
		t.Error("XML should contain different section")
	}

	if !strings.Contains(xmlStr, "same.txt") {
		t.Error("XML should contain same.txt in identical files")
	}

	if !strings.Contains(xmlStr, "different.txt") {
		t.Error("XML should contain different.txt in different files")
	}

	// Check for diff content in text files
	if !strings.Contains(xmlStr, "-content in zip1") {
		t.Error("XML should contain diff for text files")
	}

	// Check that binary files are marked as binary
	if !strings.Contains(xmlStr, "isBinary=\"true\"") {
		t.Error("XML should mark binary files as binary")
	}
}

func TestIsBinaryContent(t *testing.T) {
	tests := []struct {
		content  []byte
		expected bool
	}{
		{[]byte("hello world"), false},
		{[]byte("text with\nnewlines"), false},
		{[]byte{0x00, 0x01, 0x02}, true}, // Binary with null bytes
		{[]byte{0xFF, 0xFE}, true},       // Invalid UTF-8
		{[]byte(""), false},              // Empty content
	}

	for _, test := range tests {
		result := isBinaryContent(test.content)
		if result != test.expected {
			t.Errorf("isBinaryContent(%v) = %v; want %v", test.content, result, test.expected)
		}
	}
}

func TestExtractZipBaseName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file.zip", "file"},
		{"package_v1.2.3.zip", "package"},
		{"release_20230815_final.zip", "release_20230815"},
		{"simple.zip", "simple"},
		{"no_extension_here", "no_extension"}, // Corrected expectation
		{"multiple_under_scores_here.zip", "multiple_under_scores"},
	}

	for _, test := range tests {
		result := extractZipBaseName(test.input)
		if result != test.expected {
			t.Errorf("extractZipBaseName(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestFindZipPairs(t *testing.T) {
	// Create temporary directories
	tempDir := os.TempDir()

	dir1 := filepath.Join(tempDir, "test_zip_dir1")
	dir2 := filepath.Join(tempDir, "test_zip_dir2")

	os.MkdirAll(dir1, 0755)
	os.MkdirAll(dir2, 0755)
	defer os.RemoveAll(dir1)
	defer os.RemoveAll(dir2)

	// Create test ZIP files
	testFiles := map[string]string{"test.txt": "content"}

	// Directory 1: package_v1.zip, release_final.zip
	zip1, _ := createTestZip(testFiles)
	defer os.Remove(zip1)
	os.Rename(zip1, filepath.Join(dir1, "package_v1.zip"))

	zip2, _ := createTestZip(testFiles)
	defer os.Remove(zip2)
	os.Rename(zip2, filepath.Join(dir1, "release_final.zip"))

	// Directory 2: package_v2.zip, release_beta.zip, unmatched.zip
	zip3, _ := createTestZip(testFiles)
	defer os.Remove(zip3)
	os.Rename(zip3, filepath.Join(dir2, "package_v2.zip"))

	zip4, _ := createTestZip(testFiles)
	defer os.Remove(zip4)
	os.Rename(zip4, filepath.Join(dir2, "release_beta.zip"))

	zip5, _ := createTestZip(testFiles)
	defer os.Remove(zip5)
	os.Rename(zip5, filepath.Join(dir2, "unmatched.zip"))

	// Find pairs
	pairs, err := findZipPairs(dir1, dir2)
	if err != nil {
		t.Fatalf("findZipPairs failed: %v", err)
	}

	// Should find 2 pairs: package and release
	if len(pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(pairs))
	}

	// Check that we found the right pairs
	foundPackage := false
	foundRelease := false
	for _, pair := range pairs {
		if pair.BaseName == "package" {
			foundPackage = true
		}
		if pair.BaseName == "release" {
			foundRelease = true
		}
	}

	if !foundPackage {
		t.Error("Expected to find 'package' pair")
	}
	if !foundRelease {
		t.Error("Expected to find 'release' pair")
	}
}
