package simhash

import (
	"fmt"
	"tprgdb/cli/commands"
	"tprgdb/cli/parser"
	"tprgdb/engine"
)

type SimHashContext struct {
	owner    commands.IContext
	dbEngine *engine.KVEngine
}

func (ctx *SimHashContext) GetCommand(cmd *parser.ParsedCommand) commands.ICommand {
	switch cmd.Command {
	case "new":
		if value, exists := cmd.Flags["-f"]; exists {
			if len(cmd.Args) != 1 {
				return commands.NewErrorCommand(ctx, fmt.Errorf("usage: new <name> -f <filename>"))
			}
			if len(value) != 1 {
				return commands.NewErrorCommand(ctx, fmt.Errorf("usage: new <name> -f <filename>"))
			}
			return NewNewCommand(ctx, ctx.dbEngine, cmd.Args[0], value[0])
		}
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: new <name> <text>"))
		}
		return NewNewCommand(ctx, ctx.dbEngine, cmd.Args[0], cmd.Args[1])
	case "get":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: get <name>"))
		}
		return NewGetCommand(ctx, ctx.dbEngine, cmd.Args[0])
	case "similarity":
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: similarity <name1> <name2>"))
		}
		return NewSimilarityCommand(ctx, ctx.dbEngine, cmd.Args[0], cmd.Args[1])
	case "delete":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: delete <name>"))
		}
		return NewDeleteCommand(ctx, ctx.dbEngine, cmd.Args[0])

	case "exit":
		return commands.NewBackCommand(ctx)
	default:
		return nil
	}
}

func (ctx *SimHashContext) GetOwner() commands.IContext {
	return ctx.owner
}

func (ctx *SimHashContext) GetPrompt() string {
	return fmt.Sprintf("%s@%s (BF)> ", ctx.dbEngine.GetUsername(), ctx.dbEngine.GetDBName())
}

func (ctx *SimHashContext) Close() error {
	// The owner is database context, so we close it
	ctx.dbEngine = nil
	return ctx.owner.Close()
}

func NewSimHashContext(owner commands.IContext, dbEngine *engine.KVEngine) commands.IContext {
	return &SimHashContext{
		owner:    owner,
		dbEngine: dbEngine,
	}
}
