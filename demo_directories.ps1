# Demo für Verzeichnis-Vergleich

# Erstelle Test-Verzeichnisse
New-Item -ItemType Directory -Name "test_dir1" -Force | Out-Null
New-Item -ItemType Directory -Name "test_dir2" -Force | Out-Null
New-Item -ItemType Directory -Name "temp_files" -Force | Out-Null

# Erstelle verschiedene Test-Dateien
Set-Location "temp_files"

"Content for package v1" | Out-File -FilePath "file1.txt" -Encoding UTF8
"Common file content" | Out-File -FilePath "common.txt" -Encoding UTF8
Compress-Archive -Path "*.txt" -DestinationPath "../test_dir1/package_v1.zip" -Force

"Content for package v2" | Out-File -FilePath "file1.txt" -Encoding UTF8
"Common file content" | Out-File -FilePath "common.txt" -Encoding UTF8
"New feature file" | Out-File -FilePath "newfeature.txt" -Encoding UTF8
Compress-Archive -Path "*.txt" -DestinationPath "../test_dir2/package_v2.zip" -Force

Remove-Item "*.txt"

"Release v1 content" | Out-File -FilePath "main.js" -Encoding UTF8
"Config for v1" | Out-File -FilePath "config.json" -Encoding UTF8
Compress-Archive -Path "*.js", "*.json" -DestinationPath "../test_dir1/release_v1.zip" -Force

"Release v2 content modified" | Out-File -FilePath "main.js" -Encoding UTF8
"Config for v2" | Out-File -FilePath "config.json" -Encoding UTF8
Compress-Archive -Path "*.js", "*.json" -DestinationPath "../test_dir2/release_v2.zip" -Force

Remove-Item "*.js", "*.json"

# Erstelle ungepaarte Datei
"Unmatched content" | Out-File -FilePath "unmatched.txt" -Encoding UTF8
Compress-Archive -Path "unmatched.txt" -DestinationPath "../test_dir1/unmatched_single.zip" -Force

Set-Location ".."
Remove-Item -Recurse -Force "temp_files"

Write-Host "Test-Verzeichnisse erstellt:"
Write-Host "test_dir1:" -ForegroundColor Green
Get-ChildItem "test_dir1" | ForEach-Object { Write-Host "  - $($_.Name)" }
Write-Host "test_dir2:" -ForegroundColor Green
Get-ChildItem "test_dir2" | ForEach-Object { Write-Host "  - $($_.Name)" }

Write-Host ""
Write-Host "Führe Verzeichnis-Vergleich aus..." -ForegroundColor Yellow
./zipcompare.exe test_dir1 test_dir2 comparison_reports

Write-Host ""
Write-Host "Erstelle Reports:" -ForegroundColor Yellow
Get-ChildItem "comparison_reports" | ForEach-Object { 
    Write-Host "  📄 $($_.Name)" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "Öffne einen beispielhaften XML-Report..." -ForegroundColor Yellow
$firstReport = Get-ChildItem "comparison_reports" | Select-Object -First 1
if ($firstReport) {
    notepad $firstReport.FullName
}

Write-Host ""
Write-Host "Demo abgeschlossen! Aufräumen..." -ForegroundColor Green
# Remove-Item -Recurse -Force "test_dir1", "test_dir2", "comparison_reports"
