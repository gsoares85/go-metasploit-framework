package main

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
)

var console *readline.Instance

func initConsole() {
	rl, err := readline.New("go-msf1 >")
	if err != nil {
		fmt.Println("Error initializing console:", err)
		panic(err)
	}
	console = rl
}

func main() {
	initConsole()

	for {
		line, err := console.Readline()
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}
		args := strings.Fields(line)
		if len(args) > 0 {
			handleCommand(args)
		}
	}
}

func handleCommand(args []string) {
	switch args[0] {
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help - Display this help message")
	fmt.Println("  workspace - Manage workspaces")
}
