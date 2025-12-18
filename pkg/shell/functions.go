// Package shell provides shell integration and navigation functions.
package shell

// GenerateBashFunction returns a bash shell function for workspace navigation.
func GenerateBashFunction() string {
	return `# Workspace navigation function
w() {
    if [ $# -eq 0 ]; then
        # Interactive selection
        local result=$(workspace --output-path-only)
        if [ -n "$result" ] && [ "$result" != "quit" ]; then
            cd "$result"
        fi
    else
        # Direct workspace navigation
        local workspace_path="${HOME}/Tactic/workspaces/workspace-$1"
        if [ -d "$workspace_path" ]; then
            cd "$workspace_path"
        else
            echo "Workspace '$1' not found"
            workspace list
        fi
    fi
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
	return `# Workspace navigation function
w() {
    if [ $# -eq 0 ]; then
        # Interactive selection
        local result=$(workspace --output-path-only)
        if [ -n "$result" ] && [ "$result" != "quit" ]; then
            cd "$result"
        fi
    else
        # Direct workspace navigation
        local workspace_path="${HOME}/Tactic/workspaces/workspace-$1"
        if [ -d "$workspace_path" ]; then
            cd "$workspace_path"
        else
            echo "Workspace '$1' not found"
            workspace list
        fi
    fi
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
	return `# Workspace navigation function
function w --description "Navigate to workspace"
    if test (count $argv) -eq 0
        # Interactive selection
        set result (workspace --output-path-only)
        if test -n "$result" -a "$result" != "quit"
            cd "$result"
        end
    else
        # Direct workspace navigation
        set workspace_path "$HOME/Tactic/workspaces/workspace-$argv[1]"
        if test -d "$workspace_path"
            cd "$workspace_path"
        else
            echo "Workspace '$argv[1]' not found"
            workspace list
        end
    end
end

# Completion for w command
complete -c w -f -a "(find $HOME/Tactic/workspaces -maxdepth 1 -type d -name 'workspace-*' 2>/dev/null | sed 's|.*/workspace-||' | sort)"`
}
