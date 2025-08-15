package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type FileInfo struct {
	Name     string
	BaseName string // Name without commit code
	Size     int64
	Hash     string
	Content  string // Store content for diff generation
	IsBinary bool   // Track if file is binary
}

type DiffInfo struct {
	FileName string `xml:"fileName"`
	Diff     string `xml:"diff"`
	IsBinary bool   `xml:"isBinary,attr"`
}

type XMLReport struct {
	XMLName      xml.Name   `xml:"zipComparison"`
	Generated    string     `xml:"generated,attr"`
	Zip1         string     `xml:"zip1,attr"`
	Zip2         string     `xml:"zip2,attr"`
	Identical    []string   `xml:"identical>file"`
	Different    []DiffInfo `xml:"different>file"`
	OnlyInFirst  []string   `xml:"onlyInFirst>file"`
	OnlyInSecond []string   `xml:"onlyInSecond>file"`
	Summary      Summary    `xml:"summary"`
}

type Summary struct {
	Total        int `xml:"total"`
	Identical    int `xml:"identical"`
	Different    int `xml:"different"`
	OnlyInFirst  int `xml:"onlyInFirst"`
	OnlyInSecond int `xml:"onlyInSecond"`
}

type ComparisonResult struct {
	OnlyInFirst  []string
	OnlyInSecond []string
	Different    []string
	Identical    []string
	DiffDetails  []DiffInfo // Store detailed diff information
}

type ZipPair struct {
	BaseName string
	Zip1Path string
	Zip2Path string
}

func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		fmt.Println("Usage:")
		fmt.Println("  zipcompare <zip1> <zip2> [output.xml]           - Compare two ZIP files")
		fmt.Println("  zipcompare <dir1> <dir2> [output_dir]           - Compare ZIP files in directories")
		fmt.Println("    If output.xml is specified, results will be saved to XML file")
		fmt.Println("    If output_dir is specified, XML reports will be saved there")
		os.Exit(1)
	}

	path1 := os.Args[1]
	path2 := os.Args[2]
	var outputPath string
	if len(os.Args) == 4 {
		outputPath = os.Args[3]
	}

	// Check if paths are directories or files
	info1, err := os.Stat(path1)
	if err != nil {
		log.Fatalf("Error accessing %s: %v", path1, err)
	}

	info2, err := os.Stat(path2)
	if err != nil {
		log.Fatalf("Error accessing %s: %v", path2, err)
	}

	if info1.IsDir() && info2.IsDir() {
		// Directory comparison mode
		err = compareDirectories(path1, path2, outputPath)
		if err != nil {
			log.Fatalf("Error comparing directories: %v", err)
		}
	} else if !info1.IsDir() && !info2.IsDir() {
		// Single file comparison mode (existing functionality)
		result, err := compareZipFiles(path1, path2)
		if err != nil {
			log.Fatalf("Error comparing ZIP files: %v", err)
		}

		printResults(result)

		if outputPath != "" {
			err = generateXMLReport(result, path1, path2, outputPath)
			if err != nil {
				log.Fatalf("Error generating XML report: %v", err)
			}
			fmt.Printf("\n📄 XML-Report gespeichert: %s\n", outputPath)
		}
	} else {
		log.Fatalf("Both paths must be either files or directories")
	}
}

// extractBaseName removes commit codes from filenames
// If filename ends with _<commitcode>, the commit code is removed
func extractBaseName(filename string) string {
	// Regex pattern to match commit codes at the end of filenames
	// Assumes commit code is at least 6 characters long and alphanumeric
	re := regexp.MustCompile(`^(.+)_[a-zA-Z0-9]{6,}(\.[^.]*)?$`)
	matches := re.FindStringSubmatch(filename)

	if len(matches) >= 2 {
		baseName := matches[1]
		if len(matches) >= 3 && matches[2] != "" {
			baseName += matches[2] // Add file extension back
		}
		return baseName
	}

	return filename
}

