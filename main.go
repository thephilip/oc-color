package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

// ANSI color codes.
const (
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	reset  = "\033[0m"
)

func main() {
	// Check if --no-color flag is set and remove it from args.
	args, colorizeOutput := processArgs(os.Args[1:])

	// Execute the oc command with provided arguments.
	cmd := exec.Command("oc", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// If there was an error, output stderr (colorizing if desired).
		fmt.Fprint(os.Stderr, processOutput(stderr.String(), colorizeOutput))
		os.Exit(1)
	}

	// Process and print stdout.
	fmt.Print(processOutput(stdout.String(), colorizeOutput))
}

// processArgs removes the --no-color flag if present and returns the remaining args.
func processArgs(args []string) ([]string, bool) {
	colorize := true
	filtered := []string{}
	for _, arg := range args {
		// A simple flag check; you could make this more robust.
		if arg == "--no-color" {
			colorize = false
		} else {
			filtered = append(filtered, arg)
		}
	}
	return filtered, colorize
}

// processOutput applies colorization to output if enabled.
func processOutput(output string, colorize bool) string {
	if !colorize {
		return output
	}

	// Example: colorize common statuses.
	// You can add or modify these patterns as needed.
	output = colorizePattern(output, `\bRunning\b`, green)
	output = colorizePattern(output, `\bPending\b`, yellow)
	output = colorizePattern(output, `\bError\b`, red)

	// Extend here with more advanced highlighting or even integration
	// with a syntax highlighter for JSON/YAML if needed.

	return output
}

// colorizePattern wraps each regex match in the given ANSI color.
func colorizePattern(s, pattern, color string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// In case you want to do further processing, it goes here.
		return fmt.Sprintf("%s%s%s", color, match, reset)
	})
}
