# Macro Recorder para Windows (sem Python no PC)

Atualização: implementei agora um **motor de script com múltiplas funções da documentação** (além do gravar/reproduzir simples).

## Arquivos principais
- `windows_macro_recorder.go` → programa principal (Windows)
- `BUILD_WINDOWS.bat` → gera `dist\MacroRecorderSemAdmin.exe`
- `INICIAR_WINDOWS.bat` → inicia o `.exe`
- `macro_script_exemplo.json` → exemplo de script avançado

## Build (máquina de empacotamento)
Execute no Windows:

```bat
BUILD_WINDOWS.bat
```

Saída:
- `dist\MacroRecorderSemAdmin.exe`

## Funções adicionadas agora
### Modo 1: Gravação de eventos
- grava movimento de mouse
- grava clique esquerdo e direito
- `F9` para parar

### Modo 2: Reprodução de eventos
- reproduz JSON de eventos
- `ESC` para interromper

### Modo 3: Script avançado (ações)
Ações suportadas no JSON:
- `wait_time` (com randomização)
- `wait_until_time`
- `wait_hotkey`
- `wait_file_exists` (com `on_timeout_goto`)
- `mouse_move`
- `mouse_click` (left/right/middle)
- `mouse_scroll`
- `key_press`
- `hotkey`
- `text_output` (com `humanize`)
- `set_var`
- `calc` (+, -, *, /)
- `save_var_file` (com append)
- `show_notification`
- `show_message`
- `beep`
- `run_program`
- `if` (`equals`, `contains`, `!=`) com `then/else`
- `goto`
- `repeat`
- `data_list_next`

## Execução
1. Abra `INICIAR_WINDOWS.bat`
2. Escolha:
   - `1` gravação
   - `2` reprodução
   - `3` script avançado

## Observações
- Usuário final **não precisa de Python**.
- Fluxo pensado para sem Admin.
- Em ambiente corporativo, antivírus/política pode exigir whitelist para automação.
