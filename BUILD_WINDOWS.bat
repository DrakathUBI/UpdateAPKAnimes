@echo off
setlocal EnableExtensions
cd /d "%~dp0"

if not exist dist mkdir dist

rem Build para Windows sem depender de Python no cliente
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

go build -trimpath -ldflags "-s -w" -o dist\MacroRecorderSemAdmin.exe windows_macro_recorder.go
if errorlevel 1 (
  echo Falha no build.
  exit /b 1
)

echo Build concluido: dist\MacroRecorderSemAdmin.exe
exit /b 0
