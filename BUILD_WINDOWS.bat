@echo off
setlocal
cd /d "%~dp0"
if not exist dist mkdir dist
set GOOS=windows
set GOARCH=amd64
go build -o dist\MacroRecorderSemAdmin.exe windows_macro_recorder.go
if %errorlevel% neq 0 (
  echo Falha no build.
  exit /b 1
)
echo Build concluido: dist\MacroRecorderSemAdmin.exe
endlocal
