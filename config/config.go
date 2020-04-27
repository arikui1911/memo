package config

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/arikui1911/memo/memo"
	"github.com/urfave/cli/v2"
)

var info = cli.Command{
	Name:    "config",
	Aliases: []string{"c"},
	Usage:   "configure",
	Action:  run,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "cat",
			Usage: "cat the file",
		},
	},
}

func init() {
	memo.RegisterCommand(&info, 70)
}

func run(c *cli.Context) error {
	var cfg memo.Config
	err := cfg.Load()
	if err != nil {
		return err
	}

	dir := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		dir = os.Getenv("APPDATA")
		if dir == "" {
			dir = filepath.Join(os.Getenv("USERPROFILE"), "Application Data", "memo")
		}
		dir = filepath.Join(dir, "memo")
	} else {
		dir = filepath.Join(dir, ".config", "memo")
	}
	file := filepath.Join(dir, "config.toml")
	if c.Bool("cat") {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(os.Stdout, f)
		return err
	}

	return cfg.RunCommand(cfg.Editor, "", file)
}
