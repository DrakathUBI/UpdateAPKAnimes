# MacroClone (Windows, sem admin e sem dotnet)

Você reportou que o GitHub não aceita binário. Ajustei o projeto para **não versionar executáveis**.

## O que mudou
- Binários `.exe` removidos do versionamento Git.
- Arquivo `.gitignore` para ignorar `dist/*.exe`.
- Scripts de build para gerar os executáveis localmente:
  - `build-windows.bat`
  - `build-windows.sh`
- Launcher `run-macroclone.bat` agora avisa quando o binário não existir e orienta gerar build.

## Gerar executáveis no Windows
```powershell
.\build-windows.bat
```

## Gerar executáveis no Linux/macOS (cross-compile)
```bash
./build-windows.sh
```

## Executar
```powershell
.\run-macroclone.bat --play .\macro.sample.json
```

## CLI
- `--play <arquivo>`
- `--silent`
- `--start-at <label>`

## Funcionalidades implementadas no runtime sem admin
- Fluxo: `Goto`, `IfThenElse`, `Repeat`.
- Variáveis: `SetVariable`, expansão `${variavel}`.
- Tempo/eventos: `WaitTime`, `WaitUntilTime`, `WaitForFileEvent`.
- Execução/utilidades: `ExecuteProgram`, `SaveVariable`, `TextOutput`, `ShowNotification`, `ShowMessageBox` (modo console), `WaitForTextInput`, `WaitForHotkey`, `Beep`.

## Sobre ações avançadas de UI/visão
Ações como mouse/keyboard hook global, OCR, detecção visual e foco de janela ficam em **modo seguro sem-admin** (no-op com log), pois exigem integrações de baixo nível e permissões/SDKs específicos do Windows.
