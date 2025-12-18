package commands

import "fmt"

// PrintInfo prints an informational message with blue [INFO] prefix.
func PrintInfo(msg string) {
	fmt.Printf("%s %s\n", InfoStyle.Render("[INFO]"), msg)
}

// PrintSuccess prints a success message with green [SUCCESS] prefix.
func PrintSuccess(msg string) {
	fmt.Printf("%s %s\n", SuccessStyle.Render("[SUCCESS]"), msg)
}

// PrintError prints an error message with red [ERROR] prefix.
func PrintError(msg string) {
	fmt.Printf("%s %s\n", ErrorStyle.Render("[ERROR]"), msg)
}

// PrintWarning prints a warning message with yellow [WARNING] prefix.
func PrintWarning(msg string) {
	fmt.Printf("%s %s\n", WarningStyle.Render("[WARNING]"), msg)
}
