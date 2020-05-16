package command

import (
	"bufio"
	"github.com/linxlib/logs"
	"github.com/urfave/cli/v2"
	"html/template"
	"ink/util"
	"os"
	"strings"
	"time"
)

const (
	DATE_FORMAT_STRING = util.DATE_FORMAT_WITH_TIMEZONE
	INDENT             = "  " // 2 spaces
	POST_TEMPLATE      = `---
title: {{.Title}}
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
---
`
)

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
	date := time.Now()

	// Empty string values
	preview := ""
	cover := ""

	// Parse args
	args := c.Args()
	// ink new blog title  -> source/post/blog-title.md
	if args.Len() > 0 {
		blogTitle = strings.Join(args.Slice(), " ")
	}
	if blogTitle == "" {
		if c.String("title") != "" {
			blogTitle = c.String("title")
		} else {
			logs.Fatal("Please specify the name of the blog post")
		}
	}

	fileName = strings.ReplaceAll(blogTitle, " ", "-") + ".md"
	if c.String("file") != "" {
		fileName = c.String("file")
	}

	author = "linx"

	if c.Bool("post") && c.Bool("page") {
		logs.Fatal("The post and page arguments are mutually exclusive and cannot appear together")
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
	if postType == "post" {
		filePath = "source/post/" + fileName
	}

	file, err := os.Create(filePath)
	if err != nil {
		logs.Fatal(err)
	}
	postTemplate, err := template.New("post").Parse(POST_TEMPLATE)
	if err != nil {
		logs.Fatal(err)
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
			logs.Fatal("Illegal date string")
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
		"Preview":    preview,
		"Cover":      cover,
		"Tags":       tagString,
	}
	fileWriter := bufio.NewWriter(file)
	err = postTemplate.Execute(fileWriter, data)
	if err != nil {
		logs.Fatal(err)
	}
	err = fileWriter.Flush()
	if err != nil {
		logs.Fatal(err)
	}
}
