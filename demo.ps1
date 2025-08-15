# Beispiel PowerShell-Script für ZIP-Vergleich

# Test ZIP-Dateien erstellen
New-Item -ItemType Directory -Name "test_files" -Force | Out-Null
Set-Location "test_files"

"This is file 1" | Out-File -FilePath "file1.txt" -Encoding UTF8
"This is file 2" | Out-File -FilePath "file2_abc123.txt" -Encoding UTF8
"This is file 3" | Out-File -FilePath "file3.txt" -Encoding UTF8

Compress-Archive -Path "*.txt" -DestinationPath "../test1.zip" -Force

Remove-Item "file3.txt"
"This is file 2 modified" | Out-File -FilePath "file2_xyz789.txt" -Encoding UTF8
"This is file 4" | Out-File -FilePath "file4.txt" -Encoding UTF8

Compress-Archive -Path "*.txt" -DestinationPath "../test2.zip" -Force

Set-Location ".."
Remove-Item -Recurse -Force "test_files"

Write-Host "Test ZIP-Dateien erstellt: test1.zip und test2.zip"
Write-Host ""
Write-Host "Führe Vergleich aus (Konsolen-Ausgabe):"
./zipcompare.exe test1.zip test2.zip

Write-Host ""
Write-Host "Führe Vergleich mit XML-Report aus:"
./zipcompare.exe test1.zip test2.zip comparison_report.xml

Write-Host ""
Write-Host "XML-Report generiert: comparison_report.xml"
Write-Host "Öffne XML-Datei in Notepad..."
notepad comparison_report.xml
