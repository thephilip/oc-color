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
	colorMode       string
	themeName       string
	dryRun          bool
	showVer         bool
	showHelp        bool
	listThemes      bool
	validateTheme   string
	completionShell string
	showUpgrade     bool
}

const version = "0.5.0"

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

	if flags.completionShell != "" {
		printCompletion(flags.completionShell)
		return
	}

	if flags.showUpgrade {
		printUpgrade()
		return
	}

	if flags.listThemes {
		fmt.Println("Available themes:")
		for _, name := range theme.Names() {
			fmt.Printf("  %s\n", name)
		}
		return
	}

	if flags.validateTheme != "" {
		if err := theme.Validate(flags.validateTheme); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("theme is valid")
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
		case arg == "--list-themes":
			f.listThemes = true
		case arg == "--validate-theme" || strings.HasPrefix(arg, "--validate-theme="):
			f.validateTheme = flagValue(arg, "--validate-theme", &i, args)
		case arg == "completion":
			f.completionShell = "bash"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				f.completionShell = args[i]
			}
		case arg == "upgrade":
			f.showUpgrade = true
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
  oc color completion <bash|zsh|fish>
  oc color upgrade

Flags:
  --color <mode>       Color mode: always, never, auto (default: auto)
  --no-color           Shorthand for --color=never
  --theme <name>       Theme name (default: dracula)
  --list-themes        List available themes
  --validate-theme <path>  Validate a theme YAML file
  --dry-run            Process sample output to preview colors
  --version            Print version
  --help, -h           Show this help

Examples:
  oc color get pods
  oc color --color=always get pods | less -R
  oc color --theme dracula get pods -o json
  oc color --theme nord get pods
  oc color --list-themes
  oc color --validate-theme ~/.config/oc-color/themes/nord.yaml
  oc color --dry-run

  # Generate shell completion scripts:
  oc color completion bash > /etc/bash_completion.d/oc-color
  oc color completion zsh  > /usr/share/zsh/site-functions/_oc-color
  oc color completion fish > ~/.config/fish/completions/oc-color.fish

Config: ~/.config/oc-color/config.yaml
Themes:  ~/.config/oc-color/themes/*.yaml
`)
}

func printUpgrade() {
	fmt.Print(`To upgrade oc-color to the latest version:

  go install github.com/thephilip/oc-color@latest

This fetches the latest commit, rebuilds the binary into ~/go/bin/oc-color,
and replaces the current version.
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
