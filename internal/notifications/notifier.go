package notifications

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Notifier sends system notifications
type Notifier interface {
	Send(title string, message string) error
}

// NewNotifier creates a platform-specific notifier
func NewNotifier() Notifier {
	switch runtime.GOOS {
	case "windows":
		return &WindowsNotifier{}
	case "darwin":
		return &MacNotifier{}
	case "linux":
		return &LinuxNotifier{}
	default:
		return &StubNotifier{} // Fallback for unknown OS
	}
}

// StubNotifier is a no-op notifier (used when proper notifiers can't be instantiated)
type StubNotifier struct{}

func (n *StubNotifier) Send(title string, message string) error {
	fmt.Printf("[NOTIFICATION] %s: %s\n", title, message)
	return nil
}

// WindowsNotifier sends notifications on Windows
type WindowsNotifier struct{}

func (n *WindowsNotifier) Send(title string, message string) error {
	// TODO: Implement using go-toast library
	// For now, use stub
	fmt.Printf("[WINDOWS NOTIFICATION] %s: %s\n", title, message)
	return nil
}

// MacNotifier sends notifications on macOS
type MacNotifier struct{}

func (n *MacNotifier) Send(title string, message string) error {
	if _, err := exec.LookPath("osascript"); err != nil {
		fmt.Printf("[MACOS NOTIFICATION] %s: %s\n", title, message)
		return nil
	}

	escapedTitle := strings.ReplaceAll(title, `"`, `\\"`)
	escapedMessage := strings.ReplaceAll(message, `"`, `\\"`)
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, escapedMessage, escapedTitle)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		fmt.Printf("[MACOS NOTIFICATION] %s: %s\n", title, message)
	}
	return nil
}

// LinuxNotifier sends notifications on Linux
type LinuxNotifier struct{}

func (n *LinuxNotifier) Send(title string, message string) error {
	if _, err := exec.LookPath("notify-send"); err == nil {
		cmd := exec.Command("notify-send", title, message)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	fmt.Printf("[LINUX NOTIFICATION] %s: %s\n", title, message)
	return nil
}
