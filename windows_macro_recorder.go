package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type Event struct {
	T    int64             `json:"t_ms"`
	Kind string            `json:"kind"`
	Data map[string]string `json:"data"`
}

type Action struct {
	Type    string            `json:"type"`
	Label   string            `json:"label,omitempty"`
	Comment string            `json:"comment,omitempty"`
	Enabled bool              `json:"enabled,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
}

type MacroScript struct {
	Actions []Action `json:"actions"`
}

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procGetAsyncKey  = user32.NewProc("GetAsyncKeyState")
	procGetCursorPos = user32.NewProc("GetCursorPos")
	procSetCursorPos = user32.NewProc("SetCursorPos")
	procMouseEvent   = user32.NewProc("mouse_event")
	procKeybdEvent   = user32.NewProc("keybd_event")
	leftDown         = uintptr(0x0002)
	leftUp           = uintptr(0x0004)
	rightDown        = uintptr(0x0008)
	rightUp          = uintptr(0x0010)
	middleDown       = uintptr(0x0020)
	middleUp         = uintptr(0x0040)
	wheelEvent       = uintptr(0x0800)
	keyUpFlag        = uintptr(0x0002)
)

type point struct {
	X int32
	Y int32
}

type Runner struct {
	script      MacroScript
	vars        map[string]string
	labels      map[string]int
	repeatCount map[int]int
}

func getAsyncKeyState(vk int) uint16 {
	r, _, _ := procGetAsyncKey.Call(uintptr(vk))
	return uint16(r)
}

func keyPressed(vk int) bool {
	return (getAsyncKeyState(vk) & uint16(0x8000)) != 0
}

func getCursorPos() (int32, int32) {
	var p point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	return p.X, p.Y
}

func setCursorPos(x, y int32) {
	procSetCursorPos.Call(uintptr(x), uintptr(y))
}

func mouseClick(button string, down bool) {
	var flag uintptr
	switch strings.ToLower(button) {
	case "right":
		if down {
			flag = rightDown
		} else {
			flag = rightUp
		}
	case "middle":
		if down {
			flag = middleDown
		} else {
			flag = middleUp
		}
	default:
		if down {
			flag = leftDown
		} else {
			flag = leftUp
		}
	}
	procMouseEvent.Call(flag, 0, 0, 0, 0)
}

func mouseScroll(delta int) {
	procMouseEvent.Call(wheelEvent, 0, 0, uintptr(uint32(delta)), 0)
}

func keyTap(vk byte) {
	procKeybdEvent.Call(uintptr(vk), 0, 0, 0)
	time.Sleep(10 * time.Millisecond)
	procKeybdEvent.Call(uintptr(vk), 0, keyUpFlag, 0)
}

func parseInt(m map[string]string, key string, def int) int {
	v, ok := m[key]
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func boolParam(m map[string]string, key string, def bool) bool {
	v, ok := m[key]
	if !ok {
		return def
	}
	return strings.EqualFold(v, "true") || v == "1"
}

func (r *Runner) subst(s string) string {
	for k, v := range r.vars {
		s = strings.ReplaceAll(s, "${"+k+"}", v)
	}
	return s
}

func (r *Runner) resolveValue(s string) string {
	s = strings.TrimSpace(r.subst(s))
	if strings.HasPrefix(s, "$") {
		k := strings.TrimPrefix(s, "$")
		return r.vars[k]
	}
	return s
}

func (r *Runner) jumpLabel(label string) (int, error) {
	idx, ok := r.labels[label]
	if !ok {
		return 0, fmt.Errorf("label inexistente: %s", label)
	}
	return idx, nil
}

func (r *Runner) run() error {
	for i := 0; i < len(r.script.Actions); i++ {
		a := r.script.Actions[i]
		if !a.Enabled && a.Type != "" {
			continue
		}
		if keyPressed(0x1B) {
			return errors.New("interrompido por ESC")
		}

		next, jumped, err := r.execAction(i, a)
		if err != nil {
			return err
		}
		if jumped {
			i = next - 1
		}
	}
	return nil
}

func (r *Runner) execAction(index int, a Action) (int, bool, error) {
	p := a.Params
	t := strings.ToLower(a.Type)
	switch t {
	case "wait", "wait_time":
		ms := parseInt(p, "ms", 500)
		rndMax := parseInt(p, "random_max_ms", 0)
		if rndMax > ms {
			ms += rand.Intn(rndMax - ms + 1)
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
	case "wait_until_time":
		target := r.resolveValue(p["hhmmss"])
		tm, err := time.Parse("15:04:05", target)
		if err != nil {
			return 0, false, fmt.Errorf("wait_until_time inválido: %w", err)
		}
		now := time.Now()
		d := time.Date(now.Year(), now.Month(), now.Day(), tm.Hour(), tm.Minute(), tm.Second(), 0, now.Location())
		if d.Before(now) {
			d = d.Add(24 * time.Hour)
		}
		time.Sleep(time.Until(d))
	case "wait_hotkey":
		vk := parseInt(p, "vk", 0x78)
		for !keyPressed(vk) {
			time.Sleep(30 * time.Millisecond)
		}
	case "wait_file_exists":
		path := r.resolveValue(p["path"])
		timeout := parseInt(p, "timeout_ms", 30000)
		deadline := time.Now().Add(time.Duration(timeout) * time.Millisecond)
		for {
			if _, err := os.Stat(path); err == nil {
				break
			}
			if time.Now().After(deadline) {
				if label := p["on_timeout_goto"]; label != "" {
					idx, err := r.jumpLabel(label)
					return idx, true, err
				}
				return 0, false, fmt.Errorf("timeout esperando arquivo: %s", path)
			}
			time.Sleep(200 * time.Millisecond)
		}
	case "mouse_move":
		x := parseInt(p, "x", 0)
		y := parseInt(p, "y", 0)
		setCursorPos(int32(x), int32(y))
	case "mouse_click":
		button := r.resolveValue(p["button"])
		if button == "" {
			button = "left"
		}
		x := parseInt(p, "x", -1)
		y := parseInt(p, "y", -1)
		if x >= 0 && y >= 0 {
			setCursorPos(int32(x), int32(y))
		}
		mouseClick(button, true)
		time.Sleep(20 * time.Millisecond)
		mouseClick(button, false)
	case "mouse_scroll":
		delta := parseInt(p, "delta", 120)
		mouseScroll(delta)
	case "key_press":
		vk := parseInt(p, "vk", 0x0D)
		keyTap(byte(vk))
	case "hotkey":
		mods := strings.Split(strings.ToUpper(r.resolveValue(p["mods"])), "+")
		vk := parseInt(p, "vk", 0x0D)
		for _, m := range mods {
			switch strings.TrimSpace(m) {
			case "CTRL":
				procKeybdEvent.Call(0x11, 0, 0, 0)
			case "ALT":
				procKeybdEvent.Call(0x12, 0, 0, 0)
			case "SHIFT":
				procKeybdEvent.Call(0x10, 0, 0, 0)
			}
		}
		keyTap(byte(vk))
		for _, m := range mods {
			switch strings.TrimSpace(m) {
			case "CTRL":
				procKeybdEvent.Call(0x11, 0, keyUpFlag, 0)
			case "ALT":
				procKeybdEvent.Call(0x12, 0, keyUpFlag, 0)
			case "SHIFT":
				procKeybdEvent.Call(0x10, 0, keyUpFlag, 0)
			}
		}
	case "text_output":
		text := r.resolveValue(p["text"])
		humanize := boolParam(p, "humanize", false)
		for _, c := range text {
			if c >= 32 && c < 127 {
				keyTap(byte(c))
			}
			if humanize {
				time.Sleep(time.Duration(20+rand.Intn(120)) * time.Millisecond)
			}
		}
	case "set_var":
		name := p["name"]
		if name != "" {
			r.vars[name] = r.resolveValue(p["value"])
		}
	case "calc":
		name := p["name"]
		a := parseInt(map[string]string{"v": r.resolveValue(p["a"])}, "v", 0)
		b := parseInt(map[string]string{"v": r.resolveValue(p["b"])}, "v", 0)
		op := r.resolveValue(p["op"])
		res := 0
		switch op {
		case "+":
			res = a + b
		case "-":
			res = a - b
		case "*":
			res = a * b
		case "/":
			if b != 0 {
				res = a / b
			}
		}
		r.vars[name] = strconv.Itoa(res)
	case "save_var_file":
		name := p["name"]
		path := r.resolveValue(p["path"])
		appendMode := boolParam(p, "append", false)
		v := r.vars[name]
		if appendMode {
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return 0, false, err
			}
			defer f.Close()
			_, _ = f.WriteString(v)
		} else {
			if err := os.WriteFile(path, []byte(v), 0644); err != nil {
				return 0, false, err
			}
		}
	case "show_notification", "show_message":
		fmt.Println(r.resolveValue(p["text"]))
	case "beep":
		fmt.Print("\a")
	case "run_program":
		path := r.resolveValue(p["path"])
		args := strings.Fields(r.resolveValue(p["args"]))
		cmd := exec.Command(path, args...)
		if err := cmd.Start(); err != nil {
			return 0, false, err
		}
	case "if":
		left := r.resolveValue(p["left"])
		right := r.resolveValue(p["right"])
		op := strings.ToLower(strings.TrimSpace(p["op"]))
		ok := false
		switch op {
		case "equals", "==":
			ok = left == right
		case "contains":
			ok = strings.Contains(left, right)
		case "!=":
			ok = left != right
		}
		label := p["else"]
		if ok {
			label = p["then"]
		}
		if label != "" {
			idx, err := r.jumpLabel(label)
			return idx, true, err
		}
	case "goto":
		idx, err := r.jumpLabel(p["target"])
		return idx, true, err
	case "repeat":
		label := p["target"]
		count := parseInt(p, "count", 1)
		if count <= 0 {
			return 0, false, nil
		}
		r.repeatCount[index]++
		if r.repeatCount[index] <= count {
			idx, err := r.jumpLabel(label)
			return idx, true, err
		}
		r.repeatCount[index] = 0
	case "data_list_next":
		name := p["name"]
		file := r.resolveValue(p["file"])
		idxName := name + "__index"
		idx, _ := strconv.Atoi(r.vars[idxName])
		b, err := os.ReadFile(file)
		if err != nil {
			return 0, false, err
		}
		lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
		values := make([]string, 0, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				values = append(values, line)
			}
		}
		if len(values) == 0 {
			r.vars[name] = ""
			return 0, false, nil
		}
		if idx >= len(values) {
			idx = 0
		}
		r.vars[name] = values[idx]
		r.vars[idxName] = strconv.Itoa(idx + 1)
	default:
		return 0, false, fmt.Errorf("ação não suportada: %s", a.Type)
	}
	return 0, false, nil
}

func loadScript(path string) (MacroScript, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return MacroScript{}, err
	}
	var s MacroScript
	if err := json.Unmarshal(b, &s); err != nil {
		return MacroScript{}, err
	}
	return s, nil
}

func buildLabels(s MacroScript) map[string]int {
	labels := map[string]int{}
	for i, a := range s.Actions {
		if strings.TrimSpace(a.Label) != "" {
			labels[a.Label] = i
		}
	}
	return labels
}

func saveEvents(path string, events []Event) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	b, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func recordEvents(in *bufio.Reader) {
	fmt.Println("Arquivo de saída (ex: macro_eventos.json):")
	fmt.Print("> ")
	file, _ := in.ReadString('\n')
	file = strings.TrimSpace(file)
	if file == "" {
		file = "macro_eventos.json"
	}
	fmt.Println("Gravando em 3 segundos... F9 para parar")
	time.Sleep(3 * time.Second)
	start := time.Now()
	events := make([]Event, 0, 2048)
	prevX, prevY := int32(-1), int32(-1)
	prevL, prevR := false, false

	for {
		if keyPressed(0x78) {
			break
		}
		x, y := getCursorPos()
		if x != prevX || y != prevY {
			events = append(events, Event{T: time.Since(start).Milliseconds(), Kind: "mouse_move", Data: map[string]string{"x": strconv.Itoa(int(x)), "y": strconv.Itoa(int(y))}})
			prevX, prevY = x, y
		}
		l := keyPressed(0x01)
		rr := keyPressed(0x02)
		if l != prevL {
			events = append(events, Event{T: time.Since(start).Milliseconds(), Kind: "mouse_left", Data: map[string]string{"down": strconv.FormatBool(l)}})
			prevL = l
		}
		if rr != prevR {
			events = append(events, Event{T: time.Since(start).Milliseconds(), Kind: "mouse_right", Data: map[string]string{"down": strconv.FormatBool(rr)}})
			prevR = rr
		}
		time.Sleep(15 * time.Millisecond)
	}
	if err := saveEvents(file, events); err != nil {
		fmt.Println("Erro ao salvar:", err)
		return
	}
	fmt.Printf("Salvo: %s (%d eventos)\n", file, len(events))
}

func playEvents(in *bufio.Reader) {
	fmt.Println("Arquivo JSON de eventos:")
	fmt.Print("> ")
	file, _ := in.ReadString('\n')
	file = strings.TrimSpace(file)
	if file == "" {
		file = "macro_eventos.json"
	}
	b, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}
	var events []Event
	if err := json.Unmarshal(b, &events); err != nil {
		fmt.Println("JSON inválido:", err)
		return
	}
	fmt.Println("Reproduzindo em 2 segundos... ESC interrompe")
	time.Sleep(2 * time.Second)
	prev := int64(0)
	for _, e := range events {
		if keyPressed(0x1B) {
			fmt.Println("Interrompido por ESC")
			return
		}
		d := e.T - prev
		if d > 0 {
			time.Sleep(time.Duration(d) * time.Millisecond)
		}
		prev = e.T
		switch e.Kind {
		case "mouse_move":
			x, _ := strconv.Atoi(e.Data["x"])
			y, _ := strconv.Atoi(e.Data["y"])
			setCursorPos(int32(x), int32(y))
		case "mouse_left":
			mouseClick("left", e.Data["down"] == "true")
		case "mouse_right":
			mouseClick("right", e.Data["down"] == "true")
		}
	}
	fmt.Println("Fim")
}

func runScript(in *bufio.Reader) {
	fmt.Println("Arquivo JSON de script (acoes):")
	fmt.Print("> ")
	file, _ := in.ReadString('\n')
	file = strings.TrimSpace(file)
	if file == "" {
		file = "macro_script.json"
	}
	s, err := loadScript(file)
	if err != nil {
		fmt.Println("Erro ao ler script:", err)
		return
	}
	r := &Runner{script: s, vars: map[string]string{}, labels: buildLabels(s), repeatCount: map[int]int{}}
	fmt.Println("Executando script... ESC interrompe")
	if err := r.run(); err != nil {
		fmt.Println("Falha:", err)
		return
	}
	fmt.Println("Script concluído")
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Macro Recorder Windows (sem admin)")
	fmt.Println("1) Gravar eventos (mouse)")
	fmt.Println("2) Reproduzir eventos")
	fmt.Println("3) Executar script de ações (variáveis/fluxo/waits)")
	fmt.Print("> ")
	in := bufio.NewReader(os.Stdin)
	opt, _ := in.ReadString('\n')
	opt = strings.TrimSpace(opt)
	switch opt {
	case "1":
		recordEvents(in)
	case "2":
		playEvents(in)
	case "3":
		runScript(in)
	default:
		fmt.Println("Opção inválida")
	}
}
