package serve

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/arikui1911/memo/memo"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/browser"
	"github.com/shurcooL/github_flavored_markdown"
	"github.com/shurcooL/github_flavored_markdown/gfmstyle"
	"github.com/urfave/cli/v2"
)

var info = cli.Command{
	Name:    "serve",
	Aliases: []string{"s"},
	Usage:   "start http server",
	Action:  run,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "addr",
			Value: ":8080",
			Usage: "server address",
		},
	},
}

func init() {
	memo.RegisterCommand(&info, 80)
}

const templateDirContent = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Memo Life For You</title>
</head>
<style>
li {list-style-type: none;}
</style>
<body>
<ul>{{range .}}
  <li><a href="/{{.Name}}">{{.Name}}</a><dd>{{.Body}}</dd></li>{{end}}
</ul>
</body>
</html>
`

const templateBodyContent = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>{{.Name}}</title>
  <link href="/assets/gfm/gfm.css" media="all" rel="stylesheet" type="text/css" />
</head>
<body class="markdown-body">
{{.Body}}</body>
</html>
`

type entry struct {
	Name string
	Body template.HTML
}

func run(c *cli.Context) error {
	var cfg memo.Config
	err := cfg.Load()
	if err != nil {
		return err
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			f, err := os.Open(cfg.MemoDir)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer f.Close()
			files, err := f.Readdirnames(-1)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			files = memo.FilterMarkdown(files)
			var entries []entry
			for _, file := range files {
				entries = append(entries, entry{
					Name: file,
					Body: template.HTML(runewidth.Truncate(memo.ReadFileFirstLine(filepath.Join(cfg.MemoDir, file)), 80, "...")),
				})
			}
			w.Header().Set("content-type", "text/html")
			cfg.TemplateDirFile = memo.ExpandPath(cfg.TemplateDirFile)
			var t *template.Template
			if cfg.TemplateDirFile == "" {
				t = template.Must(template.New("dir").Parse(templateDirContent))
			} else {
				t, err = template.ParseFiles(cfg.TemplateDirFile)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			err = t.Execute(w, entries)
			if err != nil {
				log.Println(err)
			}
		} else {
			p := filepath.Join(cfg.MemoDir, memo.EscapeMemoTitle(req.URL.Path))
			b, err := ioutil.ReadFile(p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			body := string(b)
			if strings.HasPrefix(body, "---\n") {
				if pos := strings.Index(body[4:], "---\n"); pos > 0 {
					body = body[4+pos+4:]
				}
			}
			body = string(github_flavored_markdown.Markdown([]byte(body)))
			cfg.TemplateBodyFile = memo.ExpandPath(cfg.TemplateBodyFile)
			var t *template.Template
			if cfg.TemplateBodyFile == "" {
				t = template.Must(template.New("body").Parse(templateBodyContent))
			} else {
				t, err = template.ParseFiles(cfg.TemplateBodyFile)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			t.Execute(w, entry{
				Name: req.URL.Path,
				Body: template.HTML(body),
			})
		}
	})
	http.Handle("/assets/gfm/", http.StripPrefix("/assets/gfm", http.FileServer(gfmstyle.Assets)))
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir(cfg.AssetsDir))))

	addr := c.String("addr")
	var url string
	if strings.HasPrefix(addr, ":") {
		url = "http://localhost" + addr
	} else {
		url = "http://" + addr
	}
	browser.OpenURL(url)
	return http.ListenAndServe(addr, nil)
}
