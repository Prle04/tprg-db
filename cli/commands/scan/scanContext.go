package scan

import (
	"fmt"
	"strconv"
	"tprgdb/cli/commands"
	"tprgdb/cli/parser"
	"tprgdb/engine"
	engineScan "tprgdb/engine/api/scans"
)

// Assuming ICommand is defined in commands package
type ScanContext struct {
	owner    commands.IContext
	cursor   engineScan.ICursor
	dbEngine *engine.KVEngine
	scanType string
}

// Implement GetCommand
func (ctx *ScanContext) GetCommand(cmd *parser.ParsedCommand) commands.ICommand {
	switch cmd.Command {
	case "page":
		if len(cmd.Args) != 1 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("usage: page <pageNumber>"))
		}
		pageNumber, err := strconv.Atoi(cmd.Args[0])
		if err != nil || pageNumber <= 0 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("pageNumber must be a positive integer"))
		}
		return NewPageCommand(ctx, ctx.cursor, pageNumber)
	case "page-next":
		return NewPageNextCommand(ctx, ctx.cursor)
	case "page-prev":
		if ctx.cursor.GetCurrentPage() <= 1 {
			return commands.NewErrorCommand(ctx, fmt.Errorf("already at the first page"))
		}
		return NewPagePrevCommand(ctx, ctx.cursor)
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
func (ctx *ScanContext) GetOwner() commands.IContext {
	return ctx.owner
}

// Implement GetPrompt
func (ctx *ScanContext) GetPrompt() string {
	// Example prompt
	return fmt.Sprintf("%s@%s (%s)> ", ctx.dbEngine.GetUsername(), ctx.dbEngine.GetDBName(), ctx.scanType)
}

// Implement Close
func (ctx *ScanContext) Close() error {
	// The owner is database context, so we close it
	ctx.dbEngine = nil
	return ctx.owner.Close()
}

func NewScanContext(owner commands.IContext, dbEngine *engine.KVEngine, cursor engineScan.ICursor, scanType string) *ScanContext {
	return &ScanContext{
		owner:    owner,
		cursor:   cursor,
		dbEngine: dbEngine,
		scanType: scanType,
	}
}
