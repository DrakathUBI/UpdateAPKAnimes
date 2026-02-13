@echo off
setlocal

set "DIR=%~dp0"
set "ARCH=%PROCESSOR_ARCHITECTURE%"
if not "%PROCESSOR_ARCHITEW6432%"=="" set "ARCH=%PROCESSOR_ARCHITEW6432%"
set "EXE="

if /I "%ARCH%"=="AMD64" set "EXE=%DIR%dist\MacroClone-win-x64.exe"
if /I "%ARCH%"=="x86" set "EXE=%DIR%dist\MacroClone-win-x86.exe"
if /I "%ARCH%"=="ARM64" set "EXE=%DIR%dist\MacroClone-win-arm64.exe"

if "%EXE%"=="" (
  echo Arquitetura nao suportada automaticamente: %ARCH%
  echo Tente executar manualmente um dos arquivos em .\dist\
  exit /b 1
)

if not exist "%EXE%" (
  echo Binario nao encontrado: %EXE%
  echo Gere os executaveis primeiro com:
  echo   build-windows.bat
  echo ou
  echo   build-windows.sh
  exit /b 1
)

"%EXE%" %*
