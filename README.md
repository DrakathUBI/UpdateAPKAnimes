# Macro Recorder para Windows (sem Python no PC do usuário)

Você está certo: **não vamos versionar binários no Git** (compatibilidade e política de repositório).

## O que ficou no repositório
- Código-fonte Windows: `windows_macro_recorder.go`
- Launcher Windows: `INICIAR_WINDOWS.bat`
- Script de build Windows: `BUILD_WINDOWS.bat`

## Como gerar o `.exe` (para distribuição)
> Rodar em máquina Windows com Go instalado (somente para quem empacota).

1. Execute `BUILD_WINDOWS.bat`
2. O arquivo será criado em: `dist\MacroRecorderSemAdmin.exe`
3. Entregue esse `.exe` para o usuário final (ele não precisa de Python).

## Uso no Windows (usuário final)
1. Dê duplo clique em `MacroRecorderSemAdmin.exe`
2. Opção `1`: grava macro de mouse (movimento + clique esquerdo)
3. Opção `2`: reproduz macro
4. `F9` para parar gravação

## Observações
- Não precisa rodar como Administrador.
- Se política corporativa bloquear automação, liberar o executável na whitelist da empresa.
