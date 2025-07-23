package hyperloglog

import (
	"fmt"
	"strconv"
	"tprgdb/cli/commands"
	"tprgdb/cli/parser"
	"tprgdb/engine"
)

type HyperLogLogContext struct {
	owner    commands.IContext
	dbEngine *engine.KVEngine
}

func (ctx *HyperLogLogContext) GetCommand(cmd *parser.ParsedCommand) commands.ICommand {
	switch cmd.Command {
	case "new":
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: new <name> <bits>"))
		}
		bits, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			return commands.NewErrorCommand(ctx, fmt.Errorf("invalid bits: %s", cmd.Args[1]))
		}
		return NewNewCommand(ctx, ctx.dbEngine, cmd.Args[0], bits)
	case "delete":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: delete <name>"))
		}
		return NewDeleteCommand(ctx, ctx.dbEngine, cmd.Args[0])
	case "insert":
		if len(cmd.Args) != 2 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: insert <name> <value>"))
		}
		return NewInsertCommand(ctx, ctx.dbEngine, cmd.Args[0], []byte(cmd.Args[1]))
	case "count":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: count <name> "))
		}
		return NewCountCommand(ctx, ctx.dbEngine, cmd.Args[0])
	case "exit":
		return commands.NewBackCommand(ctx)
	case "clear":
		return commands.NewClearCommand(ctx)
	case "help":
		return NewHelpCommand(ctx)
	default:
		return nil
	}
}

func (ctx *HyperLogLogContext) GetOwner() commands.IContext {
	return ctx.owner
}

func (ctx *HyperLogLogContext) GetPrompt() string {
	return fmt.Sprintf("%s@%s (HLL)> ", ctx.dbEngine.GetUsername(), ctx.dbEngine.GetDBName())
}

func (ctx *HyperLogLogContext) Close() error {
	// The owner is database context, so we close it
	ctx.dbEngine = nil
	return ctx.owner.Close()
}

func NewHyperLogLogContext(owner commands.IContext, dbEngine *engine.KVEngine) *HyperLogLogContext {
	return &HyperLogLogContext{
		owner:    owner,
		dbEngine: dbEngine,
	}
}
