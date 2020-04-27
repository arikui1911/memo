package view

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arikui1911/memo/memo"

	"github.com/urfave/cli/v2"
)

var info = cli.Command{
	Name:    "view",
	Aliases: []string{"v"},
	Usage:   "view memo",
	Action:  run,
}

func init() {
	memo.RegisterCommand(&info, 40)
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

	for i, file := range files {
		if i > 0 {
			// Print new page
			fmt.Println("\x12")
		}
		err = catFile(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func catFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	return scanner.Err()
}
