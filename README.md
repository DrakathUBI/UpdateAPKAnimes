# Macro Recorder Windows No-Admin (UI completa + standalone)

Você pediu sem npm, sem Python e sem instalar linguagem: agora o projeto tem **duas formas no mesmo pacote**, incluindo uma **standalone real para Windows**.

## Opção A (recomendada para seu cenário): Standalone sem instalar nada

Use os arquivos da pasta `standalone/`:

- `Run-MacroRecorder.cmd` → clique duplo para abrir.
- `MacroRecorderStandalone.ps1` → engine + UI completa em PowerShell/Windows Forms.

### Como abrir
1. Abra a pasta `standalone`.
2. Clique duas vezes em **Run-MacroRecorder.cmd**.
3. A UI abre com:
   - Abrir/Salvar macro JSON,
   - lista de ações,
   - editor de ação (tipo, label, comment, params JSON),
   - console de execução,
   - botão Executar.

Sem npm install. Sem Python. Sem Node obrigatório.

---

## Opção B (para quem quiser): Electron

Se a máquina tiver Node:

```bash
npm install
npm run desktop
```

---

## Recursos no runtime unificado (mesma macro JSON)

- Variáveis, labels, goto, repeat, if-then-else, stop.
- Waits (tempo, horário, arquivo).
- Mouse/teclado/foco/screenshot/beep/messagebox/execução de programa.
- Ações avançadas no mesmo runtime com fallback best-effort quando necessário.

## Saída de variáveis

- `%TEMP%\\macro-recorder-no-admin\\last-vars.json`
