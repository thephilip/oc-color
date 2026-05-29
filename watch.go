package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/thephilip/oc-color/output"
	"golang.org/x/term"
)

var headerPattern = regexp.MustCompile(`^[A-Z][A-Z\s/]+$`)

func isWatchMode(args []string) bool {
	for _, a := range args {
		if a == "-w" || a == "--watch" {
			return true
		}
	}
	return false
}

func runWatch(args []string, proc *output.Processor) error {
	cmd := exec.Command("oc", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	lineCh := make(chan string, 256)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			lineCh <- sc.Text()
		}
		close(lineCh)
	}()

	stderr, _ := cmd.StderrPipe()
	var stderrBuf bytes.Buffer
	go func() {
		io.Copy(&stderrBuf, stderr)
	}()

	isTerm := term.IsTerminal(int(os.Stdout.Fd()))
	if isTerm {
		os.Stdout.WriteString("\033[?25l")
		defer func() {
			os.Stdout.WriteString("\033[?25h")
			os.Stdout.WriteString("\n")
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	batch, ok := readBatch(lineCh, 100*time.Millisecond)
	if !ok {
		return cmd.Wait()
	}

	if isTerm {
		os.Stdout.WriteString("\0337")
	}

	lines := batch
	output := proc.Process(strings.Join(lines, "\n"))
	fmt.Print(output)

	for {
		select {
		case <-sigCh:
			cmd.Process.Kill()
			return fmt.Errorf("interrupted")
		case line, ok := <-lineCh:
			if !ok {
				return cmd.Wait()
			}
			lines = updateLines(lines, line)
			if isTerm {
				os.Stdout.WriteString("\0338\033[J")
				fmt.Print(proc.Process(strings.Join(lines, "\n")))
			} else {
				fmt.Print(proc.Process(line + "\n"))
			}
		}
	}
}

func readBatch(lineCh <-chan string, timeout time.Duration) ([]string, bool) {
	var batch []string
	line, ok := <-lineCh
	if !ok {
		return nil, false
	}
	batch = append(batch, line)
	for {
		select {
		case line, ok := <-lineCh:
			if !ok {
				return batch, true
			}
			batch = append(batch, line)
		case <-time.After(timeout):
			return batch, true
		}
	}
}

func updateLines(lines []string, update string) []string {
	trimmed := strings.TrimSpace(update)
	if trimmed == "" {
		return lines
	}

	if headerPattern.MatchString(trimmed) {
		return []string{update}
	}

	fields := strings.Fields(update)
	if len(fields) == 0 {
		return append(lines, update)
	}
	name := fields[0]

	for i, line := range lines {
		if i == 0 {
			continue
		}
		lf := strings.Fields(line)
		if len(lf) > 0 && lf[0] == name {
			lines[i] = update
			return lines
		}
	}

	return append(lines, update)
}
