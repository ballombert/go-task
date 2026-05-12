package main

import (
	"fmt"
	"os"

	"github.com/beaallombert/gotask/internal/cli"
	"github.com/beaallombert/gotask/internal/tui"
)

func main() {
	app, err := cli.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if len(os.Args) < 2 {
		// No arguments - launch TUI
		if err := tui.Start("inbox.md", os.ExpandEnv("$HOME/.gotask/gotask.db")); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	command := os.Args[1]

	switch command {
	case "help":
		printHelp()

	case "tui":
		if err := tui.Start("inbox.md", os.ExpandEnv("$HOME/.gotask/gotask.db")); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "inbox":
		if err := app.HandleInboxCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "timer":
		if err := app.HandleTimerCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "focus":
		if err := app.HandleFocusCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
	}
}

func printHelp() {
	help := `
gotask - Task management for ADHD productivity

Usage: gotask [command] [args]

If no command is provided, launches the interactive TUI.

Commands:
  tui                    Launch interactive terminal UI
  
  inbox <subcommand>     Manage tasks in inbox
    top                    Show top 3 tasks
    add --description "..."  Add a new task
    start --line N         Mark task as in progress
    pause --line N         Pause a task
    complete --line N      Mark task as completed

  timer <subcommand>     Manage timers and sessions
    start --preset "Name"  Start a timer with preset
    stop                   Stop active timer
    status                 Show timer status
    presets                List available presets

  focus <subcommand>     Check focus state
    snapshot               Show current focus snapshot

  help                    Show this help message
`
	fmt.Println(help)
}
