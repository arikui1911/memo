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

	// "github.com/yuin/goldmark"
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

type serveConfig struct {
	memoDir          string
	templateDirFile  string
	templateBodyFile string
}

func run(c *cli.Context) error {
	var cfg memo.Config
	err := cfg.Load()
	if err != nil {
		return err
	}

	http.Handle("/", &serveConfig{
		memoDir:          cfg.MemoDir,
		templateDirFile:  cfg.TemplateDirFile,
		templateBodyFile: cfg.TemplateBodyFile,
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

func (c *serveConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		c.handleList(w, r)
		return
	}
	c.handleEntry(w, r)
}

type entry struct {
	Name string
	Body template.HTML
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

func (c *serveConfig) handleList(w http.ResponseWriter, req *http.Request) {
	files, err := getDirMarkdownFiles(c.memoDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	entries := makeEntries(files, c.memoDir)
	w.Header().Set("content-type", "text/html")
	t, err := prepareTemplate(memo.ExpandPath(c.templateDirFile), templateDirContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, entries)
	if err != nil {
		log.Println(err)
	}
}

func getDirMarkdownFiles(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	files, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	return memo.FilterMarkdown(files), nil
}

func makeEntries(files []string, dir string) []entry {
	entries := make([]entry, len(files))
	for i, file := range files {
		entries[i].Name = file
		entries[i].Body = template.HTML(runewidth.Truncate(memo.ReadFileFirstLine(filepath.Join(dir, file)), 80, "..."))
	}
	return entries
}

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

func (c *serveConfig) handleEntry(w http.ResponseWriter, req *http.Request) {
	src, err := readFile(filepath.Join(c.memoDir, memo.EscapeMemoTitle(req.URL.Path)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body := compileMarkdown(extractNotFrontMatter(src))
	t, err := prepareTemplate(c.templateBodyFile, templateBodyContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, entry{
		Name: req.URL.Path,
		Body: template.HTML(body),
	})
}

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func extractNotFrontMatter(src string) string {
	if !strings.HasPrefix(src, "---\n") {
		return src
	}
	if pos := strings.Index(src[4:], "---\n"); pos > 0 {
		return src[4+pos+4:]
	}
	return src
}

func compileMarkdown(src string) string {
	return string(github_flavored_markdown.Markdown([]byte(src)))
}

func prepareTemplate(filePath string, defaultSrc string) (*template.Template, error) {
	if len(filePath) == 0 {
		return template.Must(template.New("body").Parse(defaultSrc)), nil
	}
	return template.ParseFiles(memo.ExpandPath(filePath))
}
