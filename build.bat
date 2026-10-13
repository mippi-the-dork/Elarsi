@echo off
setlocal
cd /d "%~dp0"
where go >nul 2>nul
if errorlevel 1 (
  echo ERROR: Install Go before building Elarsi.
  exit /b 1
)
where py >nul 2>nul
if errorlevel 1 (
  echo ERROR: Install Python 3 with the Windows py launcher.
  echo Python is needed only to embed the Explorer icon.
  exit /b 1
)
if not exist dist mkdir dist
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags="-s -w -H=windowsgui" -o "dist\Elarsi.exe" .
if errorlevel 1 exit /b 1
py -3 embed_icon.py "dist\Elarsi.exe" "elarsi.ico"
if errorlevel 1 exit /b 1
copy /Y README.md "dist\README.md" >nul
if errorlevel 1 exit /b 1
echo Elarsi built at dist\Elarsi.exe with Windows Explorer icon.
exit /b 0
