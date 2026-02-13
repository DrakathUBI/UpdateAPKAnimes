import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import process from 'node:process';
import { spawn } from 'node:child_process';

const RESERVED_LABELS = new Set(['Start', 'End', 'Next']);

class VariableStore {
  constructor() { this.map = new Map(); }
  set(name, value) { this.map.set(String(name), value); }
  get(name, fallback = '') { return this.map.has(String(name)) ? this.map.get(String(name)) : fallback; }
  all() { return Object.fromEntries(this.map.entries()); }
}

class WindowsAdapter {
  constructor({ silent = false }) {
    this.silent = silent;
    this.isWindows = process.platform === 'win32';
  }

  warn(msg) { console.warn(`[WARN] ${msg}`); }

  async runPowerShell(script, timeoutMs = 15000) {
    if (!this.isWindows) return { ok: false, out: '', err: 'Not running on Windows' };
    return new Promise((resolve) => {
      const ps = spawn('powershell.exe', ['-NoProfile', '-ExecutionPolicy', 'Bypass', '-Command', script], { stdio: ['ignore', 'pipe', 'pipe'] });
      let out = '';
      let err = '';
      const t = setTimeout(() => {
        ps.kill('SIGTERM');
        resolve({ ok: false, out, err: 'PowerShell timeout' });
      }, timeoutMs);
      ps.stdout.on('data', d => { out += String(d); });
      ps.stderr.on('data', d => { err += String(d); });
      ps.on('close', code => {
        clearTimeout(t);
        resolve({ ok: code === 0, out: out.trim(), err: err.trim() });
      });
    });
  }

