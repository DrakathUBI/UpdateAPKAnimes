#!/usr/bin/env bash
set -euo pipefail

python3 -m pip install --user -r requirements.txt
python3 -m PyInstaller --noconfirm --onefile --windowed --name no_admin_macro_app no_admin_macro_app.py

echo "Build concluído. Verifique a pasta dist/."
