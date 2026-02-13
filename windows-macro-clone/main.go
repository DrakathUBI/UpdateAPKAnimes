package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MacroScript struct {
	Name    string       `json:"name"`
	Actions []ActionItem `json:"actions"`
}

type ActionItem struct {
	Type       string `json:"type"`
	Enabled    *bool  `json:"enabled,omitempty"`
	Label      string `json:"label,omitempty"`
	Comment    string `json:"comment,omitempty"`
	Variable   string `json:"variableName,omitempty"`
	Value      string `json:"value,omitempty"`
	Expression string `json:"expression,omitempty"`
	Target     string `json:"targetLabel,omitempty"`
	ThenLabel  string `json:"thenLabel,omitempty"`
	ElseLabel  string `json:"elseLabel,omitempty"`
	Operator   string `json:"operator,omitempty"`
	Left       string `json:"left,omitempty"`
	Right      string `json:"right,omitempty"`
	FilePath   string `json:"filePath,omitempty"`
	Arguments  string `json:"arguments,omitempty"`
	Append     bool   `json:"append,omitempty"`
	WaitMs     int    `json:"waitMs,omitempty"`
	MinMs      int    `json:"minMs,omitempty"`
	MaxMs      int    `json:"maxMs,omitempty"`
	TimeAt     string `json:"timeAt,omitempty"`
	Repeat     int    `json:"repeat,omitempty"`
	Condition  string `json:"condition,omitempty"`
}

type Runtime struct {
	vars      map[string]string
	labels    map[string]int
	repeatMap map[int]int
	silent    bool
}

func main() {
	play := flag.String("play", "", "arquivo macro JSON")
	silent := flag.Bool("silent", false, "modo silencioso")
	startAt := flag.String("start-at", "", "label inicial")
	flag.Parse()

	if *play == "" {
		fmt.Println("Uso: MacroClone.exe --play macro.json [--silent] [--start-at Label]")
		os.Exit(1)
	}

	data, err := os.ReadFile(*play)
	must(err)

	var script MacroScript
	must(json.Unmarshal(data, &script))

	rt := &Runtime{
		vars:      map[string]string{},
		labels:    map[string]int{},
		repeatMap: map[int]int{},
		silent:    *silent,
	}

	for i, a := range script.Actions {
		if strings.TrimSpace(a.Label) != "" {
			rt.labels[strings.ToLower(a.Label)] = i
		}
	}

	ptr := 0
	if strings.TrimSpace(*startAt) != "" {
		if p, ok := rt.labels[strings.ToLower(*startAt)]; ok {
			ptr = p
		}
	}

	for ptr < len(script.Actions) {
		a := script.Actions[ptr]
		if a.Enabled != nil && !*a.Enabled {
			ptr++
			continue
		}

		if !rt.silent {
			fmt.Printf("[%d] %s %s\n", ptr+1, a.Type, a.Comment)
		}

		next, err := rt.execAction(a, ptr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "erro na ação %d (%s): %v\n", ptr+1, a.Type, err)
			os.Exit(2)
		}
		ptr = next
	}
}

