@echo off
setlocal

cd /d "%~dp0"
if not exist dist mkdir dist

set "EXIT_CODE=0"
set "PAUSE_ON_EXIT=1"
if /I "%~1"=="--no-pause" set "PAUSE_ON_EXIT=0"

echo [INFO] Pasta do projeto: %CD%
echo [INFO] Verificando Go no PATH...
where go >nul 2>nul
if errorlevel 1 (
  echo [ERRO] Go nao encontrado no PATH.
  echo [DICA] Instale Go: https://go.dev/dl/
  set "EXIT_CODE=1"
  goto :END
)

go version
if errorlevel 1 (
  echo [ERRO] Falha ao executar "go version".
  set "EXIT_CODE=1"
  goto :END
)

set CGO_ENABLED=0
set GOOS=windows

echo [INFO] Build x64...
set GOARCH=amd64
go build -o dist\MacroClone-win-x64.exe .
if errorlevel 1 (
  echo [ERRO] Falha no build x64.
  set "EXIT_CODE=1"
  goto :END
)

echo [INFO] Build x86...
set GOARCH=386
go build -o dist\MacroClone-win-x86.exe .
if errorlevel 1 (
  echo [ERRO] Falha no build x86.
  set "EXIT_CODE=1"
  goto :END
)

echo [INFO] Build arm64...
set GOARCH=arm64
go build -o dist\MacroClone-win-arm64.exe .
if errorlevel 1 (
  echo [ERRO] Falha no build arm64.
  set "EXIT_CODE=1"
  goto :END
)

echo [OK] Build concluido em .\dist

dir /b dist\MacroClone-win-*.exe

:END
echo.
if "%EXIT_CODE%"=="0" (
  echo [INFO] Finalizado com sucesso.
) else (
  echo [INFO] Finalizado com erro (codigo %EXIT_CODE%).
)

if "%PAUSE_ON_EXIT%"=="1" pause
exit /b %EXIT_CODE%
