package database

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/engine"
	"tprgdb/engine/api/user"
)

type NewUserCommand struct {
	caller   commands.IContext
	dbEngine *engine.KVEngine
	username string
	password string
}

func (cmd *NewUserCommand) Execute() (commands.IContext, error) {
	newUser, err := user.NewUser(cmd.username, cmd.password)
	if err != nil {
		return nil, err
	}
	if err := cmd.dbEngine.SetUser(newUser); err != nil {
		return nil, err
	}
	fmt.Println("New user created successfully: " + newUser.Username)
	return cmd.caller, nil
}

func NewNewUserCommand(caller commands.IContext, dbEngine *engine.KVEngine, username, password string) *NewUserCommand {
	return &NewUserCommand{
		caller:   caller,
		dbEngine: dbEngine,
		username: username,
		password: password,
	}
}
