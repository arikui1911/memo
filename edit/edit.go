package edit

import (
	"path/filepath"

	"github.com/arikui1911/memo/memo"

	"github.com/urfave/cli/v2"
)

var info = cli.Command{
	Name:    "edit",
	Aliases: []string{"e"},
	Usage:   "edit memo",
	Action:  run,
}

func init() {
	memo.RegisterCommand(&info, 30)
}

func run(c *cli.Context) error {
	var cfg memo.Config
	err := cfg.Load()
	if err != nil {
		return err
	}

	var files []string
	if c.Args().Present() {
		files = append(files, filepath.Join(cfg.MemoDir, c.Args().First()))
	} else {
		files, err = cfg.FilterFiles()
		if err != nil {
			return err
		}
	}
	return cfg.RunCommand(cfg.Editor, "", files...)
}