  async textOutput(text, mode = 'typing') {
    if (!this.isWindows) return this.warn('text_output ativo somente no Windows.');
    const escaped = String(text).replace(/'/g, "''");
    if (mode === 'clipboard') {
      await this.runPowerShell(`Set-Clipboard -Value '${escaped}'; Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait('^v')`);
      return;
    }
    await this.runPowerShell(`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait('${escaped}')`);
  }

  async hotkey(keys) {
    if (!this.isWindows) return this.warn('hotkey ativa somente no Windows.');
    const sendKeys = String(keys).replace(/CTRL/gi, '^').replace(/ALT/gi, '%').replace(/SHIFT/gi, '+').replace(/ENTER/gi, '{ENTER}').replace(/WIN/gi, '^{ESC}');
    await this.runPowerShell(`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait('${sendKeys}')`);
  }

  async mouseMove(x, y) {
    if (!this.isWindows) return this.warn('mouse_move ativo somente no Windows.');
    const xi = Number(x) || 0;
    const yi = Number(y) || 0;
    await this.runPowerShell(`Add-Type @'
using System.Runtime.InteropServices;
public class U{ [DllImport("user32.dll")] public static extern bool SetCursorPos(int X, int Y); }
'@; [U]::SetCursorPos(${xi},${yi}) | Out-Null`);
  }

  async mouseClick(button = 'left') {
    if (!this.isWindows) return this.warn('mouse_click ativo somente no Windows.');
    const map = { left: [2, 4], right: [8, 16], middle: [32, 64] };
    const [down, up] = map[String(button).toLowerCase()] ?? map.left;
    await this.runPowerShell(`Add-Type @'
using System.Runtime.InteropServices;
public class U{ [DllImport("user32.dll")] public static extern void mouse_event(int dwFlags,int dx,int dy,int dwData,int dwExtraInfo); }
'@; [U]::mouse_event(${down},0,0,0,0); Start-Sleep -Milliseconds 20; [U]::mouse_event(${up},0,0,0,0)`);
  }

  async scroll(delta = 120) {
    if (!this.isWindows) return this.warn('mouse_scroll ativo somente no Windows.');
    await this.runPowerShell(`Add-Type @'
using System.Runtime.InteropServices;
public class U{ [DllImport("user32.dll")] public static extern void mouse_event(int dwFlags,int dx,int dy,int dwData,int dwExtraInfo); }
'@; [U]::mouse_event(2048,0,0,${Number(delta) || 120},0)`);
  }

  async focusWindow(title) {
    if (!this.isWindows) return this.warn('window_focus ativo somente no Windows.');
    const t = String(title).replace(/'/g, "''");
    await this.runPowerShell(`$ws=New-Object -ComObject WScript.Shell; $null=$ws.AppActivate('${t}')`);
  }

  async messageBox(text, buttons = 'OK') {
    if (!this.isWindows) return 'OK';
    const t = String(text).replace(/'/g, "''");
    const b = String(buttons).toUpperCase();
    const buttonExpr = b.includes('YESNO') ? 'YesNo' : b.includes('OKCANCEL') ? 'OKCancel' : 'OK';
    const r = await this.runPowerShell(`Add-Type -AssemblyName System.Windows.Forms; $x=[System.Windows.Forms.MessageBox]::Show('${t}','Macro Recorder','${buttonExpr}'); Write-Output $x`);
    return r.out || 'OK';
  }

  async screenshot(region, outPath) {
    if (!this.isWindows) return false;
    const [x, y, w, h] = region.map(n => Number(n) || 0);
    const p = outPath.replace(/\\/g, '/').replace(/'/g, "''");
    const s = `Add-Type -AssemblyName System.Drawing; $bmp=New-Object System.Drawing.Bitmap ${w},${h}; $g=[System.Drawing.Graphics]::FromImage($bmp); $g.CopyFromScreen(${x},${y},0,0,$bmp.Size); $bmp.Save('${p}',[System.Drawing.Imaging.ImageFormat]::Png); $g.Dispose(); $bmp.Dispose();`;
    const r = await this.runPowerShell(s);
    return r.ok;
  }

  async beep() {
    if (!this.isWindows) return;
    await this.runPowerShell('[console]::beep(1200,130)');
  }

  async postPlayback(action) {
    if (!this.isWindows) return;
    const cmd = {
      shutdown: 'shutdown /s /t 0', restart: 'shutdown /r /t 0', logout: 'shutdown /l',
      lock: 'rundll32.exe user32.dll,LockWorkStation', standby: 'rundll32.exe powrprof.dll,SetSuspendState 0,1,0'
    }[action];
    if (cmd) await this.runPowerShell(cmd);
  }
}

class MacroEngine {
  constructor({ silent = false, logger = console.log } = {}) {
    this.silent = silent;
    this.logger = logger;
    this.vars = new VariableStore();
    this.adapter = new WindowsAdapter({ silent });
    this.stopRequested = false;
  }

  log(msg) { if (!this.silent) this.logger(msg); }
  warn(msg) { this.logger(`[WARN] ${msg}`); }

  i(v) {
    if (typeof v !== 'string') return v;
    return v.replace(/\$\{([A-Za-z0-9_]+)\}/g, (_, n) => String(this.vars.get(n, '')));
  }

  labelIndex(actions) {
    const labels = new Map();
    actions.forEach((a, idx) => {
      if (!a.label) return;
      if (RESERVED_LABELS.has(a.label)) throw new Error(`Label reservada: ${a.label}`);
      if (labels.has(a.label)) throw new Error(`Label duplicada: ${a.label}`);
      labels.set(a.label, idx);
    });
    return labels;
  }

  jump(label, labels) {
    if (!label) return null;
    if (!labels.has(label)) throw new Error(`Label inexistente: ${label}`);
    return { nextPc: labels.get(label) };
  }

  async runScript(script, { startAt = '' } = {}) {
    const actions = script.actions ?? [];
    const labels = this.labelIndex(actions);
    let pc = startAt ? (labels.get(startAt) ?? 0) : 0;
    this.log(`Compat mode: ${script.compatibility ?? 'v4_v5_unified'}`);

    while (pc < actions.length && !this.stopRequested) {
      const a = actions[pc];
      if (a.enabled === false) { pc++; continue; }
      const jump = await this.execute(a, { labels, pc });
      pc = jump?.nextPc ?? pc + 1;
    }

    const post = script.playback_profile?.post_playback_action;
    if (!this.stopRequested && post && post !== 'none') {
      const seconds = Number(script.playback_profile?.post_countdown_s ?? 5);
      this.log(`Post-playback '${post}' em ${seconds}s (CTRL+C para cancelar).`);
      await sleep(seconds * 1000);
      await this.adapter.postPlayback(post);
    }

    const outDir = path.join(os.tmpdir(), 'macro-recorder-no-admin');
    fs.mkdirSync(outDir, { recursive: true });
    const outFile = path.join(outDir, 'last-vars.json');
    fs.writeFileSync(outFile, JSON.stringify(this.vars.all(), null, 2));

    return { stopped: this.stopRequested, vars: this.vars.all(), outFile };
  }

  async execute(a, ctx) {
    const p = a.params ?? {};
    switch (a.type) {
      case 'set_variable': this.vars.set(p.name, this.i(p.value ?? '')); return;
      case 'create_variable': this.vars.set(p.name, p.default ?? ''); return;
      case 'text_output': return this.adapter.textOutput(this.i(p.text ?? ''), p.mode ?? 'typing');
      case 'keypress': return this.adapter.hotkey(this.i(p.key ?? '{ENTER}'));
      case 'hotkey': return this.adapter.hotkey(this.i(p.keys ?? ''));
      case 'mouse_move': return this.adapter.mouseMove(this.i(`${p.x ?? 0}`), this.i(`${p.y ?? 0}`));
      case 'mouse_click':
        if (p.x != null && p.y != null) await this.adapter.mouseMove(this.i(`${p.x}`), this.i(`${p.y}`));
        return this.adapter.mouseClick(p.button ?? 'left');
      case 'mouse_scroll': return this.adapter.scroll(this.i(`${p.delta ?? 120}`));
      case 'window_focus': return this.adapter.focusWindow(this.i(p.window ?? p.title ?? ''));
      case 'execute_program': {
        const program = this.i(p.program ?? '');
        const args = (p.args ?? []).map(v => this.i(String(v)));
        if (program) spawn(program, args, { detached: true, stdio: 'ignore' }).unref();
        return;
      }
      case 'wait_time': {
        const min = Number(this.i(`${p.minMs ?? p.ms ?? 0}`));
        const max = Number(this.i(`${p.maxMs ?? min}`));
        await sleep(max > min ? Math.floor(Math.random() * (max - min + 1)) + min : min);
        return;
      }
      case 'wait_until_time': await sleep(Math.max(0, new Date(this.i(p.isoTime)).getTime() - Date.now())); return;
      case 'wait_for_file_event': return this.waitFile(p, ctx);
      case 'wait_for_hotkey': case 'wait_for_text_input': case 'wait_for_pixel_color': case 'wait_for_desktop_change':
      case 'smart_click': case 'find_image': case 'find_text_ocr': case 'capture_text_ocr': case 'capture_barcode_qr':
      case 'ai_object_search': case 'phrase_insertion': case 'embed_macro': return this.bestEffort(a.type, p, ctx);
      case 'capture_bitmap': {
        const out = this.i(p.outputPath ?? path.join(os.tmpdir(), `macro_capture_${Date.now()}.png`));
        const ok = await this.adapter.screenshot(p.region ?? [0, 0, 300, 200], out);
        if (p.outputVar) this.vars.set(p.outputVar, ok ? out : '');
        return;
      }
      case 'scrape_webpage': {
        const resp = await fetch(this.i(p.url ?? ''));
        const html = await resp.text();
        let value = '';
        if (p.regex) {
          const m = new RegExp(this.i(p.regex), 'm').exec(html);
          value = m?.[1] ?? m?.[0] ?? '';
        }
        this.vars.set(p.outputVar ?? 'Scrape', value);
        return;
      }
      case 'save_variable': {
        const value = this.vars.get(p.name, '');
        if (p.destination === 'clipboard') return this.adapter.textOutput(String(value), 'clipboard');
        const target = this.i(p.path);
        fs.mkdirSync(path.dirname(target), { recursive: true });
        if (p.append) fs.appendFileSync(target, String(value));
        else fs.writeFileSync(target, String(value));
        return;
      }
      case 'data_list': {
        const key = `__datalist_idx_${p.name}`;
        const idx = Number(this.vars.get(key, 0));
        const arr = p.file ? fs.readFileSync(this.i(p.file), 'utf8').split(/\r?\n/).filter(Boolean) : (p.items ?? []);
        this.vars.set(p.target ?? p.name, arr[idx] ?? '');
        this.vars.set(key, idx + 1);
        this.vars.set(`__datalist_end_${p.name}`, idx + 1 >= arr.length);
        return;
      }
      case 'calculation': {
        const expr = this.i(p.expression ?? '0').replace(/\^/g, '**').replace(/\bpi\b/gi, `${Math.PI}`).replace(/\be\b/g, `${Math.E}`);
        this.vars.set(p.target, Function(`"use strict"; return (${expr});`)());
        return;
      }
      case 'if_then_else': return this.jump(compare(this.i(`${p.left ?? ''}`), this.i(`${p.right ?? ''}`), p.operator ?? 'equals') ? p.thenLabel : p.elseLabel, ctx.labels);
      case 'goto': return this.jump(p.label, ctx.labels);
      case 'repeat': {
        const key = `__repeat_${ctx.pc}`;
        const c = Number(this.vars.get(key, 0)) + 1;
        this.vars.set(key, c);
        const byCount = p.count != null && c <= Number(p.count);
        const byList = p.untilLastDataItem && !this.vars.get(`__datalist_end_${p.dataListName}`, false);
        if (byCount || byList) return this.jump(p.label, ctx.labels);
        return this.jump(p.afterLabel, ctx.labels);
      }
      case 'show_notification': this.log(`[NOTIFY] ${this.i(p.text ?? '')}`); return;
      case 'show_message_box': {
        const result = await this.adapter.messageBox(this.i(p.text ?? ''), p.buttons?.join('') ?? 'OK');
        if (p.labelByResult?.[result]) return this.jump(p.labelByResult[result], ctx.labels);
        return;
      }
      case 'beep': return this.adapter.beep();
      case 'stop': this.stopRequested = true; return;
      default: this.warn(`Ação não implementada: ${a.type}`); return;
    }
  }

  async waitFile(p, ctx) {
    const file = this.i(p.path);
    const initial = statLite(file);
    const timeoutMs = Number(p.timeoutMs ?? 60000);
    const started = Date.now();
    while (Date.now() - started < timeoutMs) {
      const cur = statLite(file);
      if (p.mode === 'exists' && cur.exists) return;
      if (p.mode === 'changed' && initial.mtimeMs !== null && cur.mtimeMs !== initial.mtimeMs) return;
      if (p.mode === 'attributes_changed' && JSON.stringify(initial) !== JSON.stringify(cur)) return;
      await sleep(200);
    }
    if (p.onTimeout === 'goto') return this.jump(p.timeoutLabel, ctx.labels);
    if (p.onTimeout === 'abort') throw new Error('Timeout no wait_for_file_event');
  }

  async bestEffort(type, p, ctx) {
    this.warn(`${type} executado em modo best-effort.`);
    if (p.outputVarX) this.vars.set(p.outputVarX, 0);
    if (p.outputVarY) this.vars.set(p.outputVarY, 0);
    if (p.outputVar) this.vars.set(p.outputVar, '');
    if (p.timeoutMs) await sleep(Math.min(Number(p.timeoutMs), 250));
    if (p.onTimeout === 'goto' && p.timeoutLabel) return this.jump(p.timeoutLabel, ctx.labels);
    if (p.onTimeout === 'abort') throw new Error(`Timeout em ${type}`);
    if (p.onTimeout === 'repeat_from_scratch') return { nextPc: 0 };
  }
}

function compare(left, right, op) {
  switch (op) {
    case 'equals': return String(left) === String(right);
    case 'contains': return String(left).includes(String(right));
    case 'regex': return new RegExp(String(right)).test(String(left));
    case 'gt': return Number(left) > Number(right);
    case 'lt': return Number(left) < Number(right);
    case 'gte': return Number(left) >= Number(right);
    case 'lte': return Number(left) <= Number(right);
    default: return false;
  }
}

function statLite(filePath) {
  try {
    const s = fs.statSync(filePath);
    return { exists: true, mtimeMs: s.mtimeMs, size: s.size };
  } catch {
    return { exists: false, mtimeMs: null, size: null };
  }
}

function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }

function parseArgs(argv) {
  const out = { play: '', silent: false, startAt: '' };
  for (let i = 2; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--play') out.play = argv[++i];
    else if (a === '--silent') out.silent = true;
    else if (a === '--start-at') out.startAt = argv[++i];
    else if (a === '--example') out.play = 'examples/macro_full.json';
  }
  return out;
}

export { MacroEngine, VariableStore, parseArgs, WindowsAdapter };
