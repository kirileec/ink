package article

import (
	"github.com/linxlib/logs"
	"html/template"
	"ink/util"
	"path/filepath"
	"strings"
	"time"

	"ink/blackfriday"
)

type Article struct {
	GlobalConfig
	ArticleConfig
	Time     time.Time
	MTime    time.Time
	Date     int64
	Update   int64
	Author   AuthorConfig
	Category string
	Tags     []string
	Markdown string
	Preview  template.HTML
	Content  template.HTML
	Link     string
	Config   interface{}
}

const (
	CONFIG_SPLIT  = "---"
	CONFIG_SPLIT1 = "+++"
	MORE_SPLIT    = "<!--more-->"
)

func parseMarkdown(markdown string) template.HTML {
	// html.UnescapeString
	return template.HTML(blackfriday.MarkdownCommon([]byte(markdown)))
}

func replaceRootFlag(content string) string {
	return strings.Replace(content, "-/", Global.Site.Root+"/", -1)
}

func ParseArticle(markdownPath string) *Article {
	conf, content := ParseArticleConfig(markdownPath)
	if conf == nil {
		logs.Error("Invalid format: " + markdownPath)
		return nil
	}
	if conf.Config == nil {
		conf.Config = ""
	}
	var articleItem Article
	// Parse markdown content
	articleItem.Hide = conf.Hide
	articleItem.Type = conf.Type
	articleItem.Preview = conf.Preview
	articleItem.Config = conf.Config
	articleItem.Markdown = content
	articleItem.Content = parseMarkdown(content)
	if conf.Date != "" {
		articleItem.Time = util.ParseDate(conf.Date)
		articleItem.Date = articleItem.Time.Unix()
	}
	if conf.Update != "" {
		articleItem.MTime = util.ParseDate(conf.Update)
		articleItem.Update = articleItem.MTime.Unix()
	}
	articleItem.Title = conf.Title
	articleItem.Topic = conf.Topic
	articleItem.Draft = conf.Draft
	articleItem.Top = conf.Top
	if author, ok := Global.Authors[conf.Author]; ok {
		author.Id = conf.Author
		author.Avatar = replaceRootFlag(author.Avatar)
		articleItem.Author = author
	}
	if len(conf.Categories) > 0 {
		articleItem.Category = conf.Categories[0]
	} else {
		articleItem.Category = "misc"
	}
	tags := map[string]bool{}
	articleItem.Tags = conf.Tags
	for _, tag := range conf.Tags {
		tags[tag] = true
	}
	for _, cat := range conf.Categories {
		if _, ok := tags[cat]; !ok {
			articleItem.Tags = append(articleItem.Tags, cat)
		}
	}
	// Support topic and cover field
	if conf.Cover != "" {
		articleItem.Cover = conf.Cover
	} else {
		articleItem.Cover = conf.Topic
	}
	// Generate page name
	fileName := strings.TrimSuffix(strings.ToLower(filepath.Base(markdownPath)), ".md")
	link := fileName + ".html"
	// Genetate custom link
	if articleItem.Type == "post" {
		datePrefix := articleItem.Time.Format("2006-01-02-")
		if strings.HasPrefix(fileName, datePrefix) {
			fileName = fileName[len(datePrefix):]
		}
		if Global.Site.Link != "" {
			linkMap := map[string]string{
				"{year}":     articleItem.Time.Format("2006"),
				"{month}":    articleItem.Time.Format("01"),
				"{day}":      articleItem.Time.Format("02"),
				"{hour}":     articleItem.Time.Format("15"),
				"{minute}":   articleItem.Time.Format("04"),
				"{second}":   articleItem.Time.Format("05"),
				"{category}": articleItem.Category,
				"{title}":    fileName,
			}
			link = Global.Site.Link
			for key, val := range linkMap {
				link = strings.Replace(link, key, val, -1)
			}
		}
	}
	articleItem.Link = link
	articleItem.GlobalConfig = *Global
	return &articleItem
}
