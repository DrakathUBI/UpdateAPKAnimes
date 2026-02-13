import { app, BrowserWindow, dialog, ipcMain } from 'electron';
import path from 'node:path';
import fs from 'node:fs';
import { fileURLToPath } from 'node:url';
import { MacroEngine } from '../src/engine.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function createWindow() {
  const win = new BrowserWindow({
    width: 1440,
    height: 920,
    title: 'Macro Recorder Pro UI',
    autoHideMenuBar: true,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false
    }
  });
  win.loadFile(path.join(__dirname, '../ui/index.html'));
}

app.whenReady().then(createWindow);

ipcMain.handle('open-file', async () => {
  const { canceled, filePaths } = await dialog.showOpenDialog({
    filters: [{ name: 'Macro JSON', extensions: ['json'] }],
    properties: ['openFile']
  });
  if (canceled || !filePaths[0]) return { canceled: true };
  const file = filePaths[0];
  return { canceled: false, file, content: fs.readFileSync(file, 'utf8') };
});

ipcMain.handle('save-file', async (_, payload) => {
  const initial = payload?.file || 'macro.json';
  const { canceled, filePath } = await dialog.showSaveDialog({
    defaultPath: initial,
    filters: [{ name: 'Macro JSON', extensions: ['json'] }]
  });
  if (canceled || !filePath) return { canceled: true };
  fs.writeFileSync(filePath, JSON.stringify(payload.data, null, 2));
  return { canceled: false, file: filePath };
});

ipcMain.handle('run-macro', async (_, payload) => {
  const logs = [];
  const engine = new MacroEngine({ silent: false, logger: (m) => logs.push(String(m)) });
  const result = await engine.runScript(payload.macro, { startAt: payload.startAt || '' });
  return { ok: true, result, logs };
});
