package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
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

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procGetAsyncKey  = user32.NewProc("GetAsyncKeyState")
	procGetCursorPos = user32.NewProc("GetCursorPos")
	procSetCursorPos = user32.NewProc("SetCursorPos")
	procMouseEvent   = user32.NewProc("mouse_event")
	leftDown         = uintptr(0x0002)
	leftUp           = uintptr(0x0004)
)

type point struct {
	X int32
	Y int32
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

func mouseClick(down bool) {
	if down {
		procMouseEvent.Call(leftDown, 0, 0, 0, 0)
	} else {
		procMouseEvent.Call(leftUp, 0, 0, 0, 0)
	}
}

func main() {
	fmt.Println("Macro Recorder Windows (sem admin)")
	fmt.Println("1) Gravar macro")
	fmt.Println("2) Reproduzir macro")
	fmt.Print("> ")
	in := bufio.NewReader(os.Stdin)
	opt, _ := in.ReadString('\n')
	opt = strings.TrimSpace(opt)

	switch opt {
	case "1":
		record(in)
	case "2":
		play(in)
	default:
		fmt.Println("Opção inválida")
	}
}

func record(in *bufio.Reader) {
	fmt.Println("Arquivo de saída (ex: macro.json):")
	fmt.Print("> ")
	file, _ := in.ReadString('\n')
	file = strings.TrimSpace(file)
	if file == "" {
		file = "macro.json"
	}
	fmt.Println("Gravando em 3 segundos... pressione F9 para parar")
	time.Sleep(3 * time.Second)

	start := time.Now()
	events := make([]Event, 0, 1024)
	prevX, prevY := int32(-1), int32(-1)
	prevL := false

	for {
		if keyPressed(0x78) { // F9
			break
		}
		x, y := getCursorPos()
		if x != prevX || y != prevY {
			events = append(events, Event{T: time.Since(start).Milliseconds(), Kind: "mouse_move", Data: map[string]string{"x": strconv.Itoa(int(x)), "y": strconv.Itoa(int(y))}})
			prevX, prevY = x, y
		}
		l := keyPressed(0x01) // Left mouse button
		if l != prevL {
			events = append(events, Event{T: time.Since(start).Milliseconds(), Kind: "mouse_left", Data: map[string]string{"down": strconv.FormatBool(l)}})
			prevL = l
		}
		time.Sleep(15 * time.Millisecond)
	}

	b, _ := json.MarshalIndent(events, "", "  ")
	_ = os.WriteFile(file, b, 0644)
	fmt.Printf("Salvo: %s (%d eventos)\n", file, len(events))
}

func play(in *bufio.Reader) {
	fmt.Println("Arquivo macro JSON:")
	fmt.Print("> ")
	file, _ := in.ReadString('\n')
	file = strings.TrimSpace(file)
	if file == "" {
		file = "macro.json"
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
	fmt.Println("Reproduzindo em 3 segundos... Ctrl+C para interromper")
	time.Sleep(3 * time.Second)
	prev := int64(0)
	for _, e := range events {
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
			down := e.Data["down"] == "true"
			mouseClick(down)
		}
	}
	fmt.Println("Fim")
}
