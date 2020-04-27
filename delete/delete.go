package delete

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arikui1911/memo/memo"

	"github.com/fatih/color"
	"github.com/mattn/go-tty"
	"github.com/urfave/cli/v2"
)

var info = cli.Command{
	Name:    "delete",
	Aliases: []string{"d"},
	Usage:   "delete memo",
	Action:  run,
}

func init() {
	memo.RegisterCommand(&info, 50)
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
	if err != nil {
		return err
	}
	files = memo.FilterMarkdown(files)
	pat := c.Args().First()
	var args []string
	for _, file := range files {
		if pat != "" && !strings.Contains(file, pat) {
			continue
		}
		fmt.Println(file)
		args = append(args, filepath.Join(cfg.MemoDir, file))
	}
	if len(args) == 0 {
		color.Yellow("%s", "No matched entry")
		return nil
	}
	color.Red("%s", "Will delete those entry. Are you sure?")
	answer, err := ask("Are you sure? (y/N)")
	if answer == false || err != nil {
		return err
	}
	answer, err = ask("Really? (y/N)")
	if answer == false || err != nil {
		return err
	}
	for _, arg := range args {
		err = os.Remove(arg)
		if err != nil {
			return err
		}
		color.Yellow("Deleted: %v", arg)
	}
	return nil
}

func ask(prompt string) (bool, error) {
	fmt.Print(prompt + ": ")
	t, err := tty.Open()
	if err != nil {
		return false, err
	}
	defer t.Close()
	var r rune
	for r == 0 {
		r, err = t.ReadRune()
		if err != nil {
			return false, err
		}
	}
	fmt.Println()
	return r == 'y' || r == 'Y', nil
}
