package list

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/arikui1911/memo/memo"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/mattn/go-runewidth"
	"github.com/urfave/cli/v2"
)

var info = cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "list memo",
	Action:  run,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "fullpath",
			Usage: "show file path",
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "print the result using a Go template `string`",
		},
	},
}

func init() {
	memo.RegisterCommand(&info, 20)
}

func run(c *cli.Context) error {
	var cfg memo.Config
	err := cfg.Load()
	if err != nil {
		return err
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
	istty := isatty.IsTerminal(os.Stdout.Fd())
	col := cfg.Column
	if col == 0 {
		col = memo.DefaultColumn
	}
	pat := c.Args().First()

	var tmpl *template.Template
	if format := c.String("format"); format != "" {
		t, err := template.New("format").Parse(format)
		if err != nil {
			return err
		}
		tmpl = t
	}

	fullpath := c.Bool("fullpath")
	for _, file := range files {
		if pat != "" && !strings.Contains(file, pat) {
			continue
		}
		if tmpl != nil {
			var b bytes.Buffer
			err := tmpl.Execute(&b, map[string]interface{}{
				"File":     file,
				"Title":    memo.ReadFileFirstLine(filepath.Join(cfg.MemoDir, file)),
				"Fullpath": filepath.Join(cfg.MemoDir, file),
			})
			if err != nil {
				return err
			}
			fmt.Println(b.String())
		} else if istty && !fullpath {
			wi := cfg.Width
			if wi == 0 {
				wi = memo.DefaultWidth
			}
			title := runewidth.Truncate(memo.ReadFileFirstLine(filepath.Join(cfg.MemoDir, file)), wi-4-col, "...")
			file = runewidth.FillRight(runewidth.Truncate(file, col, "..."), col)
			fmt.Fprintf(color.Output, "%s : %s\n", color.GreenString(file), color.YellowString(title))
		} else {
			if fullpath {
				file = filepath.Join(cfg.MemoDir, file)
			}
			fmt.Println(file)
		}
	}
	return nil
}