// extractZipBaseName extracts the base name from ZIP file name (everything before last underscore)
func extractZipBaseName(zipFileName string) string {
	// Remove .zip extension
	name := strings.TrimSuffix(zipFileName, ".zip")

	// Find last underscore
	lastUnderscore := strings.LastIndex(name, "_")
	if lastUnderscore == -1 {
		return name
	}

	return name[:lastUnderscore]
}

// findZipPairs finds matching ZIP files in two directories
func findZipPairs(dir1, dir2 string) ([]ZipPair, error) {
	// Read ZIP files from first directory
	files1, err := filepath.Glob(filepath.Join(dir1, "*.zip"))
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", dir1, err)
	}

	// Read ZIP files from second directory
	files2, err := filepath.Glob(filepath.Join(dir2, "*.zip"))
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", dir2, err)
	}

	// Create map of base names to full paths for second directory
	dir2Map := make(map[string]string)
	for _, file2 := range files2 {
		baseName := extractZipBaseName(filepath.Base(file2))
		dir2Map[baseName] = file2
	}

	// Find matching pairs
	var pairs []ZipPair
	for _, file1 := range files1 {
		baseName := extractZipBaseName(filepath.Base(file1))
		if file2, exists := dir2Map[baseName]; exists {
			pairs = append(pairs, ZipPair{
				BaseName: baseName,
				Zip1Path: file1,
				Zip2Path: file2,
			})
		}
	}

	return pairs, nil
}

// compareDirectories compares all matching ZIP files in two directories
func compareDirectories(dir1, dir2, outputDir string) error {
	fmt.Printf("🔍 Suche nach ZIP-Dateien in Verzeichnissen...\n")
	fmt.Printf("   Verzeichnis 1: %s\n", dir1)
	fmt.Printf("   Verzeichnis 2: %s\n", dir2)
	fmt.Println()

	pairs, err := findZipPairs(dir1, dir2)
	if err != nil {
		return err
	}

	if len(pairs) == 0 {
		fmt.Println("❌ Keine passenden ZIP-Dateien gefunden!")
		return nil
	}

	fmt.Printf("✅ %d passende ZIP-Paare gefunden:\n", len(pairs))
	for _, pair := range pairs {
		fmt.Printf("   • %s\n", pair.BaseName)
	}
	fmt.Println()

	// Create output directory if specified
	if outputDir != "" {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Process each pair
	for i, pair := range pairs {
		fmt.Printf("📊 Vergleiche %d/%d: %s\n", i+1, len(pairs), pair.BaseName)

		result, err := compareZipFiles(pair.Zip1Path, pair.Zip2Path)
		if err != nil {
			fmt.Printf("   ❌ Fehler beim Vergleichen: %v\n", err)
			continue
		}

		// Print summary for this pair
		totalFiles := len(result.Identical) + len(result.Different) + len(result.OnlyInFirst) + len(result.OnlyInSecond)
		fmt.Printf("   📁 Dateien: %d | ✅ Identisch: %d | ⚠️  Unterschiedlich: %d | 📋 Nur in 1: %d | 📋 Nur in 2: %d\n",
			totalFiles, len(result.Identical), len(result.Different), len(result.OnlyInFirst), len(result.OnlyInSecond))

		// Generate XML report if output directory is specified
		if outputDir != "" {
			xmlFileName := fmt.Sprintf("%s_comparison.xml", pair.BaseName)
			xmlPath := filepath.Join(outputDir, xmlFileName)

			err = generateXMLReport(result, pair.Zip1Path, pair.Zip2Path, xmlPath)
			if err != nil {
				fmt.Printf("   ❌ Fehler beim Erstellen des XML-Reports: %v\n", err)
			} else {
				fmt.Printf("   📄 XML-Report: %s\n", xmlFileName)
			}
		}
		fmt.Println()
	}

	if outputDir != "" {
		fmt.Printf("🎉 Alle Vergleiche abgeschlossen! XML-Reports gespeichert in: %s\n", outputDir)
	} else {
		fmt.Printf("🎉 Alle Vergleiche abgeschlossen!\n")
	}

	return nil
}

// isBinaryContent checks if content is binary
func isBinaryContent(content []byte) bool {
	// Check if content is valid UTF-8 and doesn't contain null bytes
	if !utf8.Valid(content) {
		return true
	}

	// Check for null bytes which indicate binary content
	for _, b := range content {
		if b == 0 {
			return true
		}
	}

	return false
}

// generateDiff creates a simple line-by-line diff
func generateDiff(content1, content2, fileName string) string {
	if content1 == content2 {
		return ""
	}

	lines1 := strings.Split(content1, "\n")
	lines2 := strings.Split(content2, "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- %s (ZIP 1)\n", fileName))
	diff.WriteString(fmt.Sprintf("+++ %s (ZIP 2)\n", fileName))

	maxLines := len(lines1)
	if len(lines2) > maxLines {
		maxLines = len(lines2)
	}

	for i := 0; i < maxLines; i++ {
		var line1, line2 string
		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			if i < len(lines1) {
				diff.WriteString(fmt.Sprintf("-%s\n", line1))
			}
			if i < len(lines2) {
				diff.WriteString(fmt.Sprintf("+%s\n", line2))
			}
		}
	}

	return diff.String()
}

