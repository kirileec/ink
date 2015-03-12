package main

import (
    "github.com/russross/blackfriday"
    "gopkg.in/yaml.v2"
    "html/template"
    "io/ioutil"
    "strings"
)

type SiteConfig struct {
    Title    string
    Subtitle string
    Logo     string
    Limit    int
    Theme    string
    Disqus   string
}

type AuthorConfig struct {
    Id     string
    Name   string
    Intro  string
    Avatar string
}

type BuildConfig struct {
    Port  string
    Watch bool
    Copy []string
    Publish string
}

type GlobalConfig struct {
    Site   SiteConfig
    Authors map[string]AuthorConfig
    Build  BuildConfig
}

type ArticleConfig struct {
    Title   string
    Date    string
    Author  string
    Tag     string
    Topic   string
    Preview string
}

type Article struct {
    ArticleConfig
    GlobalConfig
    Date    int64
    Author  AuthorConfig
    Tag     []string
    Preview string
    Content template.HTML
    Link    string
}

type Articles []Article

func (v Articles) Len() int           { return len(v) }
func (v Articles) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v Articles) Less(i, j int) bool { return v[i].Date > v[j].Date }

const (
    CONFIG_SPLIT = "---"
    MORE_SPLIT   = "---more---"
)

func parse(markdown string) template.HTML {
    // html.UnescapeString
    return template.HTML(blackfriday.MarkdownCommon([]byte(markdown)))
}

func ParseConfig(configPath string) *GlobalConfig {
    var config *GlobalConfig
    // Read data from file
    data, err := ioutil.ReadFile(configPath)
    if err != nil {
        Fatal(err.Error())
    }
    if err = yaml.Unmarshal(data, &config); err != nil {
        Fatal(err.Error())
    }
    return config
}

func ParseMarkdown(markdownPath string) *Article {
    var (
        config      *ArticleConfig
        configStr   string
        markdownStr string
    )
    // Read data from file
    data, err := ioutil.ReadFile(markdownPath)
    if err != nil {
        Fatal(err.Error())
    }
    // Split config and markdown
    contentStr := string(data)
    content := strings.SplitN(contentStr, CONFIG_SPLIT, 2)
    contentLen := len(content)
    if contentLen > 0 {
        configStr = content[0]
    }
    if contentLen > 1 {
        markdownStr = content[1]
    }
    // Parse config content
    if err := yaml.Unmarshal([]byte(configStr), &config); err != nil {
        Fatal(err.Error())
    }
    if config == nil {
        Fatal("Article config parse error")
    }
    var article Article
    // Parse preview splited by MORE_SPLIT
    previewAry := strings.SplitN(markdownStr, MORE_SPLIT, 2)
    if len(previewAry) > 1 {
        article.Preview = previewAry[0]
    } else {
        article.Preview = config.Preview
    }
    // Parse markdown content
    markdownStr = strings.Replace(markdownStr, MORE_SPLIT, "", 1)
    article.Content = parse(markdownStr)
    article.Date = ParseDate(config.Date).Unix()
    article.Title = config.Title
    if author, ok := globalConfig.Authors[config.Author]; ok {
        author.Id = config.Author
        article.Author = author
    }
    article.Tag = strings.Split(config.Tag, " ")
    article.Topic = config.Topic
    return &article
}