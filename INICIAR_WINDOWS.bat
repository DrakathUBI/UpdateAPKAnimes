@echo off
setlocal
cd /d "%~dp0"
if not exist "%~dp0dist\MacroRecorderSemAdmin.exe" (
  echo Executavel nao encontrado em dist\MacroRecorderSemAdmin.exe
  echo Rode BUILD_WINDOWS.bat para gerar o programa.
  pause
  exit /b 1
)
"%~dp0dist\MacroRecorderSemAdmin.exe"
endlocal
