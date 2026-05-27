package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thephilip/oc-color/config"
	"github.com/thephilip/oc-color/output"
	"github.com/thephilip/oc-color/theme"
	"golang.org/x/term"
)

type flags struct {
	colorMode string
	themeName string
	dryRun    bool
	showVer   bool
	showHelp  bool
}

const version = "0.4.0"

func main() {
	flags, args := parseFlags(os.Args[1:])

	if flags.showVer {
		fmt.Printf("oc-color v%s\n", version)
		return
	}

	if flags.showHelp {
		printHelp()
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
	}

	colorMode := resolveColorMode(flags.colorMode, cfg.Color)
	useColor := shouldColorize(colorMode)

	th, ok := theme.Get(flags.themeName)
	if !ok {
		fmt.Fprintf(os.Stderr, "error: unknown theme %q (available: %s)\n",
			flags.themeName, strings.Join(theme.Names(), ", "))
		os.Exit(1)
	}

	if flags.dryRun {
		dryRun(th, useColor)
		return
	}

	cmd := exec.Command("oc", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	proc := output.Processor{Theme: th, Colour: useColor}

	if err != nil {
		fmt.Fprint(os.Stderr, proc.Process(stderr.String()))
		os.Exit(1)
	}

	fmt.Print(proc.Process(stdout.String()))
}

func parseFlags(args []string) (flags, []string) {
	f := flags{colorMode: "", themeName: "dracula"}
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--no-color":
			f.colorMode = "never"
		case arg == "--color" || strings.HasPrefix(arg, "--color="):
			f.colorMode = flagValue(arg, "--color", &i, args)
		case arg == "--theme" || strings.HasPrefix(arg, "--theme="):
			f.themeName = flagValue(arg, "--theme", &i, args)
		case arg == "--dry-run":
			f.dryRun = true
		case arg == "--version":
			f.showVer = true
		case arg == "--help" || arg == "-h":
			f.showHelp = true
		default:
			remaining = append(remaining, arg)
		}
	}
	return f, remaining
}

func flagValue(arg, prefix string, i *int, args []string) string {
	if strings.HasPrefix(arg, prefix+"=") {
		return strings.TrimPrefix(arg, prefix+"=")
	}
	if *i+1 < len(args) {
		*i++
		return args[*i]
	}
	return ""
}

func resolveColorMode(flagMode, cfgMode string) string {
	if flagMode != "" {
		return flagMode
	}
	return cfgMode
}

func shouldColorize(mode string) bool {
	switch mode {
	case "always":
		return true
	case "never":
		return false
	default:
		return term.IsTerminal(int(os.Stdout.Fd()))
	}
}

func printHelp() {
	fmt.Print(`oc-color — colorize oc command output

Usage:
  oc color [flags] -- <oc-args>

Flags:
  --color <mode>    Color mode: always, never, auto (default: auto)
  --no-color        Shorthand for --color=never
  --theme <name>    Theme name (default: dracula)
  --dry-run         Process sample output to preview colors
  --version         Print version
  --help, -h        Show this help

Examples:
  oc color get pods
  oc color --color=always get pods | less -R
  oc color --theme dracula get pods -o json
  oc color --dry-run

Config: ~/.config/oc-color/config.yaml
`)
}

func dryRun(th theme.Theme, useColor bool) {
	sample := `NAMESPACE     NAME                        READY   STATUS              RESTARTS   AGE
default       web-1                        1/1     Running             0          12h
default       web-2                        0/1     CrashLoopBackOff    7          12h
default       db-0                         0/1     Pending             0          5m
default       cache-6b8d4                 0/1     ContainerCreating   0          30s
default       old-job-x7f2                 0/1     Evicted             0          24h
kube-system   coredns-5d4b                1/1     Running             0          30d
kube-system   metrics-server              0/1     ImagePullBackOff    3          2h
default       batch-processor              0/1     Error               1          10m
default       init-container-pod            0/1     Init:0/1            0          1m
default       long-running                  1/1     Running             0          7d
default       failed-build-1                0/1     Failed              0          1h
default       node-affinity-pod             0/1     NodeAffinity        0          15m
default       big-data                      1/1     Running             0          3d
default       pending-pod                   0/1     Unknown             0          5m
default       OOM-killed-app                0/1     OOMKilled           0          1m
default       terminated-job                0/1     Completed           0          6h
`

	proc := output.Processor{Theme: th, Colour: useColor}
	fmt.Print(proc.Process(sample))
}
