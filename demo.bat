@echo off
echo Creating test ZIP files for demonstration...

mkdir test_files 2>nul
cd test_files

echo This is file 1 > file1.txt
echo This is file 2 > file2_abc123.txt
echo This is file 3 > file3.txt

powershell Compress-Archive -Path *.txt -DestinationPath ../test1.zip -Force

del file3.txt
echo This is file 2 modified > file2_xyz789.txt
echo This is file 4 > file4.txt

powershell Compress-Archive -Path *.txt -DestinationPath ../test2.zip -Force

cd ..
rmdir /s /q test_files

echo.
echo Test ZIP files created: test1.zip and test2.zip
echo.
echo Running comparison (console output):
zipcompare.exe test1.zip test2.zip

echo.
echo Running comparison with XML report:
zipcompare.exe test1.zip test2.zip comparison_report.xml

echo.
echo XML report generated: comparison_report.xml
echo Opening XML file in notepad...
notepad comparison_report.xml

pause
