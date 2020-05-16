package article

import (
	"github.com/BurntSushi/toml"
	"github.com/linxlib/logs"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var Global *GlobalConfig
var RootPath string

const DEFAULT_ROOT = "template"

// Parse config build
var (
	PageTpl    template.Template
	ArchiveTpl template.Template
	TagTpl     template.Template
	ArticleTpl template.Template
)
var (
	PublicPath string
	SourcePath string
	ThemePath  string
)

type SiteConfig struct {
	Root     string
	Title    string
	Subtitle string
	Logo     string
	Limit    int
	Theme    string
	Comment  string
	Lang     string
	Url      string
	Link     string
	Config   interface{}
}

type AuthorConfig struct {
	Id     string
	Name   string
	Intro  string
	Avatar string
}

type BuildConfig struct {
	Output  string
	Port    string
	Watch   bool
	Copy    []string
	Publish string
}

type GlobalConfig struct {
	I18n    map[string]string
	Site    SiteConfig
	Authors map[string]AuthorConfig
	Build   BuildConfig
	Develop bool
}

type ArticleConfig struct {
	Title      string
	Date       string
	Update     string
	Author     string
	Tags       []string
	Categories []string
	Topic      string
	Cover      string
	Draft      bool
	Preview    template.HTML
	Top        bool
	Type       string
	Hide       bool
	Config     interface{}
}

type ThemeConfig struct {
	Copy []string
	Lang map[string]map[string]string
}

func ParseGlobalConfigByCli(c *cli.Context, develop bool) {
	if c.Args().Len() > 0 {
		RootPath = c.Args().Slice()[0]
	} else {
		RootPath = "."
	}
	ParseGlobalConfigWrap(RootPath, develop)
	if Global == nil {
		ParseGlobalConfigWrap(DEFAULT_ROOT, develop)
		if Global == nil {
			logs.Fatal("Parse config.yml failed, please specify a valid path")
		}
	}
}

func ParseGlobalConfigWrap(root string, develop bool) {
	RootPath = root
	Global = ParseGlobalConfig(filepath.Join(RootPath, "config.yml"), develop)
	if Global == nil {
		return
	}
}

func ParseGlobalConfig(configPath string, develop bool) *GlobalConfig {
	var conf *GlobalConfig
	// Parse Global Config
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil
	}
	if err = yaml.Unmarshal(data, &conf); err != nil {
		logs.Fatal(err.Error())
	}
	if conf.Site.Config == nil {
		conf.Site.Config = ""
	}
	conf.Develop = develop
	if develop {
		conf.Site.Root = ""
	}
	conf.Site.Logo = strings.Replace(conf.Site.Logo, "-/", conf.Site.Root+"/", -1)
	if conf.Site.Url != "" && strings.HasSuffix(conf.Site.Url, "/") {
		conf.Site.Url = strings.TrimSuffix(conf.Site.Url, "/")
	}
	if conf.Build.Output == "" {
		conf.Build.Output = "public"
	}
	// Parse Theme Config
	themeConfig := ParseThemeConfig(filepath.Join(RootPath, conf.Site.Theme, "config.yml"))
	for _, copyItem := range themeConfig.Copy {
		conf.Build.Copy = append(conf.Build.Copy, filepath.Join(conf.Site.Theme, copyItem))
	}
	conf.I18n = make(map[string]string)
	for item, langItem := range themeConfig.Lang {
		conf.I18n[item] = langItem[conf.Site.Lang]
	}
	return conf
}

func ParseThemeConfig(configPath string) *ThemeConfig {
	// Read data from file
	var themeConfig *ThemeConfig
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		logs.Fatal(err)
	}
	// Parse config content
	if err := yaml.Unmarshal(data, &themeConfig); err != nil {
		logs.Fatal(err)
	}
	return themeConfig
}

func ParseArticleConfig(markdownPath string) (config *ArticleConfig, content string) {
	var configStr string
	// Read data from file
	data, err := ioutil.ReadFile(markdownPath)
	if err != nil {
		logs.Fatal(err)
	}
	// Split config and markdown
	contentStr := string(data)
	contentStr = replaceRootFlag(contentStr)
	var markdownStr []string
	contentLen := 0
	if strings.HasPrefix(contentStr, "+") {
		markdownStr = strings.SplitN(contentStr, CONFIG_SPLIT1, 3)
		contentLen = len(markdownStr)
	} else {
		markdownStr = strings.SplitN(contentStr, CONFIG_SPLIT, 3)
		contentLen = len(markdownStr)
	}

	if contentLen > 0 {
		configStr = markdownStr[1]
	}
	if contentLen > 1 {
		content = markdownStr[2]
	}
	// Parse config content
	if err := yaml.Unmarshal([]byte(configStr), &config); err != nil {
		if err = toml.Unmarshal([]byte(configStr), &config); err != nil {
			logs.Error(err)
			return nil, ""
		}

	}
	if config == nil {
		return nil, ""
	}
	if config.Type == "" {
		config.Type = "post"
	}
	// Parse preview splited by MORE_SPLIT
	previewAry := strings.SplitN(content, MORE_SPLIT, 2)
	if len(config.Preview) <= 0 && len(previewAry) > 1 {
		config.Preview = parseMarkdown(previewAry[0])
		content = strings.Replace(content, MORE_SPLIT, "", 1)
	} else {
		config.Preview = parseMarkdown(string(config.Preview))
	}
	return config, content
}
