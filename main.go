package main

import (
	"bytes"
	"errors"
	"flag"
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
	colorMode     string
	themeName     string
	dryRun        bool
	showVer       bool
	listThemes    bool
	validateTheme string
	watchMode     bool
	noShade       bool
}

const version = "0.8.0"

const drySample = `NAMESPACE     NAME                        READY   STATUS              RESTARTS   AGE
default       web-1                       1/1     Running             0          12h
default       web-2                       0/1     CrashLoopBackOff    7          12h
default       db-0                        0/1     Pending             0          5m
default       cache-6b8d4                 0/1     ContainerCreating   0          30s
default       old-job-x7f2                0/1     Evicted             0          24h
kube-system   coredns-5d4b                1/1     Running             0          30d
kube-system   metrics-server              0/1     ImagePullBackOff    3          2h
default       batch-processor             0/1     Error               1          10m
default       init-container-pod          0/1     Init:0/1            0          1m
default       long-running                1/1     Running             0          7d
default       failed-build-1              0/1     Failed              0          1h
default       node-affinity-pod           0/1     NodeAffinity        0          15m
default       big-data                    1/1     Running             0          3d
default       pending-pod                 0/1     Unknown             0          5m
default       OOM-killed-app              0/1     OOMKilled           0          1m
default       terminated-job              0/1     Completed           0          6h
`

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "upgrade":
			printUpgrade()
			return
		case "themes", "theme":
			runThemePicker()
			return
		}
	}

	flags, remaining := parseFlags(args)

	if flags.showVer {
		fmt.Printf("oc-color v%s\n", version)
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

	themeName := flags.themeName
	if themeName == "" {
		themeName = cfg.Theme
	}
	if themeName == "" {
		themeName = "dracula"
	}

	th, ok := theme.Get(themeName)
	if !ok {
		fmt.Fprintf(os.Stderr, "error: unknown theme %q (available: %s)\n",
			themeName, strings.Join(theme.Names(), ", "))
		os.Exit(1)
	}

	shadeEnabled := cfg.ShadeEnabled()
	if flags.noShade {
		shadeEnabled = false
	}

	if flags.dryRun {
		dryRun(th, useColor, shadeEnabled)
		return
	}

	proc := output.Processor{Theme: th, Colour: useColor, Shade: shadeEnabled}

	if flags.watchMode || isWatchMode(remaining) {
		if !useColor {
			cmd := exec.Command("oc", remaining...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			fmt.Print(stdout.String())
			if err != nil {
				fmt.Fprint(os.Stderr, stderr.String())
				os.Exit(1)
			}
			return
		}
		watchArgs := remaining
		if flags.watchMode {
			watchArgs = append([]string{"-w"}, remaining...)
		}
		err := runWatch(watchArgs, &proc)
		if err != nil && !errors.Is(err, errInterrupted) {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	cmd := exec.Command("oc", remaining...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		fmt.Fprint(os.Stderr, proc.Process(stderr.String()))
		os.Exit(1)
	}

	fmt.Print(proc.Process(stdout.String()))
}

func parseFlags(args []string) (flags, []string) {
	f := flags{}
	var noColor bool
	fs := flag.NewFlagSet("oc-color", flag.ContinueOnError)
	fs.StringVar(&f.colorMode,     "color",   "",  "Color mode: always, never, auto")
	fs.BoolVar(&noColor,           "no-color", false, "Shorthand for --color=never")
	fs.BoolVar(&f.noShade,         "no-shade", false, "Disable zebra-stripe row shading")
	fs.StringVar(&f.themeName,     "theme",   "",  "Theme name (default: dracula, or from config)")
	fs.BoolVar(&f.dryRun,          "dry-run",        false,     "Process sample output to preview colors")
	fs.BoolVar(&f.showVer,         "version",        false,     "Print version")
	fs.BoolVar(&f.listThemes,      "list-themes",    false,     "List available themes")
	fs.StringVar(&f.validateTheme, "validate-theme", "",        "Validate a theme YAML file")
	fs.BoolVar(&f.watchMode,       "watch",          false,     "Watch mode")
	fs.Usage = printHelp
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		os.Exit(2)
	}
	if noColor {
		f.colorMode = "never"
	}
	return f, fs.Args()
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
  oc color upgrade
  oc color themes

Flags:
  --color <mode>       Color mode: always, never, auto (default: auto)
  --no-color           Shorthand for --color=never
  --no-shade           Disable zebra-stripe row shading
  --theme <name>       Theme name (default: dracula)
  --list-themes        List available themes
  --validate-theme <path>  Validate a theme YAML file
  --watch              Watch mode (equivalent to oc -w). Clean in-place redraw.
  --dry-run            Process sample output to preview colors
  --version            Print version
  --help, -h           Show this help

Examples:
  oc color get pods
  oc color get pods -w
  oc color --watch get pods
  oc color --color=always get pods | less -R
  oc color --theme dracula get pods -o json
  oc color --theme nord get pods
  oc color --no-shade get pods
  oc color --list-themes
  oc color --validate-theme ~/.config/oc-color/themes/nord.yaml
  oc color --dry-run
  oc color themes              # interactive theme picker

Config: ~/.config/oc-color/config.yaml
Themes:  ~/.config/oc-color/themes/*.yaml
`)
}

func goInstalled(lookPath func(string) (string, error)) bool {
	_, err := lookPath("go")
	return err == nil
}

func printUpgrade() {
	if !goInstalled(exec.LookPath) {
		fmt.Fprintln(os.Stderr, "error: 'oc color upgrade' requires Go to be installed.")
		fmt.Fprintln(os.Stderr, "Install Go from https://golang.org/dl, then re-run this command.")
		fmt.Fprintln(os.Stderr, "If you installed via Krew, upgrade with: kubectl krew upgrade oc-color")
		os.Exit(1)
	}
	fmt.Println("Upgrading oc-color to the latest version...")
	cmd := exec.Command("go", "install", "github.com/thephilip/oc-color@latest")
	cmd.Env = append(os.Environ(), "GOPROXY=direct")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: upgrade failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Upgrade complete.")
}

func dryRun(th theme.Theme, useColor bool, shade bool) {
	proc := output.Processor{Theme: th, Colour: useColor, Shade: shade}
	fmt.Print(proc.Process(drySample))
}
