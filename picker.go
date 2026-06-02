package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/thephilip/oc-color/config"
	"github.com/thephilip/oc-color/output"
	"github.com/thephilip/oc-color/theme"
	"golang.org/x/term"
)

func runThemePicker() {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprintln(os.Stderr, "error: theme picker requires an interactive terminal")
		os.Exit(1)
	}

	themes := theme.Names()
	cfg, _ := config.Load()

	cursor := 0
	for i, name := range themes {
		if name == cfg.Theme {
			cursor = i
			break
		}
	}

	selected, ok := runPickerLoop(themes, cursor)
	fmt.Print("\033[2J\033[H")
	if !ok {
		fmt.Println("No changes made.")
		return
	}

	cfg.Theme = selected
	path, err := config.Save(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Theme set: %s\n", selected)
	fmt.Printf("Written to %s\n", path)
}

func runPickerLoop(themes []string, cursor int) (selected string, ok bool) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot enter raw terminal mode: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	drawPicker(themes, cursor)

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf[:1])
		if n == 0 || err != nil {
			return "", false
		}
		switch buf[0] {
		case 'q', 3: // q or Ctrl-C
			return "", false
		case '\r', '\n': // Enter
			return themes[cursor], true
		case '\033': // escape sequence (arrow keys)
			n2, _ := os.Stdin.Read(buf[1:3])
			if n2 == 2 && buf[1] == '[' {
				switch buf[2] {
				case 'A': // up arrow
					cursor = (cursor - 1 + len(themes)) % len(themes)
				case 'B': // down arrow
					cursor = (cursor + 1) % len(themes)
				default:
					return "", false
				}
			} else {
				return "", false // bare ESC or unrecognized sequence → cancel
			}
		case 'k': // vim up
			cursor = (cursor - 1 + len(themes)) % len(themes)
		case 'j': // vim down
			cursor = (cursor + 1) % len(themes)
		}
		drawPicker(themes, cursor)
	}
}

func drawPicker(themes []string, cursor int) {
	fmt.Print("\033[2J\033[H")
	fmt.Print("Select a theme  (↑↓ or j/k navigate · Enter select · q quit)\r\n\r\n")
	for i, name := range themes {
		if i == cursor {
			fmt.Printf("▶ %s\r\n", name)
		} else {
			fmt.Printf("  %s\r\n", name)
		}
	}
	fmt.Print("\r\n── Preview ─────────────────────────────────────────────\r\n")
	th, ok := theme.Get(themes[cursor])
	if ok {
		proc := output.Processor{Theme: th, Colour: true, Shade: true}
		preview := strings.ReplaceAll(proc.Process(drySample), "\n", "\r\n")
		fmt.Print(preview)
	}
}
