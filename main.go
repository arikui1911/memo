package main

import (
	"fmt"
	"os"

	_ "github.com/arikui1911/memo/config"
	_ "github.com/arikui1911/memo/delete"
	_ "github.com/arikui1911/memo/edit"
	_ "github.com/arikui1911/memo/grep"
	_ "github.com/arikui1911/memo/list"
	"github.com/arikui1911/memo/memo"
	_ "github.com/arikui1911/memo/new"
	_ "github.com/arikui1911/memo/serve"
	_ "github.com/arikui1911/memo/view"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = memo.Name
	app.Usage = "Memo Life For You"
	app.Version = memo.Version
	app.Action = action
	memo.RegisterCommandsToApplication(app)

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
	os.Exit(0)
}

func action(c *cli.Context) error {
	args := c.Args()
	if !args.Present() {
		cli.ShowAppHelp(c)
		memo.ShowPluginHelps()
		return nil
	}
	return memo.ExecPlugin(args.First(), args.Tail())
}
