# Mega Prompt Avançado (pt-BR)

## Objetivo
Gerar, com IA de código, um aplicativo desktop multiplataforma (Windows + macOS) que replique **apenas** as funcionalidades documentadas no site oficial `macrorecorder.com`, sem inferir recursos não documentados.

## Regras obrigatórias para a IA geradora
- Não inventar funcionalidades.
- Quando algo não estiver descrito no site, marcar como **NÃO ESPECIFICADO** e:
  - implementar como opção conservadora; ou
  - deixar stub/TODO explícito.
- Recursos marcados como experimentais devem ser opcionais e desligados por padrão.
- Variáveis devem ser voláteis (descartadas ao fechar o app).

## Escopo funcional mínimo

### 1) Gravação e edição
- Record / Stop / Play.
- Gravar mouse, teclado, foco de janela e pausas.
- Lista cronológica de ações com edição por duplo-clique/Enter.
- Multi-seleção (CTRL/SHIFT), drag&drop, copiar/colar, delete.
- Labels (únicas), comentários, line numbers, breakpoints, grupos, desativar ações.

### 2) Reprodução
- Stop por botão e ESC (sempre).
- Execução parcial por modificadores (ALT/CTRL/SHIFT + Play).
- Restauração de posição/tamanho de janelas.
- Playback profile: velocidade, repetições, filtro por tipo de ação, mouse path modes, post-playback action com countdown cancelável.

### 3) Ações de mouse e teclado
- Mouse click (left/middle/right/X1/X2), coordenadas fixas/variáveis, offset, wiggle.
- Mouse move, scroll-wheel, SmartClick.
- Keypress (down/up), Hotkey action, Text output (clipboard vs typing), Humanize.
- Phrase insertion por referência (integração opcional/degradável).

### 4) Waits, detecção visual e OCR
- Wait time (com randomize), wait until time, wait for hotkey/text/file.
- Wait for pixel color, wait for desktop change (com timeout + branching).
- Find image (pixel/feature matching, teste, timeout + label).
- Find text OCR (literal/regex, timeout + label).
- OCR com idiomas adicionais via `tessdata`.

### 5) Captura e scraping
- Screenshot para variável imagem.
- OCR “wait for any text” para variável texto.
- Leitura de Barcode/QR para variável.
- Extração de texto de webpage por URL + XPath-derivate/Regex.

### 6) Variáveis e controle de fluxo
- Create/set/use variable.
- Cálculo com operadores/funções matemáticas.
- Salvar variável em arquivo/clipboard (append + path com variáveis).
- Data list por texto manual ou TXT.
- Goto, Repeat, If-Then-Else, Window focus, Execute program, Embed macro.

### 7) Debug e observabilidade
- Show notification, message box com botões e labels, beep.
- Variable Explorer (texto/imagem, copiar/salvar).
- Breakpoints e execução pausável para inspeção.

### 8) Configurações
- Recording, Playback, Hotkeys, UI language, AI settings, reset de settings.
- Tray/menu bar com controle quando minimizado.

### 9) Segurança e permissões
- Fluxos claros para permissões macOS (Accessibility, Screen Recording, Full Disk Access).
- Aviso quando alvo exige privilégio elevado.
- Se usar IA online: consentimento explícito + aviso de transmissão de dados + API keys criptografadas e não reveláveis.

### 10) CLI
Implementar no mínimo:
- `--play <macro-file>`
- `--silent`

Opcional:
- `--start-at <label>`
- `--exit`, `--shutdown`, `--restart`

## Entregáveis exigidos
1. Estrutura de monorepo:
   - `/apps/desktop-ui`
   - `/packages/core-engine`
   - `/packages/os-adapters`
   - `/packages/vision-ocr`
   - `/packages/ai-connectors` (opcional)
   - `/docs`
2. Build scripts para Windows/macOS.
3. Testes unitários e de integração (modo headless com mocks).
4. Documentação:
   - `README.md`
   - `SPEC.md` (incluindo seção “NÃO ESPECIFICADO”)
   - `CLI.md`
   - matriz de rastreabilidade requisito → módulo → teste.

## Casos de teste de referência
- OCR + If-Then-Else + Repeat + Data list.
- SmartClick + fallback Find image.
- Post-playback action com countdown cancelável.

## Ponto crítico de escopo
### Exportar para EXE/Script
Tratar como **NÃO ESPECIFICADO** no escopo estrito da documentação consultada em `macrorecorder.com`.
Se implementar, marcar como extensão fora do escopo de fidelidade.

## Lista de fontes-base (macrorecorder.com)
- `https://www.macrorecorder.com/doc/`
- `https://www.macrorecorder.com/doc/first-steps/`
- `https://www.macrorecorder.com/doc/create-macro/record/`
- `https://www.macrorecorder.com/doc/create-macro/step-by-step/`
- `https://www.macrorecorder.com/doc/create-macro/smart-recording/`
- `https://www.macrorecorder.com/doc/edit/`
- `https://www.macrorecorder.com/doc/playback/`
- `https://www.macrorecorder.com/doc/wait/`
- `https://www.macrorecorder.com/doc/mouse/`
- `https://www.macrorecorder.com/doc/keyboard/`
- `https://www.macrorecorder.com/doc/control/`
- `https://www.macrorecorder.com/doc/variables/`
- `https://www.macrorecorder.com/doc/find/`
- `https://www.macrorecorder.com/doc/capture/`
- `https://www.macrorecorder.com/doc/debugging-tools/`
- `https://www.macrorecorder.com/doc/settings/`
- `https://www.macrorecorder.com/doc/settings/recording/`
- `https://www.macrorecorder.com/doc/settings/playback/`
- `https://www.macrorecorder.com/doc/settings/hotkeys/`
- `https://www.macrorecorder.com/doc/settings/user-interface/`
- `https://www.macrorecorder.com/doc/settings/ai/`
- `https://www.macrorecorder.com/doc/settings/reset/`
- `https://www.macrorecorder.com/doc/settings/adding-languages/`
- `https://www.macrorecorder.com/doc/license/`
- `https://www.macrorecorder.com/doc/installation/`
- `https://www.macrorecorder.com/doc/file/`
- `https://www.macrorecorder.com/doc/command-line-parameter/`
- `https://www.macrorecorder.com/doc/reference/file-locations/`
- `https://www.macrorecorder.com/doc/reference/image-detection-method/`
- `https://www.macrorecorder.com/doc/troubleshooting/`
