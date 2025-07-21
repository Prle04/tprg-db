package root

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/cli/commands/database"
	"tprgdb/engine"
)

type NewDBCommand struct {
	caller   commands.IContext
	dbName   string
	username string
	password string
}

func (c *NewDBCommand) Execute() (commands.IContext, error) {
	// We need to create a new database and return database context
	dbEngine, err := engine.NewKVDatabase(c.dbName, c.username, c.password)
	if err != nil {
		return c.caller, err
	}
	dbContext := database.NewDatabaseContext(c.caller, dbEngine)

	fmt.Printf("Database '%s' created successfully.\n", c.dbName)
	return dbContext, nil
}

func NewNewDBCommand(caller commands.IContext, dbName, username, password string) *NewDBCommand {
	return &NewDBCommand{
		caller:   caller,
		dbName:   dbName,
		username: username,
		password: password,
	}
}
