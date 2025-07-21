package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/root"
	"tprgdb/cli/parser"
)

type CLIApp struct {
	context commands.IContext
}

func NewCLIApp() *CLIApp {
	return &CLIApp{
		context: root.NewRootContext(), // Initialize the CLIApp with the root context
	}
}

func (cli *CLIApp) Run() {

	// defer cli.context.Close() // Ensure the context is closed when done

	reader := bufio.NewReader(os.Stdin) // Create a new reader from standard input

	for cli.context != nil {
		fmt.Print(cli.context.GetPrompt())   // Prompt the user
		line, err := reader.ReadString('\n') // Read until newline character
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		line = strings.TrimSpace(line)                    // Remove trailing newline and spaces
		parsedCommand, err := parser.ParseInputLine(line) // Parse the command
		if err != nil {
			fmt.Println("Error parsing command:", err)
			continue
		}
		command := cli.context.GetCommand(parsedCommand) // Get the command from the context
		if command == nil {
			fmt.Println("Unknown command:", parsedCommand.Command)
			continue
		}
		cli.context, err = command.Execute() // Execute the command
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}
}

func (cli *CLIApp) Close() error {
	if cli.context != nil {
		return cli.context.Close() // Close the current context
	}
	return nil
}
