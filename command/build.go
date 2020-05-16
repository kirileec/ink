package command

import (
	"fmt"
	"github.com/linxlib/logs"
	"ink/article"
	"sync"

	"ink/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/facebookgo/symwalk"
)

func Build() {
	startTime := time.Now()
	var articles = make(article.Collections, 0)
	var visibleArticles = make(article.Collections, 0)
	var pages = make(article.Collections, 0)
	var tagMap = make(map[string]article.Collections)
	var archiveMap = make(map[string]article.Collections)
	// Parse config
	article.ThemePath = filepath.Join(article.RootPath, article.Global.Site.Theme)
	article.PublicPath = filepath.Join(article.RootPath, article.Global.Build.Output)
	article.SourcePath = filepath.Join(article.RootPath, "source")
	// Append all partial html
	var partialTpl string
	files, _ := filepath.Glob(filepath.Join(article.ThemePath, "*.html"))
	for _, path := range files {
		fileExt := strings.ToLower(filepath.Ext(path))
		baseName := strings.ToLower(filepath.Base(path))
		if fileExt == ".html" && strings.HasPrefix(baseName, "_") {
			html, err := ioutil.ReadFile(path)
			if err != nil {
				logs.Fatal(err)
			}
			tplName := strings.TrimPrefix(baseName, "_")
			tplName = strings.TrimSuffix(tplName, ".html")
			htmlStr := "{{define \"" + tplName + "\"}}" + string(html) + "{{end}}"
			partialTpl += htmlStr
		}
	}
	// Compile template
	article.ArticleTpl = article.CompileTpl(filepath.Join(article.ThemePath, "article.html"), partialTpl, "article")
	article.PageTpl = article.CompileTpl(filepath.Join(article.ThemePath, "page.html"), partialTpl, "page")
	article.ArchiveTpl = article.CompileTpl(filepath.Join(article.ThemePath, "archive.html"), partialTpl, "archive")
	article.TagTpl = article.CompileTpl(filepath.Join(article.ThemePath, "tag.html"), partialTpl, "tag")
	// Clean public folder
	cleanPatterns := []string{"post", "tag", "images", "js", "css", "*.html", "favicon.ico", "robots.txt"}
	for _, pattern := range cleanPatterns {
		files, _ := filepath.Glob(filepath.Join(article.PublicPath, pattern))
		for _, path := range files {
			os.RemoveAll(path)
		}
	}
	// Find all .md to generate article
	symwalk.Walk(article.SourcePath, func(path string, info os.FileInfo, err error) error {
		fileExt := strings.ToLower(filepath.Ext(path))
		if fileExt == ".md" {
			// Parse markdown data
			articleItem := article.ParseArticle(path)
			if articleItem == nil || articleItem.Draft {
				return nil
			}
			logs.Info("Building " + articleItem.Link)
			// Generate file path
			directory := filepath.Dir(articleItem.Link)
			err := os.MkdirAll(filepath.Join(article.PublicPath, directory), 0777)
			if err != nil {
				logs.Fatal(err.Error())
			}
			// Append to collections
			if articleItem.Type == "page" {
				pages = append(pages, *articleItem)
				return nil
			}
			articles = append(articles, *articleItem)
			if articleItem.Hide {
				return nil
			}
			visibleArticles = append(visibleArticles, *articleItem)
			// Get tags info
			for _, tag := range articleItem.Tags {
				if _, ok := tagMap[tag]; !ok {
					tagMap[tag] = make(article.Collections, 0)
				}
				tagMap[tag] = append(tagMap[tag], *articleItem)
			}
			// Get archive info
			dateYear := articleItem.Time.Format("2006")
			if _, ok := archiveMap[dateYear]; !ok {
				archiveMap[dateYear] = make(article.Collections, 0)
			}
			articleInfo := article.ArticleInfo{
				DetailDate: articleItem.Date,
				Date:       articleItem.Time.Format("2006-01-02"),
				Title:      articleItem.Title,
				Link:       articleItem.Link,
				Top:        articleItem.Top,
			}
			archiveMap[dateYear] = append(archiveMap[dateYear], articleInfo)
		}
		return nil
	})
	if len(visibleArticles) == 0 {
		logs.Fatal("Must be have at least one article")
	}
	// Sort by date
	sort.Sort(articles)
	sort.Sort(visibleArticles)
	// Generate RSS page
	wg := sync.WaitGroup{}

	wg.Add(1)
	go article.GenerateRSS(visibleArticles, &wg)
	// Generate article list JSON
	wg.Add(1)
	go article.GenerateJSON(visibleArticles, &wg)
	// Render articles
	wg.Add(1)
	go article.RenderArticles(article.ArticleTpl, articles, &wg)
	// Render pages
	wg.Add(1)
	go article.RenderArticles(article.ArticleTpl, pages, &wg)
	// Generate article list pages
	wg.Add(1)
	go article.RenderArticleList("", visibleArticles, "", &wg)
	// Generate article list pages by tag
	for tagName, articles := range tagMap {
		sort.Sort(articles)
		wg.Add(1)
		go article.RenderArticleList(filepath.Join("tag", tagName), articles, tagName, &wg)
	}
	// Generate archive page
	archives := make(article.Collections, 0)
	for year, articleInfos := range archiveMap {
		// Sort by date
		sort.Sort(articleInfos)
		archives = append(archives, article.Archive{
			Year:     year,
			Articles: articleInfos,
		})
	}
	// Sort by year
	sort.Sort(archives)
	wg.Add(1)
	go article.RenderPage(article.ArchiveTpl, map[string]interface{}{
		"Total":   len(visibleArticles),
		"Archive": archives,
		"Site":    article.Global.Site,
		"I18n":    article.Global.I18n,
	}, filepath.Join(article.PublicPath, "archive.html"), &wg)
	// Generate tag page
	tags := make(article.Collections, 0)
	for tagName, tagArticles := range tagMap {
		articleInfos := make(article.Collections, 0)
		for _, item := range tagArticles {
			articleValue := item.(article.Article)
			articleInfos = append(articleInfos, article.ArticleInfo{
				DetailDate: articleValue.Date,
				Date:       articleValue.Time.Format("2006-01-02"),
				Title:      articleValue.Title,
				Link:       articleValue.Link,
				Top:        articleValue.Top,
			})
		}
		// Sort by date
		sort.Sort(articleInfos)
		tags = append(tags, article.Tag{
			Name:     tagName,
			Count:    len(tagArticles),
			Articles: articleInfos,
		})
	}
	// Sort by count
	sort.Sort(tags)
	wg.Add(1)
	go article.RenderPage(article.TagTpl, map[string]interface{}{
		"Total": len(visibleArticles),
		"Tag":   tags,
		"Site":  article.Global.Site,
		"I18n":  article.Global.I18n,
	}, filepath.Join(article.PublicPath, "tag.html"), &wg)
	// Generate other pages
	files, _ = filepath.Glob(filepath.Join(article.SourcePath, "*.html"))
	for _, path := range files {
		fileExt := strings.ToLower(filepath.Ext(path))
		baseName := filepath.Base(path)
		if fileExt == ".html" && !strings.HasPrefix(baseName, "_") {
			htmlTpl := article.CompileTpl(path, partialTpl, baseName)
			relPath, _ := filepath.Rel(article.SourcePath, path)
			wg.Add(1)
			go article.RenderPage(htmlTpl, article.Global, filepath.Join(article.PublicPath, relPath), &wg)
		}
	}
	// Copy static files
	copyStaticFile(&wg)
	wg.Wait()
	endTime := time.Now()
	usedTime := endTime.Sub(startTime)
	fmt.Printf("\nFinished to build in public folder (%v)\n", usedTime)
}

// Copy static files
func copyStaticFile(wg *sync.WaitGroup) {
	srcList := article.Global.Build.Copy
	for _, source := range srcList {
		if matches, err := filepath.Glob(filepath.Join(article.RootPath, source)); err == nil {
			for _, srcPath := range matches {
				logs.Info("Copying " + srcPath)
				file, err := os.Stat(srcPath)
				if err != nil {
					logs.Fatal("Not exist: " + srcPath)
				}
				fileName := file.Name()
				desPath := filepath.Join(article.PublicPath, fileName)
				wg.Add(1)
				if file.IsDir() {
					go util.CopyDir(srcPath, desPath, wg)
				} else {
					go util.CopyFile(srcPath, desPath, wg)
				}
			}
		} else {
			logs.Fatal(err.Error())
		}
	}
}
