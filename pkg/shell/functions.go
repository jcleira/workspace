// Package shell provides shell integration and navigation functions.
package shell

// GenerateBashFunction returns a bash shell function for workspace navigation.
func GenerateBashFunction() string {
	return `# Workspace navigation function with Claude Code integration
w() {
    if ! command -v claude &> /dev/null; then
        echo "Error: 'claude' command not found"
        echo "Install: https://claude.ai/code"
        return 1
    fi

    while true; do
        if [ $# -eq 0 ]; then
            local result=$(workspace --output-path-only)
            if [ -z "$result" ] || [ "$result" = "quit" ]; then
                break
            fi
            cd "$result"
        else
            local workspace_path="${HOME}/Tactic/workspaces/workspace-$1"
            if [ -d "$workspace_path" ]; then
                cd "$workspace_path"
            else
                echo "Workspace '$1' not found"
                workspace list
                return 1
            fi
            set --
        fi

        claude --continue
    done
}

# Completion function for w command
_w_complete() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local workspaces=$(find "${HOME}/Tactic/workspaces" -maxdepth 1 -type d -name "workspace-*" 2>/dev/null | sed 's|.*/workspace-||' | sort)
    COMPREPLY=($(compgen -W "${workspaces}" -- "${cur}"))
}

# Register completion for w function
complete -F _w_complete w`
}

// GenerateZshFunction returns a zsh shell function for workspace navigation.
func GenerateZshFunction() string {
	return `# Workspace navigation function with Claude Code integration
w() {
    if ! command -v claude &> /dev/null; then
        echo "Error: 'claude' command not found"
        echo "Install: https://claude.ai/code"
        return 1
    fi

    while true; do
        if [ $# -eq 0 ]; then
            local result=$(workspace --output-path-only)
            if [ -z "$result" ] || [ "$result" = "quit" ]; then
                break
            fi
            cd "$result"
        else
            local workspace_path="${HOME}/Tactic/workspaces/workspace-$1"
            if [ -d "$workspace_path" ]; then
                cd "$workspace_path"
            else
                echo "Workspace '$1' not found"
                workspace list
                return 1
            fi
            set --
        fi

        claude --continue
    done
}

# Completion function for w command
_w_complete() {
    local -a workspaces
    workspaces=(${(f)"$(find "${HOME}/Tactic/workspaces" -maxdepth 1 -type d -name "workspace-*" 2>/dev/null | sed 's|.*/workspace-||' | sort)"})
    _describe 'workspace' workspaces
}

# Register completion for w function
compdef _w_complete w`
}

// GenerateFishFunction returns a fish shell function for workspace navigation.
func GenerateFishFunction() string {
	return `# Workspace navigation function with Claude Code integration
function w --description "Navigate to workspace and launch Claude"
    if not command -v claude &> /dev/null
        echo "Error: 'claude' command not found"
        echo "Install: https://claude.ai/code"
        return 1
    end

    while true
        if test (count $argv) -eq 0
            set result (workspace --output-path-only)
            if test -z "$result" -o "$result" = "quit"
                break
            end
            cd "$result"
        else
            set workspace_path "$HOME/Tactic/workspaces/workspace-$argv[1]"
            if test -d "$workspace_path"
                cd "$workspace_path"
            else
                echo "Workspace '$argv[1]' not found"
                workspace list
                return 1
            end
            set -e argv
        end

        claude --continue
    end
end

# Completion for w command
complete -c w -f -a "(find $HOME/Tactic/workspaces -maxdepth 1 -type d -name 'workspace-*' 2>/dev/null | sed 's|.*/workspace-||' | sort)"`
}
