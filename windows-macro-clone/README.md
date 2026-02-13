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
### Jeito mais simples (duplo clique)
- Dê duplo clique em `build-windows.bat`.
- A janela **não fecha automaticamente**; ela mostra erros e aguarda tecla no final.

### Pelo PowerShell/CMD
```powershell
.\build-windows.bat
```

### Rodar sem pausa no final (automação)
```powershell
.\build-windows.bat --no-pause
```

Se aparecer `[ERRO] Go nao encontrado no PATH`, instale o Go e abra um novo terminal:
- https://go.dev/dl/

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
