package main

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"golang.org/x/text/encoding/simplifiedchinese"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/facebookgo/symwalk"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

const (
	VERSION            = "RELEASE 2018-07-27"
	DEFAULT_ROOT       = "blog"
	DATE_FORMAT_STRING = "2006-01-02 15:04:05"
	INDENT             = "  " // 2 spaces
	POST_TEMPLATE      = `title: {{.Title}}
date: {{.DateString}}
author: {{.Author}}
{{- if .Cover}}
cover: {{.Cover}}
{{- end}}
draft: {{.Draft}}
top: {{.Top}}
{{- if .Preview}}
preview: {{.Preview}}
{{- end}}
{{- if .Tags}}
{{.Tags}}
{{- end}}
type: {{.Type}}
hide: {{.Hide}}
toc: {{.Toc}}
---
`
)

var globalConfig *GlobalConfig
var themeConfig *ThemeConfig
var rootPath string

func ConvertByte2String(byte []byte, charset Charset) string {

	var str string
	switch charset {
	case GB18030:
		decodeBytes, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}

	return str
}

func main() {

	app := cli.NewApp()
	app.Name = "ink"
	app.Usage = "静态博客生成器"
	app.Authors = []*cli.Author{
		{Name: "linx", Email: "sulinke1133@gmail.com"},
	}
	//app.Email = "imeoer@gmail.com"
	app.Version = VERSION
	app.Commands = []*cli.Command{
		{
			Name:  "build",
			Usage: "构建静态页面到public目录",
			Action: func(c *cli.Context) error {
				ParseGlobalConfigByCli(c, false)
				Build()
				return nil
			},
		},
		{
			Name:  "preview",
			Usage: "预览博客",
			Action: func(c *cli.Context) error {
				ParseGlobalConfigByCli(c, true)
				Build()
				Watch()
				Serve()
				return nil
			},
		},
		{
			Name:  "publish",
			Usage: "发布博客",
			Action: func(c *cli.Context) error {
				ParseGlobalConfigByCli(c, false)
				Build()
				Publish()
				return nil
			},
		},
		{
			Name:  "serve",
			Usage: "服务模式",
			Action: func(c *cli.Context) error {
				ParseGlobalConfigByCli(c, true)
				Build()
				Serve()
				return nil
			},
		},
		{
			Name:  "convert",
			Usage: "转换 Jekyll/Hexo 格式到 Ink 格式 (Beta)",
			Action: func(c *cli.Context) error {
				Convert(c)
				return nil
			},
		},
		{
			Name:  "new",
			Usage: "创建新文章",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "hide",
					Usage: "隐藏文章",
				},
				&cli.BoolFlag{
					Name:  "toc",
					Usage: "文章目录",
				},
				&cli.BoolFlag{
					Name:  "top",
					Usage: "置顶",
				},
				&cli.BoolFlag{
					Name:  "post",
					Usage: "文章类型",
				},
				&cli.BoolFlag{
					Name:  "page",
					Usage: "页面类型",
				},
				&cli.BoolFlag{
					Name:  "draft",
					Usage: "草稿",
				},

				&cli.StringFlag{
					Name:  "title",
					Usage: "标题",
				},
				&cli.StringFlag{
					Name:  "author",
					Usage: "作者",
				},
				&cli.StringFlag{
					Name:  "cover",
					Usage: "文章封面",
				},
				&cli.StringFlag{
					Name:  "date",
					Usage: "创建日期",
				},
				&cli.StringFlag{
					Name:  "file",
					Usage: "文件路径",
				},

				&cli.StringSliceFlag{
					Name:  "tag",
					Usage: "文章标签",
				},
			},
			Action: func(c *cli.Context) error {
				New(c)
				return nil
			},
		},
	}
	app.Run(os.Args)
	os.Exit(exitCode)
}

func ParseGlobalConfigByCli(c *cli.Context, develop bool) {
	if c.Args().Len() > 0 {
		rootPath = c.Args().Slice()[0]
	} else {
		rootPath = "."
	}
	ParseGlobalConfigWrap(rootPath, develop)
	if globalConfig == nil {
		ParseGlobalConfigWrap(DEFAULT_ROOT, develop)
		if globalConfig == nil {
			Fatal("Parse config.yml failed, please specify a valid path")
		}
	}
}

