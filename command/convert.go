package command

import (
	"fmt"
	"github.com/facebookgo/symwalk"
	"github.com/linxlib/logs"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"ink/article"
	"ink/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Convert(c *cli.Context) {
	// Parse arguments
	var sourcePath, rootPath string
	args := c.Args()
	if args.Len() > 0 {
		sourcePath = args.Slice()[0]
	} else {
		logs.Fatal("Please specify the posts source path")
	}
	if args.Len() > 1 {
		rootPath = args.Slice()[1]
	} else {
		rootPath = "."
	}
	// Check if path exist
	if !util.Exists(sourcePath) || !util.Exists(rootPath) {
		logs.Fatal("Please specify valid path")
	}
	// Parse Jekyll/Hexo post file
	count := 0
	symwalk.Walk(sourcePath, func(path string, f os.FileInfo, err error) error {
		fileExt := strings.ToLower(filepath.Ext(path))
		if fileExt == ".md" || fileExt == ".html" {
			// Read data from file
			data, err := ioutil.ReadFile(path)
			fileName := filepath.Base(path)
			logs.Info("Converting " + fileName)
			if err != nil {
				logs.Fatal(err)
			}
			// Split config and markdown
			var configStr, contentStr string
			content := strings.TrimSpace(string(data))
			parseAry := strings.SplitN(content, "---", 3)
			parseLen := len(parseAry)
			if parseLen == 3 { // Jekyll
				configStr = parseAry[1]
				contentStr = parseAry[2]
			} else if parseLen == 2 { // Hexo
				configStr = parseAry[0]
				contentStr = parseAry[1]
			}
			// Parse config
			var articleItem article.ArticleConfig
			if err = yaml.Unmarshal([]byte(configStr), &articleItem); err != nil {
				logs.Fatal(err)
			}
			tags := make(map[string]bool)
			for _, t := range articleItem.Tags {
				tags[t] = true
			}
			for _, c := range articleItem.Categories {
				if _, ok := tags[c]; !ok {
					articleItem.Tags = append(articleItem.Tags, c)
				}
			}
			if articleItem.Author == "" {
				articleItem.Author = "me"
			}
			// Convert date
			dateAry := strings.SplitN(articleItem.Date, ".", 2)
			if len(dateAry) == 2 {
				articleItem.Date = dateAry[0]
			}
			if len(articleItem.Date) == 10 {
				articleItem.Date = articleItem.Date + " 00:00:00"
			}
			if len(articleItem.Date) == 0 {
				articleItem.Date = "1970-01-01 00:00:00"
			}
			articleItem.Update = ""
			// Generate Config
			var inkConfig []byte
			if inkConfig, err = yaml.Marshal(articleItem); err != nil {
				logs.Fatal(err.Error())
			}
			inkConfigStr := string(inkConfig)
			markdownStr := inkConfigStr + "\n\n---\n\n" + contentStr + "\n"
			targetName := "source/" + fileName
			if fileExt != ".md" {
				targetName = targetName + ".md"
			}
			ioutil.WriteFile(filepath.Join(rootPath, targetName), []byte(markdownStr), 0644)
			count++
		}
		return nil
	})
	fmt.Printf("\nConvert finish, total %v articles\n", count)
}
