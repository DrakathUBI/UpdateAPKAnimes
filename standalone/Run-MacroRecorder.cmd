@echo off
setlocal
set SCRIPT=%~dp0MacroRecorderStandalone.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File "%SCRIPT%"
endlocal
