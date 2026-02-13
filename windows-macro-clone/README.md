# MacroClone (Windows, sem admin)

Este projeto entrega um **executável .exe** baseado em .NET 8 que roda macros em JSON via linha de comando.

## O que já funciona
- CLI: `--play`, `--silent`, `--start-at`.
- Engine de fluxo: `Goto`, `IfThenElse`, variáveis, cálculo, `WaitTime`, `SaveVariable`, `ExecuteProgram`, `Beep`.
- Enum com todas as ações da especificação para manter compatibilidade de schema.
- Modo seguro sem admin: ações de integração profunda com UI/SO geram aviso e não quebram execução.

## Build (gerar .exe)
```bash
cd windows-macro-clone
dotnet publish -c Release -r win-x64 --self-contained true
```
Saída esperada:
`bin/Release/net8.0/win-x64/publish/MacroClone.exe`

## Executar
```bash
MacroClone.exe --play macro.sample.json
```

## Importante
Sem privilégios de administrador, hooks globais de entrada, manipulação de janelas elevadas, OCR de apps protegidos e automação visual avançada podem ter restrições do Windows.
