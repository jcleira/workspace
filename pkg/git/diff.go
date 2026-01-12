package git

import (
	"errors"
	"os"
	"os/exec"
)

func HasDifftastic() bool {
	_, err := exec.LookPath("difft")
	return err == nil
}

func GetDiffWithColor(repoPath string, staged bool) (string, error) {
	var args []string

	if HasDifftastic() {
		args = []string{"-c", "diff.external=difft", "diff", "--color=always"}
	} else {
		args = []string{"diff", "--color=always"}
	}

	if staged {
		args = append(args, "--staged")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_PAGER=cat", "DFT_COLOR=always")

	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return "", err
		}
		return "", err
	}

	return string(output), nil
}