func (rt *Runtime) execAction(a ActionItem, ptr int) (int, error) {
	switch strings.ToLower(strings.TrimSpace(a.Type)) {
	case "waittime":
		ms := a.WaitMs
		if a.MinMs > 0 && a.MaxMs >= a.MinMs {
			ms = rand.Intn(a.MaxMs-a.MinMs+1) + a.MinMs
		}
		if ms <= 0 {
			ms = 1000
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return ptr + 1, nil

	case "waituntiltime":
		t, err := parseTimeToday(a.TimeAt)
		if err != nil {
			return 0, err
		}
		for time.Now().Before(t) {
			time.Sleep(200 * time.Millisecond)
		}
		return ptr + 1, nil

	case "waitforfileevent":
		if a.FilePath == "" {
			return 0, errors.New("filePath vazio")
		}
		deadline := time.Now().Add(timeoutFromAction(a))
		lastInfo, _ := os.Stat(rt.expand(a.FilePath))
		for {
			path := rt.expand(a.FilePath)
			info, err := os.Stat(path)
			if err == nil {
				if strings.EqualFold(a.Condition, "exists") || a.Condition == "" {
					break
				}
				if strings.EqualFold(a.Condition, "changed") && lastInfo != nil && info.ModTime().After(lastInfo.ModTime()) {
					break
				}
			}
			if time.Now().After(deadline) {
				return rt.onTimeout(a, ptr)
			}
			time.Sleep(300 * time.Millisecond)
		}
		return ptr + 1, nil

	case "setvariable":
		if a.Variable == "" {
			return 0, errors.New("variableName vazio")
		}
		rt.vars[strings.ToLower(a.Variable)] = rt.expand(a.Value)
		return ptr + 1, nil

	case "calculation":
		if a.Variable == "" {
			return 0, errors.New("variableName vazio")
		}
		res, err := evalSimple(rt.expand(a.Expression))
		if err != nil {
			return 0, err
		}
		rt.vars[strings.ToLower(a.Variable)] = strconv.FormatFloat(res, 'f', -1, 64)
		return ptr + 1, nil

	case "textoutput", "shownotification":
		fmt.Println(rt.expand(a.Value))
		return ptr + 1, nil

	case "showmessagebox", "waitfortextinput":
		fmt.Printf("%s ", rt.expand(a.Value))
		r := bufio.NewReader(os.Stdin)
		line, _ := r.ReadString('\n')
		line = strings.TrimSpace(line)
		rt.vars["lastinput"] = line
		if strings.EqualFold(a.Type, "waitfortextinput") && a.Value != "" && !strings.EqualFold(line, rt.expand(a.Value)) {
			return rt.onTimeout(a, ptr)
		}
		return ptr + 1, nil

	case "waitforhotkey":
		fmt.Println("Pressione ENTER para continuar...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		return ptr + 1, nil

	case "ifthenelse":
		ok, err := rt.evalCondition(a)
		if err != nil {
			return 0, err
		}
		label := a.ElseLabel
		if ok {
			label = a.ThenLabel
		}
		if label == "" {
			return ptr + 1, nil
		}
		p, ok2 := rt.labels[strings.ToLower(label)]
		if !ok2 {
			return 0, fmt.Errorf("label não encontrada: %s", label)
		}
		return p, nil

	case "goto":
		if a.Target == "" {
			return 0, errors.New("targetLabel vazio")
		}
		p, ok := rt.labels[strings.ToLower(a.Target)]
		if !ok {
			return 0, fmt.Errorf("label não encontrada: %s", a.Target)
		}
		return p, nil

	case "repeat":
		if a.Target == "" {
			return 0, errors.New("targetLabel vazio")
		}
		if a.Repeat <= 0 {
			return ptr + 1, nil
		}
		count := rt.repeatMap[ptr] + 1
		rt.repeatMap[ptr] = count
		if count <= a.Repeat {
			p, ok := rt.labels[strings.ToLower(a.Target)]
			if !ok {
				return 0, fmt.Errorf("label não encontrada: %s", a.Target)
			}
			return p, nil
		}
		delete(rt.repeatMap, ptr)
		return ptr + 1, nil

	case "executeprogram":
		if a.FilePath == "" {
			return ptr + 1, nil
		}
		cmd := exec.Command(rt.expand(a.FilePath), strings.Fields(rt.expand(a.Arguments))...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Start()
		return ptr + 1, nil

	case "savevariable":
		if a.FilePath == "" || a.Variable == "" {
			return ptr + 1, nil
		}
		v := rt.vars[strings.ToLower(a.Variable)]
		p := rt.expand(a.FilePath)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil && filepath.Dir(p) != "." {
			return 0, err
		}
		if a.Append {
			f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return 0, err
			}
			defer f.Close()
			_, err = f.WriteString(v + "\n")
			return ptr + 1, err
		}
		return ptr + 1, os.WriteFile(p, []byte(v), 0o644)

	case "scrapewebpage", "findimage", "findtextocr", "capturebitmap", "capturetextocr", "capturebarcodeqr", "waitpixelcolor", "waitdesktopchange", "smartclick", "mouseclick", "mousemove", "mousescroll", "keypress", "hotkey", "windowfocus", "embedmacro", "datalist", "beep":
		if strings.EqualFold(a.Type, "beep") {
			fmt.Print("\a")
		}
		fmt.Printf("[INFO] %s: recurso avançado depende de integração de SO/UI; executando em modo sem-admin (no-op seguro).\n", a.Type)
		return ptr + 1, nil

	default:
		fmt.Printf("[INFO] ação não reconhecida: %s (ignorada)\n", a.Type)
		return ptr + 1, nil
	}
}

func timeoutFromAction(a ActionItem) time.Duration {
	if a.WaitMs > 0 {
		return time.Duration(a.WaitMs) * time.Millisecond
	}
	if a.MaxMs > 0 {
		return time.Duration(a.MaxMs) * time.Millisecond
	}
	return 10 * time.Second
}

func (rt *Runtime) onTimeout(a ActionItem, ptr int) (int, error) {
	if strings.TrimSpace(a.Target) != "" {
		p, ok := rt.labels[strings.ToLower(a.Target)]
		if !ok {
			return 0, fmt.Errorf("label de timeout não encontrada: %s", a.Target)
		}
		return p, nil
	}
	return ptr + 1, nil
}

func (rt *Runtime) evalCondition(a ActionItem) (bool, error) {
	left := rt.expand(a.Left)
	right := rt.expand(a.Right)
	op := strings.ToLower(strings.TrimSpace(a.Operator))

	switch op {
	case "equals", "==":
		return strings.EqualFold(left, right), nil
	case "contains":
		return strings.Contains(strings.ToLower(left), strings.ToLower(right)), nil
	case "regex":
		return regexp.MatchString(right, left)
	case ">", "<", ">=", "<=":
		lf, err := strconv.ParseFloat(left, 64)
		if err != nil {
			return false, err
		}
		rf, err := strconv.ParseFloat(right, 64)
		if err != nil {
			return false, err
		}
		switch op {
		case ">":
			return lf > rf, nil
		case "<":
			return lf < rf, nil
		case ">=":
			return lf >= rf, nil
		default:
			return lf <= rf, nil
		}
	default:
		return false, fmt.Errorf("operador não suportado: %s", a.Operator)
	}
}

func (rt *Runtime) expand(s string) string {
	out := os.Expand(s, func(k string) string {
		v, ok := rt.vars[strings.ToLower(k)]
		if ok {
			return v
		}
		return os.Getenv(k)
	})
	return out
}

func parseTimeToday(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, errors.New("timeAt vazio")
	}
	if t, err := time.Parse("15:04:05", v); err == nil {
		n := time.Now()
		return time.Date(n.Year(), n.Month(), n.Day(), t.Hour(), t.Minute(), t.Second(), 0, n.Location()), nil
	}
	if t, err := time.Parse("15:04", v); err == nil {
		n := time.Now()
		return time.Date(n.Year(), n.Month(), n.Day(), t.Hour(), t.Minute(), 0, 0, n.Location()), nil
	}
	return time.Time{}, errors.New("formato timeAt inválido (use HH:MM ou HH:MM:SS)")
}

func evalSimple(expr string) (float64, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	if expr == "" {
		return 0, errors.New("expressão vazia")
	}
	// parser simples: + - * /
	vals := []float64{}
	ops := []rune{}
	num := ""
	pushNum := func() error {
		if num == "" {
			return nil
		}
		v, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return err
		}
		vals = append(vals, v)
		num = ""
		return nil
	}
	apply := func() error {
		if len(vals) < 2 || len(ops) == 0 {
			return errors.New("expressão inválida")
		}
		b := vals[len(vals)-1]
		a := vals[len(vals)-2]
		op := ops[len(ops)-1]
		vals = vals[:len(vals)-2]
		ops = ops[:len(ops)-1]
		switch op {
		case '+':
			vals = append(vals, a+b)
		case '-':
			vals = append(vals, a-b)
		case '*':
			vals = append(vals, a*b)
		case '/':
			vals = append(vals, a/b)
		}
		return nil
	}
	prec := func(op rune) int {
		if op == '+' || op == '-' {
			return 1
		}
		return 2
	}
	for _, ch := range expr {
		if (ch >= '0' && ch <= '9') || ch == '.' {
			num += string(ch)
			continue
		}
		if err := pushNum(); err != nil {
			return 0, err
		}
		if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			for len(ops) > 0 && prec(ops[len(ops)-1]) >= prec(ch) {
				if err := apply(); err != nil {
					return 0, err
				}
			}
			ops = append(ops, ch)
		} else {
			return 0, fmt.Errorf("caractere inválido: %c", ch)
		}
	}
	if err := pushNum(); err != nil {
		return 0, err
	}
	for len(ops) > 0 {
		if err := apply(); err != nil {
			return 0, err
		}
	}
	if len(vals) != 1 {
		return 0, errors.New("expressão inválida")
	}
	return vals[0], nil
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