// readZipContents reads a ZIP file and returns file information
func readZipContents(zipPath string) (map[string]FileInfo, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP file %s: %w", zipPath, err)
	}
	defer reader.Close()

	files := make(map[string]FileInfo)

	for _, file := range reader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s in ZIP: %w", file.Name, err)
		}

		// Read file content into memory
		content, err := io.ReadAll(fileReader)
		if err != nil {
			fileReader.Close()
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}
		fileReader.Close()

		// Calculate hash of file content
		hash := sha256.Sum256(content)

		baseName := extractBaseName(filepath.Base(file.Name))

		// Check if content is binary
		isBinary := isBinaryContent(content)
		var contentStr string
		if !isBinary {
			contentStr = string(content)
		}

		fileInfo := FileInfo{
			Name:     file.Name,
			BaseName: baseName,
			Size:     int64(len(content)),
			Hash:     fmt.Sprintf("%x", hash),
			Content:  contentStr,
			IsBinary: isBinary,
		}

		// Use full path + baseName as key to handle duplicates
		key := baseName
		if existingFile, exists := files[key]; exists {
			// If we have a duplicate base name, prefer the one without commit code
			if len(existingFile.Name) > len(file.Name) {
				files[key] = fileInfo
			}
		} else {
			files[key] = fileInfo
		}
	}

	return files, nil
} // compareZipFiles compares two ZIP files and returns the comparison result
func compareZipFiles(zip1Path, zip2Path string) (*ComparisonResult, error) {
	files1, err := readZipContents(zip1Path)
	if err != nil {
		return nil, fmt.Errorf("error reading first ZIP file: %w", err)
	}

	files2, err := readZipContents(zip2Path)
	if err != nil {
		return nil, fmt.Errorf("error reading second ZIP file: %w", err)
	}

	result := &ComparisonResult{
		OnlyInFirst:  []string{},
		OnlyInSecond: []string{},
		Different:    []string{},
		Identical:    []string{},
		DiffDetails:  []DiffInfo{},
	}

	// Check files in first ZIP
	for baseName, file1 := range files1 {
		if file2, exists := files2[baseName]; exists {
			if file1.Hash == file2.Hash && file1.Size == file2.Size {
				result.Identical = append(result.Identical, baseName)
			} else {
				result.Different = append(result.Different, baseName)

				// Generate diff for non-binary files
				var diff string
				isBinary := file1.IsBinary || file2.IsBinary
				if !isBinary {
					diff = generateDiff(file1.Content, file2.Content, baseName)
				}

				result.DiffDetails = append(result.DiffDetails, DiffInfo{
					FileName: baseName,
					Diff:     diff,
					IsBinary: isBinary,
				})
			}
		} else {
			result.OnlyInFirst = append(result.OnlyInFirst, baseName)
		}
	}

	// Check files only in second ZIP
	for baseName := range files2 {
		if _, exists := files1[baseName]; !exists {
			result.OnlyInSecond = append(result.OnlyInSecond, baseName)
		}
	}

	return result, nil
}

