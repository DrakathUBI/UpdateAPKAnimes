const ACTION_TYPES = [
  'set_variable','create_variable','text_output','keypress','hotkey','mouse_move','mouse_click','mouse_scroll',
  'smart_click','window_focus','execute_program','embed_macro','wait_time','wait_until_time','wait_for_file_event',
  'wait_for_hotkey','wait_for_text_input','wait_for_pixel_color','wait_for_desktop_change','find_image','find_text_ocr',
  'capture_bitmap','capture_text_ocr','capture_barcode_qr','scrape_webpage','save_variable','data_list','calculation',
  'if_then_else','goto','repeat','show_notification','show_message_box','beep','phrase_insertion','ai_object_search','stop'
];

let currentFile = '';
let selected = -1;
let macro = {
  name: 'macro-ui',
  compatibility: 'v4_v5_unified',
  playback_profile: { speed: 1, repetitions: 1, filters: [], mouse_path_mode: 'natural', post_playback_action: 'none', post_countdown_s: 5 },
  actions: []
};

const rows = document.getElementById('actionRows');
const fType = document.getElementById('fType');
const fLabel = document.getElementById('fLabel');
const fComment = document.getElementById('fComment');
const fParams = document.getElementById('fParams');
const logView = document.getElementById('logView');

ACTION_TYPES.forEach(t => {
  const o = document.createElement('option');
  o.value = t; o.textContent = t; fType.appendChild(o);
});

function render() {
  rows.innerHTML = '';
  macro.actions.forEach((a, i) => {
    const tr = document.createElement('tr');
    if (i === selected) tr.classList.add('sel');
    tr.innerHTML = `<td>${i + 1}</td><td>${a.type || ''}</td><td>${a.label || ''}</td><td>${a.comment || ''}</td>`;
    tr.onclick = () => { selected = i; fillEditor(); render(); };
    rows.appendChild(tr);
  });
}

function fillEditor() {
  const a = macro.actions[selected];
  if (!a) return;
  fType.value = a.type || ACTION_TYPES[0];
  fLabel.value = a.label || '';
  fComment.value = a.comment || '';
  fParams.value = JSON.stringify(a.params || {}, null, 2);
}

function appendLog(line) {
  logView.textContent += `${line}\n`;
  logView.scrollTop = logView.scrollHeight;
}

document.getElementById('btnAdd').onclick = () => {
  macro.actions.push({ type: ACTION_TYPES[0], params: {} });
  selected = macro.actions.length - 1;
  fillEditor();
  render();
};

document.getElementById('btnApply').onclick = () => {
  if (selected < 0) return;
  try {
    macro.actions[selected] = {
      ...macro.actions[selected],
      type: fType.value,
      label: fLabel.value || undefined,
      comment: fComment.value || undefined,
      params: JSON.parse(fParams.value || '{}')
    };
    appendLog('Ação atualizada.');
    render();
  } catch (e) {
    appendLog(`Erro JSON params: ${e.message}`);
  }
};

document.getElementById('btnDelete').onclick = () => {
  if (selected < 0) return;
  macro.actions.splice(selected, 1);
  selected = -1;
  render();
};

document.getElementById('btnOpen').onclick = async () => {
  const r = await window.macroApp.openFile();
  if (r.canceled) return;
  currentFile = r.file;
  macro = JSON.parse(r.content);
  selected = -1;
  appendLog(`Arquivo aberto: ${currentFile}`);
  render();
};

document.getElementById('btnSave').onclick = async () => {
  const r = await window.macroApp.saveFile({ file: currentFile, data: macro });
  if (r.canceled) return;
  currentFile = r.file;
  appendLog(`Arquivo salvo: ${currentFile}`);
};

document.getElementById('btnRun').onclick = async () => {
  appendLog('Executando macro...');
  try {
    const startAt = document.getElementById('startAt').value;
    const r = await window.macroApp.runMacro({ macro, startAt });
    r.logs.forEach(appendLog);
    appendLog(`Concluído. Vars em: ${r.result.outFile}`);
  } catch (e) {
    appendLog(`Falha: ${e.message}`);
  }
};

render();
