package main

import (
	"fmt"
	"os"
)

func printCompletion(shell string) {
	switch shell {
	case "bash":
		printBashCompletion()
	case "zsh":
		printZshCompletion()
	case "fish":
		printFishCompletion()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown shell %q (supported: bash, zsh, fish)\n", shell)
		os.Exit(1)
	}
}

func printBashCompletion() {
	fmt.Print(`_oc-color() {
    local cur prev words cword
    _init_completion || return

    case $prev in
        --color)
            COMPREPLY=($(compgen -W "always never auto" -- "$cur"))
            return
            ;;
        --theme)
            COMPREPLY=($(compgen -W "$(oc-color --list-themes 2>/dev/null)" -- "$cur"))
            return
            ;;
        --validate-theme)
            COMPREPLY=($(compgen -f -- "$cur"))
            return
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur"))
            return
            ;;
    esac

    if [[ $cur == -* ]]; then
        COMPREPLY=($(compgen -W '--color --no-color --theme --list-themes --validate-theme --dry-run --version --help' -- "$cur"))
    else
        COMPREPLY=($(compgen -W 'completion' -- "$cur"))
    fi
} && complete -F _oc-color oc-color
`)
}

func printZshCompletion() {
	fmt.Print(`#compdef oc-color

_oc-color() {
    local context state state_descr line
    typeset -A opt_args

    _arguments \
        '--color[Color mode]:mode:(always never auto)' \
        '--no-color[Disable color]' \
        '--theme[Theme name]:theme:(dracula)' \
        '--list-themes[List available themes]' \
        '--validate-theme[Validate a theme file]:file:_files' \
        '--dry-run[Preview colors with sample output]' \
        '--version[Print version]' \
        '--help[Show help]' \
        '1: :->subcommand' \
        '*: :->args'

    case $state in
        subcommand)
            _values 'subcommand' completion
            ;;
        args)
            case $words[1] in
                completion)
                    _values 'shell' bash zsh fish
                    ;;
            esac
            ;;
    esac
}

_oc-color "$@"
`)
}

func printFishCompletion() {
	fmt.Print(`complete -c oc-color -l color -xa "always never auto" -d "Color mode"
complete -c oc-color -l no-color -d "Disable color"
complete -c oc-color -l theme -xa "(oc-color --list-themes 2>/dev/null)" -d "Theme name"
complete -c oc-color -l list-themes -d "List available themes"
complete -c oc-color -l validate-theme -r -d "Validate a theme file"
complete -c oc-color -l dry-run -d "Preview colors with sample output"
complete -c oc-color -l version -d "Print version"
complete -c oc-color -l help -d "Show help"
complete -c oc-color -n "not __fish_seen_subcommand_from completion" -xa completion -d "Generate completion script"
complete -c oc-color -n "__fish_seen_subcommand_from completion" -xa "bash zsh fish" -d "Shell type"
`)
}
