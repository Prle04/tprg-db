package root

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/database"
	"tprgdb/engine"
)

type LoginCommand struct {
	caller   commands.IContext
	dbName   string
	username string
	password string
}

func (c *LoginCommand) Execute() (commands.IContext, error) {
	// We need to a database and return database context
	dbEngine, err := engine.Login(c.dbName, c.username, c.password)
	if err != nil {
		return c.caller, err
	}
	dbContext := database.NewDatabaseContext(c.caller, dbEngine)
	fmt.Printf("Logged in to database '%s' successfully.\n", c.dbName)
	return dbContext, nil
}

func NewLoginCommand(caller commands.IContext, dbName, username, password string) *LoginCommand {
	return &LoginCommand{
		caller:   caller,
		dbName:   dbName,
		username: username,
		password: password,
	}
}
