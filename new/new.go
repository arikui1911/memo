package new

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/arikui1911/memo/memo"

	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

var info = cli.Command{
	Name:    "new",
	Aliases: []string{"n"},
	Usage:   "create memo",
	Action:  run,
}

func init() {
	memo.RegisterCommand(&info, 10)
}

const templateMemoContent = `# {{.Title}}
`

func run(c *cli.Context) error {
	var cfg memo.Config
	err := cfg.Load()
	if err != nil {
		return err
	}

	var title string
	var file string
	now := time.Now()
	if c.Args().Present() {
		title = c.Args().First()
		file = now.Format("2006-01-02-") + memo.EscapeMemoTitle(title) + ".md"
	} else {
		fmt.Print("Title: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return errors.New("canceled")
		}
		if scanner.Err() != nil {
			return scanner.Err()
		}
		title = scanner.Text()
		if title == "" {
			title = now.Format("2006-01-02")
			file = title + ".md"

		} else {
			file = now.Format("2006-01-02-") + memo.EscapeMemoTitle(title) + ".md"
		}
	}
	file = filepath.Join(cfg.MemoDir, file)
	if fileExists(file) {
		if !isatty.IsTerminal(os.Stdin.Fd()) {
			return copyFromStdin(file)
		}
		return cfg.RunCommand(cfg.Editor, "", file)
	}

	tmplString := templateMemoContent

	if fileExists(cfg.MemoTemplate) {
		b, err := ioutil.ReadFile(cfg.MemoTemplate)
		if err != nil {
			return err
		}
		tmplString = filterTmpl(string(b))
	}
	t := template.Must(template.New("memo").Parse(tmplString))

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	err = t.Execute(f, struct {
		Title, Date, Tags, Categories string
	}{
		title, now.Format("2006-01-02 15:04"), "", "",
	})
	f.Close()
	if err != nil {
		return err
	}

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return copyFromStdin(file)
	}
	return cfg.RunCommand(cfg.Editor, "", file)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

var filterReg = regexp.MustCompile(`{{_(.+?)_}}`)

func filterTmpl(tmpl string) string {
	return filterReg.ReplaceAllStringFunc(tmpl, func(substr string) string {
		m := filterReg.FindStringSubmatch(substr)
		return fmt.Sprintf("{{.%s}}", strings.Title(m[1]))
	})
}

func copyFromStdin(filename string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, os.Stdin)
	return err
}