// printResults prints the comparison results in a readable format
func printResults(result *ComparisonResult) {
	fmt.Println("=== ZIP-Datei Vergleich ===")
	fmt.Println()

	if len(result.Identical) > 0 {
		fmt.Printf("✅ Identische Dateien (%d):\n", len(result.Identical))
		for _, file := range result.Identical {
			fmt.Printf("  • %s\n", file)
		}
		fmt.Println()
	}

	if len(result.Different) > 0 {
		fmt.Printf("⚠️  Unterschiedliche Dateien (%d):\n", len(result.Different))
		for _, file := range result.Different {
			fmt.Printf("  • %s\n", file)
		}
		fmt.Println()
	}

	if len(result.OnlyInFirst) > 0 {
		fmt.Printf("📁 Nur in der ersten ZIP-Datei (%d):\n", len(result.OnlyInFirst))
		for _, file := range result.OnlyInFirst {
			fmt.Printf("  • %s\n", file)
		}
		fmt.Println()
	}

	if len(result.OnlyInSecond) > 0 {
		fmt.Printf("📁 Nur in der zweiten ZIP-Datei (%d):\n", len(result.OnlyInSecond))
		for _, file := range result.OnlyInSecond {
			fmt.Printf("  • %s\n", file)
		}
		fmt.Println()
	}

	// Summary
	totalFiles := len(result.Identical) + len(result.Different) + len(result.OnlyInFirst) + len(result.OnlyInSecond)
	fmt.Printf("📊 Zusammenfassung:\n")
	fmt.Printf("  Gesamt Dateien: %d\n", totalFiles)
	fmt.Printf("  Identisch: %d\n", len(result.Identical))
	fmt.Printf("  Unterschiedlich: %d\n", len(result.Different))
	fmt.Printf("  Nur in ZIP 1: %d\n", len(result.OnlyInFirst))
	fmt.Printf("  Nur in ZIP 2: %d\n", len(result.OnlyInSecond))

	if len(result.Different) == 0 && len(result.OnlyInFirst) == 0 && len(result.OnlyInSecond) == 0 {
		fmt.Println("\n🎉 Die ZIP-Dateien sind identisch!")
	} else {
		fmt.Println("\n⚠️  Die ZIP-Dateien unterscheiden sich.")
	}
}

// generateXMLReport creates an XML report with detailed comparison results
func generateXMLReport(result *ComparisonResult, zip1Path, zip2Path, outputPath string) error {
	totalFiles := len(result.Identical) + len(result.Different) + len(result.OnlyInFirst) + len(result.OnlyInSecond)

	report := XMLReport{
		Generated:    time.Now().Format(time.RFC3339),
		Zip1:         zip1Path,
		Zip2:         zip2Path,
		Identical:    result.Identical,
		Different:    result.DiffDetails,
		OnlyInFirst:  result.OnlyInFirst,
		OnlyInSecond: result.OnlyInSecond,
		Summary: Summary{
			Total:        totalFiles,
			Identical:    len(result.Identical),
			Different:    len(result.Different),
			OnlyInFirst:  len(result.OnlyInFirst),
			OnlyInSecond: len(result.OnlyInSecond),
		},
	}

	// Create XML content
	xmlData, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML header
	xmlContent := []byte(xml.Header + string(xmlData))

	// Write to file
	err = os.WriteFile(outputPath, xmlContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write XML file: %w", err)
	}

	return nil
}