func ParseGlobalConfigWrap(root string, develop bool) {
	rootPath = root
	globalConfig, themeConfig = ParseGlobalConfig(filepath.Join(rootPath, "config.yml"), develop)
	if globalConfig == nil || themeConfig == nil {
		return
	}
}

func New(c *cli.Context) {
	// If source folder does not exist, create
	if _, err := os.Stat("source/"); os.IsNotExist(err) {
		os.Mkdir("source", os.ModePerm)
	}

	var author, blogTitle, fileName string
	var tags []string

	// Default values
	draft := "false"
	top := "false"
	postType := "post"
	hide := "false"
	toc := "false"
	date := time.Now()

	// Empty string values
	preview := ""
	cover := ""

	// Parse args
	args := c.Args()
	if args.Len() > 0 {
		blogTitle = args.Slice()[0]
	}
	if blogTitle == "" {
		if c.String("title") != "" {
			blogTitle = c.String("title")
		} else {
			Fatal("Please specify the name of the blog post")
		}
	}

	fileName = blogTitle + ".md"
	if c.String("file") != "" {
		fileName = c.String("file")
	}

	if args.Len() > 1 {
		author = args.Slice()[1]
	}
	if author == "" {
		author = c.String("author")
	}

	if c.Bool("post") && c.Bool("page") {
		Fatal("The post and page arguments are mutually exclusive and cannot appear together")
	}
	if c.Bool("post") {
		postType = "post"
	}
	if c.Bool("page") {
		postType = "page"
	}
	if c.Bool("hide") {
		hide = "true"
	}
	if c.Bool("toc") {
		toc = "true"
	}
	if c.Bool("draft") {
		draft = "true"
	}
	if c.Bool("top") {
		top = "true"
	}

	if c.String("preview") != "" {
		preview = c.String("preview")
	}
	if c.String("cover") != "" {
		cover = c.String("cover")
	}

	var filePath = "source/" + fileName
	file, err := os.Create(filePath)
	if err != nil {
		Fatal(err)
	}
	postTemplate, err := template.New("post").Parse(POST_TEMPLATE)
	if err != nil {
		Fatal(err)
	}

	if c.StringSlice("tag") != nil {
		tags = c.StringSlice("tag")
	}

	var tagString string
	if len(tags) > 0 {
		tagString = "tags:"
		for _, tag := range tags {
			tagString += "\n" + INDENT + "- " + tag
		}
	}

	var dateString string
	if c.String("date") != "" {
		dateString = c.String("date")
		_, err = time.Parse(DATE_FORMAT_STRING, dateString)
		if err != nil {
			Fatal("Illegal date string")
		}
	} else {
		dateString = date.Format(DATE_FORMAT_STRING)
	}
	data := map[string]string{
		"Title":      blogTitle,
		"DateString": dateString,
		"Author":     author,
		"Draft":      draft,
		"Top":        top,
		"Type":       postType,
		"Hide":       hide,
		"Toc":        toc,
		"Preview":    preview,
		"Cover":      cover,
		"Tags":       tagString,
	}
	fileWriter := bufio.NewWriter(file)
	err = postTemplate.Execute(fileWriter, data)
	if err != nil {
		Fatal(err)
	}
	err = fileWriter.Flush()
	if err != nil {
		Fatal(err)
	}
}

