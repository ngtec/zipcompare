# ZIP Compare Tool

A Go program for comparing the contents of two ZIP files.

## Features

- Compares the contents of two ZIP files
- **NEW**: Compares entire directories containing ZIP files
- Detects identical, different, and missing files
- Ignores commit codes in filenames (e.g., `file_abc123.txt` → `file.txt`)
- Uses SHA-256 hash for content comparison
- Clear console output of results
- Optional XML output with detailed diff information
- Automatic binary file detection
- Line-by-line diff for text files
- **NEW**: Batch processing with automatic ZIP pairing

## Installation

```bash
go mod tidy
go build -o zipcompare.exe
```

## Usage

### Compare Single ZIP Files

#### Basic comparison (console output only)
```bash
zipcompare.exe <zip1> <zip2>
```

#### With XML report
```bash
zipcompare.exe <zip1> <zip2> <output.xml>
```

### Compare Directories with ZIP Files

#### Batch comparison (console output only)
```bash
zipcompare.exe <dir1> <dir2>
```

#### With XML reports for each pair
```bash
zipcompare.exe <dir1> <dir2> <output_dir>
```

### Examples

```bash
# Single ZIP files
zipcompare.exe archive1.zip archive2.zip

# Single ZIP files with XML report
zipcompare.exe archive1.zip archive2.zip comparison_report.xml

# Compare directories
zipcompare.exe releases_v1/ releases_v2/

# Compare directories with XML reports
zipcompare.exe releases_v1/ releases_v2/ comparison_reports/
```

## How It Works

1. **Filename Normalization**: Files with names like `file_abc123.txt` are treated as `file.txt`
2. **Binary File Detection**: Automatic detection of binary files based on content
3. **Content Comparison**: SHA-256 hash is calculated for each file content
4. **Categorization**: Files are divided into the following categories:
   - ✅ Identical (same content)
   - ⚠️ Different (different content)
   - 📁 Only in ZIP 1
   - 📁 Only in ZIP 2
5. **Diff Generation**: Line-by-line diffs for text files (only in XML output)

## Directory Comparison Features

### Automatic ZIP Pairing
The tool automatically finds matching ZIP files in two directories:
- Pairing based on name up to the last underscore
- `package_v1.zip` and `package_v2.zip` → Pair: **package**
- `release_beta.zip` and `release_final.zip` → Pair: **release**

### Batch Processing
- Automatically processes all found pairs
- Generates a separate XML report for each pair
- Clear progress display in console
- Collects all reports in an output directory

### Example Directory Structure
```
releases_v1/
├── package_v1.0.zip
├── tools_beta.zip
└── docs_draft.zip

releases_v2/
├── package_v2.0.zip
├── tools_final.zip
└── docs_final.zip

Found pairs:
• package (v1.0 ↔ v2.0)
• tools (beta ↔ final)
• docs (draft ↔ final)
```

## XML Report Features

- **Structured Data**: Complete comparison results in XML format
- **Diff Details**: Detailed line-by-line diffs for different text files
- **Binary File Marking**: Binary files are specially marked
- **Timestamps**: Automatic generation timestamp
- **Summary**: Statistical overview of all comparison results
- **Batch Reports**: For directory comparison, a separate report is created for each pair

## Commit Code Detection

The program automatically recognizes commit codes at the end of filenames:
- `file_a1b2c3.txt` → `file.txt`
- `script_def456.js` → `script.js`
- `image_789abc.png` → `image.png`

Regex pattern: `^(.+)_[a-zA-Z0-9]{6,}(\.[^.]*)?$`

## XML Report Example

```xml
<?xml version="1.0" encoding="UTF-8"?>
<zipComparison generated="2025-08-14T10:30:00Z" zip1="archive1.zip" zip2="archive2.zip">
  <identical>
    <file>config.txt</file>
    <file>readme.md</file>
  </identical>
  <different>
    <file fileName="script.js" isBinary="false">
      <diff>--- script.js (ZIP 1)
+++ script.js (ZIP 2)
-console.log("old version");
+console.log("new version");
      </diff>
    </file>
    <file fileName="binary.exe" isBinary="true">
      <diff></diff>
    </file>
  </different>
  <onlyInFirst>
    <file>deprecated.txt</file>
  </onlyInFirst>
  <onlyInSecond>
    <file>newfeature.js</file>
  </onlyInSecond>
  <summary>
    <total>6</total>
    <identical>2</identical>
    <different>2</different>
    <onlyInFirst>1</onlyInFirst>
    <onlyInSecond>1</onlyInSecond>
  </summary>
</zipComparison>
```

## Command Line Arguments

- **Single file mode**: `zipcompare <zip1> <zip2> [output.xml]`
- **Directory mode**: `zipcompare <dir1> <dir2> [output_dir]`
- If the third argument is provided, XML reports will be generated
- For directory mode, XML files are named `{basename}_comparison.xml`

## Output Format

### Console Output
- Uses emojis and colors for better readability
- Shows statistics for each comparison
- Progress indication for batch processing

### XML Output
- Structured data suitable for further processing
- Contains complete diff information for text files
- Binary files are marked but contain no diff content
- Includes generation timestamp and source paths

## Technical Details

- **Language**: Go
- **Dependencies**: Standard library only
- **Hash Algorithm**: SHA-256 for content comparison
- **Binary Detection**: UTF-8 validation + null byte detection
- **Memory Usage**: Efficient - content is only stored when diffs are needed
- **Platform**: Cross-platform (Windows, Linux, macOS)

## License

This project is licensed under the APACHE 2.0 License - see the LICENSE file for details.
