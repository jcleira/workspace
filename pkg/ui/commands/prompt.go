package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptYesNo displays a prompt and returns true if user responds y/yes.
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
