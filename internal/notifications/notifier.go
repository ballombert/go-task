package notifications

import (
	"fmt"
	"runtime"
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
	// TODO: Implement using osascript or NSUserNotificationCenter
	// For now, use stub
	fmt.Printf("[MACOS NOTIFICATION] %s: %s\n", title, message)
	return nil
}

// LinuxNotifier sends notifications on Linux
type LinuxNotifier struct{}

func (n *LinuxNotifier) Send(title string, message string) error {
	// TODO: Implement using d-bus or notify-send
	// For now, use stub
	fmt.Printf("[LINUX NOTIFICATION] %s: %s\n", title, message)
	return nil
}
