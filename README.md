# ZIP Compare Tool

Ein Go-Programm zum Vergleichen des Inhalts zweier ZIP-Dateien.

## Features

- Vergleicht den Inhalt zweier ZIP-Dateien
- **NEU**: Vergleicht ganze Verzeichnisse mit ZIP-Dateien
- Erkennt identische, unterschiedliche und fehlende Dateien
- Ignoriert Commit-Codes in Dateinamen (z.B. `file_abc123.txt` → `file.txt`)
- Verwendet SHA-256 Hash für Inhaltsvergleich
- Übersichtliche Ausgabe der Ergebnisse in der Konsole
- Optionale XML-Ausgabe mit detaillierten Diff-Informationen
- Automatische Binärdatei-Erkennung
- Line-by-Line Diff für Textdateien
- **NEU**: Batch-Verarbeitung mit automatischem ZIP-Pairing

## Installation

```bash
go mod tidy
go build -o zipcompare.exe
```

## Verwendung

### Einzelne ZIP-Dateien vergleichen

#### Basis-Vergleich (nur Konsolen-Ausgabe)
```bash
zipcompare.exe <zip1> <zip2>
```

#### Mit XML-Report
```bash
zipcompare.exe <zip1> <zip2> <output.xml>
```

### Verzeichnisse mit ZIP-Dateien vergleichen

#### Batch-Vergleich (nur Konsolen-Ausgabe)
```bash
zipcompare.exe <dir1> <dir2>
```

#### Mit XML-Reports für jedes Paar
```bash
zipcompare.exe <dir1> <dir2> <output_dir>
```

### Beispiele

```bash
# Einzelne ZIP-Dateien
zipcompare.exe archive1.zip archive2.zip

# Einzelne ZIP-Dateien mit XML-Report
zipcompare.exe archive1.zip archive2.zip comparison_report.xml

# Verzeichnisse vergleichen
zipcompare.exe releases_v1/ releases_v2/

# Verzeichnisse vergleichen mit XML-Reports
zipcompare.exe releases_v1/ releases_v2/ comparison_reports/
```

## Funktionsweise

1. **Dateiname-Normalisierung**: Dateien mit Namen wie `datei_abc123.txt` werden als `datei.txt` behandelt
2. **Binärdatei-Erkennung**: Automatische Erkennung von Binärdateien basierend auf Inhalt
3. **Inhaltsvergleich**: SHA-256 Hash wird für jeden Dateiinhalt berechnet
4. **Kategorisierung**: Dateien werden in folgende Kategorien eingeteilt:
   - ✅ Identisch (gleicher Inhalt)
   - ⚠️ Unterschiedlich (verschiedener Inhalt)
   - 📁 Nur in ZIP 1
   - 📁 Nur in ZIP 2
5. **Diff-Generierung**: Line-by-Line Diffs für Textdateien (nur in XML-Ausgabe)

## Verzeichnis-Vergleich Features

### Automatisches ZIP-Pairing
Das Tool findet automatisch passende ZIP-Dateien in zwei Verzeichnissen:
- Pairing basiert auf dem Namen bis zum letzten Unterstrich
- `package_v1.zip` und `package_v2.zip` → Paar: **package**
- `release_beta.zip` und `release_final.zip` → Paar: **release**

### Batch-Verarbeitung
- Verarbeitet alle gefundenen Paare automatisch
- Generiert für jedes Paar einen separaten XML-Report
- Übersichtliche Fortschrittsanzeige in der Konsole
- Sammelt alle Reports in einem Ausgabeverzeichnis

### Beispiel-Verzeichnisstruktur
```
releases_v1/
├── package_v1.0.zip
├── tools_beta.zip
└── docs_draft.zip

releases_v2/
├── package_v2.0.zip
├── tools_final.zip
└── docs_final.zip

Gefundene Paare:
• package (v1.0 ↔ v2.0)
• tools (beta ↔ final)
• docs (draft ↔ final)
```

## XML-Report Features

- **Strukturierte Daten**: Vollständige Vergleichsergebnisse in XML-Format
- **Diff-Details**: Detaillierte Line-by-Line Diffs für unterschiedliche Textdateien
- **Binärdatei-Kennzeichnung**: Binärdateien werden speziell markiert
- **Zeitstempel**: Automatischer Zeitstempel der Generierung
- **Zusammenfassung**: Statistische Übersicht aller Vergleichsergebnisse
- **Batch-Reports**: Bei Verzeichnis-Vergleich wird für jedes Paar ein separater Report erstellt

## Commit-Code Erkennung

Das Programm erkennt automatisch Commit-Codes am Ende von Dateinamen:
- `datei_a1b2c3.txt` → `datei.txt`
- `script_def456.js` → `script.js`
- `image_789abc.png` → `image.png`

Der Regex-Pattern: `^(.+)_[a-zA-Z0-9]{6,}(\.[^.]*)?$`

## XML-Report Beispiel

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
