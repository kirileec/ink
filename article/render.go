package article

import (
	"encoding/json"
	"github.com/gorilla/feeds"
	"github.com/linxlib/logs"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Data interface{}

type RenderArticle struct {
	Article
	Next *Article
	Prev *Article
}

// Compile html template
func CompileTpl(tplPath string, partialTpl string, name string) template.Template {
	// Read template data from file
	html, err := ioutil.ReadFile(tplPath)
	if err != nil {
		logs.Fatal(err.Error())
	}
	// Append partial template
	htmlStr := string(html) + partialTpl
	funcMap := template.FuncMap{
		"i18n": func(val string) string {
			return Global.I18n[val]
		},
	}
	// Generate html content
	tpl, err := template.New(name).Funcs(funcMap).Parse(htmlStr)
	if err != nil {
		logs.Fatal(err)
	}
	return *tpl
}

// Render html file by data
func RenderPage(tpl template.Template, tplData interface{}, outPath string, wg *sync.WaitGroup) {
	// Create file
	outFile, err := os.Create(outPath)
	if err != nil {
		logs.Fatal(err)
	}
	defer func() {
		outFile.Close()
	}()
	defer wg.Done()
	// Template render
	err = tpl.Execute(outFile, tplData)
	if err != nil {
		logs.Fatal(err)
	}
}

// Generate all article page
func RenderArticles(tpl template.Template, articles Collections, wg *sync.WaitGroup) {
	defer wg.Done()
	articleCount := len(articles)
	for i, _ := range articles {
		currentArticle := articles[i].(Article)
		var renderArticle = RenderArticle{currentArticle, nil, nil}
		if i >= 1 {
			article := articles[i-1].(Article)
			renderArticle.Prev = &article
			if i <= articleCount-2 {
				article := articles[i+1].(Article)
				renderArticle.Next = &article
			}
			outPath := filepath.Join(PublicPath, currentArticle.Link)
			wg.Add(1)
			go RenderPage(tpl, renderArticle, outPath, wg)
		}
	}
}

// Generate rss page
func GenerateRSS(articles Collections, wg *sync.WaitGroup) {
	defer wg.Done()
	var feedArticles Collections
	if len(articles) < Global.Site.Limit {
		feedArticles = articles
	} else {
		feedArticles = articles[0:Global.Site.Limit]
	}
	if Global.Site.Url != "" {
		feed := &feeds.Feed{
			Title:       Global.Site.Title,
			Link:        &feeds.Link{Href: Global.Site.Url},
			Description: Global.Site.Subtitle,
			Author:      &feeds.Author{Global.Site.Title, ""},
			Created:     time.Now(),
		}
		feed.Items = make([]*feeds.Item, 0)
		for _, item := range feedArticles {
			articleItem := item.(Article)
			feed.Items = append(feed.Items, &feeds.Item{
				Title:       articleItem.Title,
				Link:        &feeds.Link{Href: Global.Site.Url + "/" + articleItem.Link},
				Description: string(articleItem.Preview),
				Author:      &feeds.Author{articleItem.Author.Name, ""},
				Created:     articleItem.Time,
				Updated:     articleItem.MTime,
			})
		}
		if atom, err := feed.ToAtom(); err == nil {
			err := ioutil.WriteFile(filepath.Join(PublicPath, "atom.xml"), []byte(atom), 0644)
			if err != nil {
				logs.Fatal(err.Error())
			}
		} else {
			logs.Fatal(err.Error())
		}
	}
}

// Generate article list page
func RenderArticleList(rootPath string, articles Collections, tagName string, wg *sync.WaitGroup) {
	defer wg.Done()
	// Create path
	pagePath := filepath.Join(PublicPath, RootPath)
	os.MkdirAll(pagePath, 0777)
	// Split page
	limit := Global.Site.Limit
	total := len(articles)
	page := total / limit
	rest := total % limit
	if rest != 0 {
		page++
	}
	if total < limit {
		page = 1
	}
	for i := 0; i < page; i++ {
		var prev = filepath.Join(RootPath, "page"+strconv.Itoa(i)+".html")
		var next = filepath.Join(RootPath, "page"+strconv.Itoa(i+2)+".html")
		outPath := filepath.Join(pagePath, "index.html")
		if i != 0 {
			fileName := "page" + strconv.Itoa(i+1) + ".html"
			outPath = filepath.Join(pagePath, fileName)
		} else {
			prev = ""
		}
		if i == 1 {
			prev = filepath.Join(RootPath, "index.html")
		}
		first := i * limit
		count := first + limit
		if i == page-1 {
			if rest != 0 {
				count = first + rest
			}
			next = ""
		}
		var data = map[string]interface{}{
			"Articles": articles[first:count],
			"Site":     Global.Site,
			"Develop":  Global.Develop,
			"Page":     i + 1,
			"Total":    page,
			"Prev":     prev,
			"Next":     next,
			"TagName":  tagName,
			"TagCount": len(articles),
		}
		wg.Add(1)
		go RenderPage(PageTpl, data, outPath, wg)
	}
}

// Generate article list JSON
func GenerateJSON(articles Collections, wg *sync.WaitGroup) {
	defer wg.Done()
	datas := make([]map[string]interface{}, 0)
	for i, _ := range articles {
		articleItem := articles[i].(Article)
		var data = map[string]interface{}{
			"title":   articleItem.Title,
			"content": articleItem.Markdown,
			"preview": string(articleItem.Preview),
			"link":    articleItem.Link,
			"cover":   articleItem.Cover,
		}
		datas = append(datas, data)
	}
	str, _ := json.Marshal(datas)
	ioutil.WriteFile(filepath.Join(PublicPath, "index.json"), []byte(str), 0644)
}
