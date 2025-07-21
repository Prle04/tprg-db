package database

import (
	"fmt"
	"strconv"
	"tprgdb/cli/commands"
	"tprgdb/cli/parser"
	"tprgdb/engine"
)

type DatabaseContext struct {
	owner    commands.IContext
	dbEngine *engine.KVEngine
}

func NewDatabaseContext(owner commands.IContext, dbEngine *engine.KVEngine) *DatabaseContext {
	return &DatabaseContext{
		owner:    owner,
		dbEngine: dbEngine,
	}
}

func (rc *DatabaseContext) GetCommand(cmd *parser.ParsedCommand) commands.ICommand {
	switch cmd.Command {
	case "logout":
		return NewLogoutCommand(rc)

	case "get":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(rc, fmt.Errorf("usage: get <key>"))
		}
		if value, exists := cmd.Flags["-f"]; exists {
			if len(value) != 1 {
				return commands.NewErrorCommand(rc, fmt.Errorf("usage: get <key> -f <filepath>"))
			}
			return NewGetToFileCommand(rc, rc.dbEngine, cmd.Args[0], value[0])
		}
		return NewGetCommand(rc, rc.dbEngine, cmd.Args[0])
	case "get-many":
		if len(cmd.Args) < 1 {
			return commands.NewErrorCommand(rc, fmt.Errorf("usage: get-many <key1> <key2> ... <keyN>"))
		}
		if value, exists := cmd.Flags["-f"]; exists {
			if len(value) != len(cmd.Args) {
				return commands.NewErrorCommand(rc, fmt.Errorf("usage: get-many <key1> <key2> ... <keyN> -f <file1> <file2> ... <fileN>"))
			}
			return NewGetManyFileCommand(rc, rc.dbEngine, cmd.Args, value)
		}
		return NewGetManyCommand(rc, rc.dbEngine, cmd.Args)
	case "set":
		if value, exists := cmd.Flags["-f"]; exists {
			if len(value) != 1 {
				return commands.NewErrorCommand(rc, fmt.Errorf("usage: set <key> -f <filepath>"))
			}
			if len(cmd.Args) != 1 {
				return commands.NewErrorCommand(rc, fmt.Errorf("usage: set <key> -f <filepath>"))
			}
			return NewSetFileCommand(rc, rc.dbEngine, cmd.Args[0], value[0])
		}
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(rc, fmt.Errorf("usage: set <key> <value>"))
		}
		return NewSetCommand(rc, rc.dbEngine, cmd.Args[0], []byte(cmd.Args[1]))
	case "delete":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(rc, fmt.Errorf("usage: delete <key>"))
		}
		return NewDeleteCommand(rc, rc.dbEngine, cmd.Args[0])
	case "validate":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(rc, fmt.Errorf("usage: validate <tableName>"))
		}
		return NewValidateCommand(rc, rc.dbEngine, cmd.Args[0])
	case "prefix-scan":
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(rc, fmt.Errorf("usage: prefix-scan <prefix> <pageSize>"))
		}
		pageSize, err := strconv.Atoi(cmd.Args[1])
		if err != nil || pageSize <= 0 {
			return commands.NewErrorCommand(rc, fmt.Errorf("pageSize must be a positive integer"))
		}
		return NewPrefixScanCommand(rc, rc.dbEngine, cmd.Args[0], pageSize)
	case "range-scan":
		if len(cmd.Args) != 3 {
			return commands.NewErrorCommand(rc, fmt.Errorf("usage: range-scan <startKey> <endKey> <pageSize>"))
		}
		pageSize, err := strconv.Atoi(cmd.Args[2])
		if err != nil || pageSize <= 0 {
			return commands.NewErrorCommand(rc, fmt.Errorf("pageSize must be a positive integer"))
		}
		return NewRangeScanCommand(rc, rc.dbEngine, cmd.Args[0], cmd.Args[1], pageSize)
	case "prefix-iter":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(rc, fmt.Errorf("usage: prefix-iter <prefix>"))
		}
		return NewPrefixIterCommand(rc, rc.dbEngine, cmd.Args[0])
	case "range-iter":
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(rc, fmt.Errorf("usage: range-iter <startKey> <endKey>"))
		}
		return NewRangeIterCommand(rc, rc.dbEngine, cmd.Args[0], cmd.Args[1])
	case "bf":
		return NewBFCommand(rc, rc.dbEngine)
	case "hll":
		return NewHLLCommand(rc, rc.dbEngine)
	case "cms":
		return NewCMSCommand(rc, rc.dbEngine)
	case "sh":
		return NewSHCommand(rc, rc.dbEngine)
	case "clear":
		return commands.NewClearCommand(rc)
	case "help":
		return NewHelpCommand(rc)
	default:
		return nil
	}
}

func (rc *DatabaseContext) GetPrompt() string {
	return fmt.Sprintf("%s@%s> ", rc.dbEngine.GetUsername(), rc.dbEngine.GetDBName())
}

func (rc *DatabaseContext) Close() error {
	if rc.dbEngine == nil {
		return nil
	}
	err := rc.dbEngine.Shutdown()
	if err != nil {
		return err
	}
	rc.dbEngine = nil
	return rc.owner.Close()
}

func (rc *DatabaseContext) GetOwner() commands.IContext {
	return rc.owner
}
