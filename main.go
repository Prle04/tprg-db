package main

import (
	"tprgdb/cli"
)

func main() {
	app := cli.NewCLIApp() // Create a new CLI instance

	defer func() {
		app.Close()
	}()

	app.Run() // Run the CLI application

}
