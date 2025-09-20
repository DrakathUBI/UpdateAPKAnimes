'use strict';

Java.perform(function () {
  const Activity = Java.use('android.app.Activity');
  const View = Java.use('android.view.View');
  const ViewGroup = Java.use('android.view.ViewGroup');
  const WindowManagerGlobal = Java.use('android.view.WindowManagerGlobal');
  const KeyEvent = Java.use('android.view.KeyEvent');

  // ----- Configuráveis -----
  const KEYWORDS = [
    'Dublado','Legendado','Episódio','Episodio','Animes Brasil',
    'Pular abertura','Pular','Reproduzir','Pausar','Pause','Play',
    'Voltar','Navegar para cima'
  ];
  // px de margem pro “perto da borda”
  const EDGE = 180;                 // ajuste se necessário
  const SMALL = 160;                // views menor que isso: suspeitas
  const THIN = 26;                  // barras/linhas finas
  const DOUBLE_MS = 320;            // janela do duplo-toque ↓
  // -------------------------

  const touched = [];  // {v, vis, alpha, clickable, focusable}
  let lastDown = 0;

  function isVideoView(v) {
    const name = v.getClass().getName().toString();
    return name.endsWith('SurfaceView') || name.endsWith('TextureView');
  }
  function hasVideoDescendant(v) {
    try {
      if (isVideoView(v)) return true;
      if (!Java.cast(v, ViewGroup)) return false;
      const vg = Java.cast(v, ViewGroup);
      for (let i = 0; i < vg.getChildCount(); i++) {
        const c = vg.getChildAt(i);
        if (hasVideoDescendant(c)) return true;
      }
    } catch (e) {}
    return false;
  }

  function getTextIfAny(v) {
    try {
      const tv = Java.cast(v, Java.use('android.widget.TextView'));
      const s = tv.getText();
      return s ? s.toString() : '';
    } catch (_) { return ''; }
  }

  function bounds(v) {
    const r = Java.use('android.graphics.Rect').$new();
    v.getGlobalVisibleRect(r);
    return {x: r.left.value, y: r.top.value, w: r.right.value - r.left.value, h: r.bottom.value - r.top.value};
  }

  function likelyOverlay(v) {
    if (isVideoView(v)) return false;

    // 1) Texto com palavras-chave
    const txt = getTextIfAny(v);
    for (const k of KEYWORDS) if (txt.indexOf(k) >= 0) return true;

    // 2) Heurística geométrica
    const b = bounds(v);
    const nearEdge = (b.x < EDGE) || (b.y < EDGE) ||
                     (Math.abs(screenW() - (b.x + b.w)) < EDGE) ||
                     (Math.abs(screenH() - (b.y + b.h)) < EDGE);
    const smallish = (b.w <= SMALL) || (b.h <= SMALL);
    const thin = (b.h <= THIN) || (b.w <= THIN);

    // 3) Classes comuns de overlay
    const name = v.getClass().getName().toString();
    const isImgOrTxt = name.indexOf('Image') >= 0 || name.indexOf('Text') >= 0 || name.indexOf('Chip') >= 0 || name.indexOf('Button') >= 0;

    // 4) Containers sem vídeo dentro (menus, overlays)
    let containerNoVideo = false;
    try {
      if (Java.cast(v, ViewGroup) && !hasVideoDescendant(v)) {
        // só marque como overlay se não ocupar a tela inteira
        const b2 = bounds(v);
        const notFull = (b2.w < screenW()*0.95) || (b2.h < screenH()*0.95);
        containerNoVideo = notFull;
      }
    } catch (e) {}

    return isImgOrTxt || thin || (nearEdge && smallish) || containerNoVideo;
  }

  function hideView(v) {
    try {
      const prev = {
        v: v,
        vis: v.getVisibility(),
        alpha: v.getAlpha(),
        clickable: v.isClickable(),
        focusable: v.isFocusable()
      };
      touched.push(prev);
      v.setVisibility(View.GONE.value);
      v.setAlpha(0.0);
      v.setClickable(false);
      v.setFocusable(false);
    } catch (e) {}
  }

  function restoreAll() {
    runUI(function () {
      let n = 0;
      while (touched.length) {
        const t = touched.pop();
        try {
          t.v.setVisibility(t.vis);
          t.v.setAlpha(t.alpha);
          t.v.setClickable(t.clickable);
          t.v.setFocusable(t.focusable);
          n++;
        } catch (_) {}
      }
      send(`[SWEEP] restore: ${n}`);
    });
  }

  function screenW() {
    try {
      const ctx = currentActivity();
      return ctx.getResources().getDisplayMetrics().widthPixels.value;
    } catch (_) { return 1080; }
  }
  function screenH() {
    try {
      const ctx = currentActivity();
      return ctx.getResources().getDisplayMetrics().heightPixels.value;
    } catch (_) { return 1920; }
  }

  function walk(root, fn) {
    const stack = [root];
    while (stack.length) {
      const v = stack.pop();
      try {
        fn(v);
        if (Java.cast(v, ViewGroup)) {
          const vg = Java.cast(v, ViewGroup);
          for (let i = 0; i < vg.getChildCount(); i++) stack.push(vg.getChildAt(i));
        }
      } catch (_) {}
    }
  }

  function rootsViaAT() {
    const out = [];
    try {
      const AT = Java.use('android.app.ActivityThread').currentActivityThread();
      const acts = AT.mActivities.value; // ArrayMap
      const it = acts.keySet().iterator();
      while (it.hasNext()) {
        const k = it.next();
        const rec = acts.get(k);
        const act = rec.activity.value;
        if (!act) continue;
        const root = act.getWindow().getDecorView().getRootView();
        out.push(root);
      }
    } catch (_) {}
    return out;
  }

  function rootsViaWM() {
    const out = [];
    try {
      const wmg = WindowManagerGlobal.getInstance();
      const names = wmg.getViewRootNames();
      for (let i = 0; i < names.length; i++) {
        const n = names[i];
        try { out.push(wmg.getRootView(n)); } catch (_) {}
      }
    } catch (_) {}
    return out;
  }

  function allRoots() {
    const a = rootsViaAT();
    const b = rootsViaWM();
    return a.concat(b);
  }

  function runUI(fn) {
    try {
      const act = currentActivity();
      if (act) {
        act.runOnUiThread(Java.registerClass({
          name: 'com.frida.RunUi' + Date.now(),
          superClass: Java.use('java.lang.Runnable'),
          methods: {
            run: function () { try { fn(); } catch (e) {} }
          }
        }).$new());
        return;
      }
    } catch (_) {}
    // fallback: roda direto
    try { fn(); } catch (_) {}
  }

  function currentActivity() {
    try {
      const AT = Java.use('android.app.ActivityThread').currentActivityThread();
      const acts = AT.mActivities.value;
      const it = acts.keySet().iterator();
      while (it.hasNext()) {
        const k = it.next();
        const rec = acts.get(k);
        const act = rec.activity.value;
        if (act && !act.isFinishing()) return act;
      }
    } catch (_) {}
    return null;
  }

  function hideNow() {
    runUI(function () {
      let total = 0, hidden = 0;
      const roots = allRoots();
      roots.forEach(r => {
        walk(r, function (v) {
          total++;
          try {
            if (v.getVisibility() !== View.VISIBLE.value) return;
            if (likelyOverlay(v)) hideView(v), hidden++;
          } catch (_) {}
        });
      });
      send(`[SWEEP] hide: total=${total} ocultados=${hidden}`);
    });
  }

  function scanOnly() {
    let total = 0, cand = 0;
    const roots = allRoots();
    roots.forEach(r => {
      walk(r, function (v) {
        total++;
        try { if (likelyOverlay(v)) cand++; } catch (_) {}
      });
    });
    send(`[SWEEP] scan: total=${total} candidatos=${cand}`);
  }

  // Hotkeys no nível do Activity
  try {
    const orig = Activity.dispatchKeyEvent.overload('android.view.KeyEvent');
    Activity.dispatchKeyEvent.overload('android.view.KeyEvent').implementation = function (ev) {
      const isDown = ev.getAction() === KeyEvent.ACTION_DOWN.value;
      const code = ev.getKeyCode();
      const now = Date.now();
      if (isDown && code === KeyEvent.KEYCODE_DPAD_DOWN.value) {
        if (now - lastDown <= DOUBLE_MS) { hideNow(); return true; }
        lastDown = now;
      }
      if (isDown && (code === KeyEvent.KEYCODE_DPAD_UP.value || code === KeyEvent.KEYCODE_6.value)) {
        restoreAll(); return true;
      }
      if (isDown && code === KeyEvent.KEYCODE_5.value) {
        if (touched.length) restoreAll(); else hideNow();
        return true;
      }
      return orig.call(this, ev);
    };
  } catch (_) {}

  send('[SWEEP] ativo. ↓↓/5 = esconder, ↑↑/6 = restaurar. RPC: hide/restore/toggle/scan');

  rpc.exports = {
    hide: hideNow,
    restore: restoreAll,
    toggle: function () { if (touched.length) restoreAll(); else hideNow(); },
    scan: scanOnly
  };
});
