# Macro Recorder Windows No-Admin (UI completa)

Agora o projeto é **um programa desktop com UI bonita (Electron)** + engine all-in-one no mesmo repositório.

## Instalação (sem Python)

### Requisito único
- Node.js 20+ no Windows.

### Passos
1. Abra o terminal na pasta do projeto.
2. Rode:
   ```bash
   npm install
   ```
3. Para abrir a interface gráfica:
   ```bash
   npm run desktop
   ```

## Uso rápido da UI

- **Abrir**: carrega macro JSON.
- **Salvar**: salva macro JSON.
- **+ Ação**: adiciona ação.
- **Editor de ação**: tipo, label, comentário e params JSON.
- **Executar**: roda a macro e mostra logs no painel da direita.
- Abas visuais: File / Record and Edit / Playback / Help.

## CLI (opcional)

```bash
npm run start -- --play examples/macro_full.json
```

ou

```bash
node src/main.js --play examples/macro_full.json --start-at Inicio
```

## Testes

```bash
npm test
```

## O que está incluso no engine único

- Variáveis, labels, goto, repeat, if-then-else, stop.
- Waits (tempo, horário, arquivo).
- Mouse/teclado/foco/screenshot/beep/messagebox/execução de programa em Windows user-space.
- Actions avançadas no mesmo runtime com fallback best-effort quando necessário.

## Saída de variáveis

- `%TEMP%\\macro-recorder-no-admin\\last-vars.json`
