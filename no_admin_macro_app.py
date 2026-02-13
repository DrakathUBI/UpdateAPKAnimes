#!/usr/bin/env python3
"""Mini Macro Recorder/Player sem exigir privilégios de administrador.

Requisitos:
- Python 3.10+
- pynput
- pyautogui

Observação:
- Em macOS pode exigir permissões de Acessibilidade/Screen Recording do usuário,
  mas não requer executar como administrador.
"""
from __future__ import annotations

import json
import threading
import time
from dataclasses import asdict, dataclass
from pathlib import Path
from tkinter import BOTH, END, LEFT, RIGHT, Button, Frame, Label, Listbox, Tk, filedialog, messagebox

import pyautogui
from pynput import keyboard, mouse


@dataclass
class MacroEvent:
    t: float
    kind: str
    data: dict


class MacroApp:
    def __init__(self, root: Tk) -> None:
        self.root = root
        self.root.title("Macro Recorder (Sem Admin)")
        self.root.geometry("820x480")

        self.events: list[MacroEvent] = []
        self.recording = False
        self.playing = False
        self.start_ts = 0.0

        self.k_listener: keyboard.Listener | None = None
        self.m_listener: mouse.Listener | None = None

        self._build_ui()

    def _build_ui(self) -> None:
        top = Frame(self.root)
        top.pack(fill=BOTH, padx=12, pady=12)

        Button(top, text="Iniciar gravação", command=self.start_recording, width=18).pack(side=LEFT, padx=4)
        Button(top, text="Parar gravação", command=self.stop_recording, width=18).pack(side=LEFT, padx=4)
        Button(top, text="Reproduzir", command=self.play_macro, width=14).pack(side=LEFT, padx=4)
        Button(top, text="Salvar JSON", command=self.save_macro, width=14).pack(side=LEFT, padx=4)
        Button(top, text="Carregar JSON", command=self.load_macro, width=14).pack(side=LEFT, padx=4)
        Button(top, text="Limpar", command=self.clear_macro, width=10).pack(side=LEFT, padx=4)

        self.status = Label(self.root, text="Pronto", anchor="w")
        self.status.pack(fill=BOTH, padx=12)

        self.listbox = Listbox(self.root)
        self.listbox.pack(fill=BOTH, expand=True, padx=12, pady=12)

        note = Label(
            self.root,
            text=(
                "Dica: execute como usuário normal. Não requer Admin. "
                "Se o SO bloquear automação, habilite permissões de acessibilidade do app."
            ),
            fg="#444",
            justify=LEFT,
            anchor="w",
        )
        note.pack(fill=BOTH, padx=12, pady=(0, 10))

    def _set_status(self, msg: str) -> None:
        self.status.config(text=msg)

    def _append_event(self, kind: str, data: dict) -> None:
        if not self.recording:
            return
        event = MacroEvent(t=time.time() - self.start_ts, kind=kind, data=data)
        self.events.append(event)
        self.listbox.insert(END, f"{event.t:7.3f}s | {kind} | {data}")

    def start_recording(self) -> None:
        if self.recording:
            return
        if self.playing:
            messagebox.showwarning("Atenção", "Pare a reprodução antes de gravar.")
            return
        self.recording = True
        self.start_ts = time.time()
        self._set_status("Gravando... (teclado + mouse)")

        self.k_listener = keyboard.Listener(
            on_press=lambda key: self._append_event("key_press", {"key": str(key)}),
            on_release=lambda key: self._append_event("key_release", {"key": str(key)}),
        )
        self.m_listener = mouse.Listener(
            on_move=lambda x, y: self._append_event("mouse_move", {"x": x, "y": y}),
            on_click=lambda x, y, button, pressed: self._append_event(
                "mouse_click", {"x": x, "y": y, "button": str(button), "pressed": pressed}
            ),
            on_scroll=lambda x, y, dx, dy: self._append_event(
                "mouse_scroll", {"x": x, "y": y, "dx": dx, "dy": dy}
            ),
        )
        self.k_listener.start()
        self.m_listener.start()

    def stop_recording(self) -> None:
        if not self.recording:
            return
        self.recording = False
        self._set_status(f"Gravação finalizada ({len(self.events)} eventos)")
        if self.k_listener:
            self.k_listener.stop()
            self.k_listener = None
        if self.m_listener:
            self.m_listener.stop()
            self.m_listener = None

    def _play_thread(self) -> None:
        self.playing = True
        self._set_status("Reproduzindo...")

        prev_t = 0.0
        for event in self.events:
            if not self.playing:
                break
            sleep_for = max(0.0, event.t - prev_t)
            time.sleep(sleep_for)
            prev_t = event.t

            kind = event.kind
            data = event.data

            try:
                if kind == "mouse_move":
                    pyautogui.moveTo(data["x"], data["y"])
                elif kind == "mouse_click":
                    btn = "left"
                    if "right" in data["button"]:
                        btn = "right"
                    elif "middle" in data["button"]:
                        btn = "middle"
                    if data["pressed"]:
                        pyautogui.mouseDown(data["x"], data["y"], button=btn)
                    else:
                        pyautogui.mouseUp(data["x"], data["y"], button=btn)
                elif kind == "mouse_scroll":
                    pyautogui.scroll(int(data["dy"] * 120))
                elif kind in {"key_press", "key_release"}:
                    key_text = data["key"]
                    if key_text.startswith("'") and key_text.endswith("'"):
                        key_name = key_text.strip("'")
                    elif key_text.startswith("Key."):
                        key_name = key_text.split(".", 1)[1]
                    else:
                        key_name = key_text
                    if kind == "key_press":
                        pyautogui.keyDown(key_name)
                    else:
                        pyautogui.keyUp(key_name)
            except Exception:
                # Mantém execução robusta mesmo com evento não suportado
                pass

        self.playing = False
        self._set_status("Reprodução finalizada")

    def play_macro(self) -> None:
        if self.recording:
            messagebox.showwarning("Atenção", "Pare a gravação antes de reproduzir.")
            return
        if not self.events:
            messagebox.showinfo("Info", "Nenhum evento para reproduzir.")
            return
        if self.playing:
            self.playing = False
            self._set_status("Parando reprodução...")
            return
        threading.Thread(target=self._play_thread, daemon=True).start()

    def save_macro(self) -> None:
        file = filedialog.asksaveasfilename(
            title="Salvar macro",
            defaultextension=".json",
            filetypes=[("JSON", "*.json")],
        )
        if not file:
            return
        Path(file).write_text(json.dumps([asdict(e) for e in self.events], ensure_ascii=False, indent=2), encoding="utf-8")
        self._set_status(f"Macro salva em {file}")

    def load_macro(self) -> None:
        file = filedialog.askopenfilename(title="Abrir macro", filetypes=[("JSON", "*.json")])
        if not file:
            return
        raw = json.loads(Path(file).read_text(encoding="utf-8"))
        self.events = [MacroEvent(**ev) for ev in raw]
        self.listbox.delete(0, END)
        for event in self.events:
            self.listbox.insert(END, f"{event.t:7.3f}s | {event.kind} | {event.data}")
        self._set_status(f"Macro carregada ({len(self.events)} eventos)")

    def clear_macro(self) -> None:
        self.events.clear()
        self.listbox.delete(0, END)
        self._set_status("Eventos limpos")


if __name__ == "__main__":
    pyautogui.FAILSAFE = True
    app = Tk()
    MacroApp(app)
    app.mainloop()
