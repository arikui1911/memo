package grep

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/arikui1911/memo/memo"

	"github.com/urfave/cli/v2"
)

var info = cli.Command{
	Name:    "grep",
	Aliases: []string{"g"},
	Usage:   "grep memo",
	Action:  run,
}

func init() {
	memo.RegisterCommand(&info, 60)
}

func run(c *cli.Context) error {
	var cfg memo.Config
	err := cfg.Load()
	if err != nil {
		return err
	}

	if !c.Args().Present() {
		return errors.New("pattern required")
	}
	f, err := os.Open(cfg.MemoDir)
	if err != nil {
		return err
	}
	defer f.Close()
	files, err := f.Readdirnames(-1)
	if err != nil || len(files) == 0 {
		return err
	}
	files = memo.FilterMarkdown(files)
	var args []string
	for _, file := range files {
		args = append(args, filepath.Join(cfg.MemoDir, file))
	}
	return cfg.RunCommand(cfg.GrepCmd, c.Args().First(), args...)
}
