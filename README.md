# Macro Recorder para Windows (sem Python no PC)

## Objetivo
Executar automação em Windows com usuário comum (sem admin e sem Python instalado no cliente).

## Arquivos
- `windows_macro_recorder.go` → motor principal
- `BUILD_WINDOWS.bat` → compila `dist\MacroRecorderSemAdmin.exe`
- `INICIAR_WINDOWS.bat` → inicia sem elevação (RUNASINVOKER)
- `macro_script_exemplo.json` → exemplo de script

## Build no Windows
```bat
BUILD_WINDOWS.bat
```

## Execução
```bat
INICIAR_WINDOWS.bat
```

## Melhorias solicitadas aplicadas
- `.bat` inicia em modo usuário comum (sem pedir admin).
- `.bat` tenta auto-build se o `.exe` não existir.
- Executor de script expandido com ampla cobertura de ações/documentação.

## Ações suportadas no script JSON
### Implementadas funcionalmente
- waits: `wait_time`, `wait_until_time`, `wait_hotkey`, `wait_file_exists`, `wait_text_input`
- mouse/teclado: `mouse_move`, `mouse_click`, `mouse_scroll`, `key_press`, `hotkey`, `text_output`
- variáveis/fluxo: `set_var`, `calc`, `if`, `goto`, `repeat`, `data_list_next`, `save_var_file`
- utilidades: `show_notification`, `show_message`, `beep`, `run_program`, `scrape_webpage`, `embed_macro`

### Compatibilidade (placeholder/no-op no executor console)
- `smart_click`, `ai_object_search`, `phrase_insertion`, `window_focus`, `post_playback`
- `find_image`, `find_text`, `find_text_ocr`, `wait_pixel_change`, `wait_desktop_change`
- `capture_bitmap`, `capture_text`, `capture_text_ocr`, `capture_barcode`, `capture_qr`
- metadados: `label`, `comment`, `group_start`, `group_end`, `breakpoint`, `disable_action`

## Observações
- O usuário final não precisa de Python.
- Em ambiente corporativo, antivírus/políticas podem exigir whitelist para automação.
