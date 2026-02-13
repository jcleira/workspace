package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptYesNo displays a prompt and returns true if user responds y/yes.
// Defaults to no if the user just presses enter.
func PromptYesNo(prompt string) bool {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}

// PromptYesNoDefault displays a prompt with a configurable default.
// If defaultYes is true, pressing enter without input returns true.
func PromptYesNoDefault(prompt string, defaultYes bool) bool {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}
