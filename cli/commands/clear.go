package commands

import (
	"os"
	"os/exec"
	"runtime"
)

type ClearCommand struct {
	caller IContext
}

func (c *ClearCommand) Execute() (IContext, error) {
	clearConsole()
	return c.caller, nil
}

func NewClearCommand(caller IContext) *ClearCommand {
	return &ClearCommand{
		caller: caller,
	}
}

func clearConsole() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default: // Linux, macOS, etc.
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
