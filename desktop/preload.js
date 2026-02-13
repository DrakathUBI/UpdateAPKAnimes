import { contextBridge, ipcRenderer } from 'electron';

contextBridge.exposeInMainWorld('macroApp', {
  openFile: () => ipcRenderer.invoke('open-file'),
  saveFile: (payload) => ipcRenderer.invoke('save-file', payload),
  runMacro: (payload) => ipcRenderer.invoke('run-macro', payload),
  platform: process.platform
});