func Publish() {
	command := globalConfig.Build.Publish
	commandWin := globalConfig.Build.PublishW
	// Prepare exec command
	var shell, flag string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
		command = commandWin
	} else {
		shell = "/bin/sh"
		flag = "-c"
	}
	cmd := exec.Command(shell, flag, command)
	cmd.Dir = filepath.Join(rootPath, globalConfig.Build.Output)
	// Start print stdout and stderr of process
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	out := bufio.NewScanner(stdout)
	err := bufio.NewScanner(stderr)
	// Print stdout
	go func() {
		for out.Scan() {
			if runtime.GOOS == "windows" {
				garbledStr := ConvertByte2String(out.Bytes(), GB18030)
				Log(garbledStr)
			} else {
				Log(out.Text())
			}

		}
	}()
	// Print stdin
	go func() {
		for err.Scan() {
			if runtime.GOOS == "windows" {
				garbledStr := ConvertByte2String(out.Bytes(), GB18030)
				Log(garbledStr)
			} else {
				Log(err.Text())
			}
		}
	}()
	// Exec command
	cmd.Run()
}

func Convert(c *cli.Context) {
	// Parse arguments
	var sourcePath, rootPath string
	args := c.Args()
	if args.Len() > 0 {
		sourcePath = args.Slice()[0]
	} else {
		Fatal("Please specify the posts source path")
	}
	if args.Len() > 1 {
		rootPath = args.Slice()[1]
	} else {
		rootPath = "."
	}
	// Check if path exist
	if !Exists(sourcePath) || !Exists(rootPath) {
		Fatal("Please specify valid path")
	}
	// Parse Jekyll/Hexo post file
	count := 0
	symwalk.Walk(sourcePath, func(path string, f os.FileInfo, err error) error {
		fileExt := strings.ToLower(filepath.Ext(path))
		if fileExt == ".md" || fileExt == ".html" {
			// Read data from file
			data, err := os.ReadFile(path)
			fileName := filepath.Base(path)
			Log("转换中 " + fileName)
			if err != nil {
				Fatal(err.Error())
			}
			// Split config and markdown
			var configStr, contentStr string
			content := strings.TrimSpace(string(data))
			var sep = "---"
			if strings.Contains(string(data), "+++") {
				sep = "+++"
			}
			parseAry := strings.SplitN(content, sep, 3)
			parseLen := len(parseAry)
			if parseLen == 3 { // Jekyll
				configStr = parseAry[1]
				contentStr = parseAry[2]
			} else if parseLen == 2 { // Hexo
				configStr = parseAry[0]
				contentStr = parseAry[1]
			}
			// Parse config
			var article ArticleConfig
			if sep == "+++" {
				if err = toml.Unmarshal([]byte(configStr), &article); err != nil {
					Fatal(err.Error())
				}
			} else {
				if err = yaml.Unmarshal([]byte(configStr), &article); err != nil {
					Fatal(err.Error())
				}
			}

			tags := make(map[string]bool)
			for _, t := range article.Tags {
				tags[t] = true
			}
			for _, c := range article.Categories {
				if _, ok := tags[c]; !ok {
					article.Tags = append(article.Tags, c)
				}
			}
			if article.Author == "" {
				article.Author = "me"
			}
			// Convert date
			//"2006-01-02T15:04:05Z07:00"
			tf1 := "2006-01-02T15:04:05Z08:00"
			tf2 := "2006-01-02 15:04:05"
			tf3 := "2006-01-02 15:04:05 +0800"
			tf4 := "2006-01-02"
			if t, err := time.Parse(tf1, article.Date); err != nil {
				if _, err := time.Parse(tf2, article.Date); err != nil {
					if _, err := time.Parse(tf3, article.Date); err != nil {
						if _, err := time.Parse(tf4, article.Date); err != nil {
						} else {
							article.Date += " 00:00:00"
						}
					} else {

					}
				} else {

				}
			} else {
				article.Date = t.Format(tf3)
			}

			article.Update = ""
			// Generate Config
			var inkConfig []byte
			if inkConfig, err = yaml.Marshal(article); err != nil {
				Fatal(err.Error())
			}
			inkConfigStr := string(inkConfig)
			markdownStr := inkConfigStr + "\n\n---\n\n" + contentStr + "\n"
			targetName := fileName
			if fileExt != ".md" {
				targetName = targetName + ".md"
			}

			os.WriteFile(filepath.Join(rootPath, targetName), []byte(markdownStr), 0644)
			count++
		}
		return nil
	})
	fmt.Printf("\nConvert finish, total %v articles\n", count)
}
