@echo off
setlocal
cd /d "%~dp0"
if not exist dist mkdir dist

go env >nul 2>nul
if errorlevel 1 (
  echo Go nao encontrado no PATH.
  exit /b 1
)

set CGO_ENABLED=0
set GOOS=windows

set GOARCH=amd64
call go build -o dist\MacroClone-win-x64.exe .
if errorlevel 1 exit /b 1

set GOARCH=386
call go build -o dist\MacroClone-win-x86.exe .
if errorlevel 1 exit /b 1

set GOARCH=arm64
call go build -o dist\MacroClone-win-arm64.exe .
if errorlevel 1 exit /b 1

echo Build concluido em .\dist
