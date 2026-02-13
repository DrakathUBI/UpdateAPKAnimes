@echo off
setlocal EnableExtensions
cd /d "%~dp0"

rem Evita prompt de elevacao (roda com privilegio do usuario atual)
set "__COMPAT_LAYER=RUNASINVOKER"

if not exist "%~dp0dist\MacroRecorderSemAdmin.exe" (
  echo Executavel nao encontrado em dist\MacroRecorderSemAdmin.exe
  echo Tentando gerar automaticamente...
  call "%~dp0BUILD_WINDOWS.bat"
  if errorlevel 1 (
    echo Nao foi possivel gerar o executavel agora.
    echo Abra BUILD_WINDOWS.bat em um Windows com Go instalado.
    pause
    exit /b 1
  )
)

echo Iniciando MacroRecorderSemAdmin como usuario comum...
"%~dp0dist\MacroRecorderSemAdmin.exe"

endlocal
