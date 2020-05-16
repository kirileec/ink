package main

import (
	"github.com/urfave/cli/v2"
	"ink/article"
	"ink/command"
	"os"
)

const (
	VERSION = "RELEASE 2020-05-16"
)

func main() {

	app := cli.NewApp()
	app.Name = "ink"
	app.Usage = "An elegant static blog generator"
	app.Authors = []*cli.Author{
		{Name: "Harrison", Email: "harrison@lolwut.com"},
		{Name: "Oliver Allen", Email: "oliver@toyshop.com"},
	}
	//app.Email = "imeoer@gmail.com"
	app.Version = VERSION
	app.Commands = []*cli.Command{
		{
			Name:  "build",
			Usage: "构建博客到public目录",
			Action: func(c *cli.Context) error {
				article.ParseGlobalConfigByCli(c, false)
				command.Build()
				return nil
			},
		},
		{
			Name:  "preview",
			Usage: "预览博客",
			Action: func(c *cli.Context) error {
				article.ParseGlobalConfigByCli(c, true)
				command.Serve(true)
				return nil
			},
		},
		{
			Name:  "publish",
			Usage: "发布博客",
			Action: func(c *cli.Context) error {
				article.ParseGlobalConfigByCli(c, false)
				command.Publish()
				return nil
			},
		},
		{
			Name:  "serve",
			Usage: "运行博客",
			Action: func(c *cli.Context) error {
				article.ParseGlobalConfigByCli(c, true)
				command.Serve(false)
				return nil
			},
		},
		{
			Name:  "convert",
			Usage: "Convert Jekyll/Hexo post format to Ink format (Beta)",
			Action: func(c *cli.Context) error {
				command.Convert(c)
				return nil
			},
		},
		{
			Name:  "new",
			Usage: "创建文章",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "hide",
					Usage: "Hides the article",
				},
				&cli.BoolFlag{
					Name:  "top",
					Usage: "Places the article at the top",
				},
				&cli.BoolFlag{
					Name:  "post",
					Usage: "The article is a post",
				},
				&cli.BoolFlag{
					Name:  "page",
					Usage: "The article is a page",
				},
				&cli.BoolFlag{
					Name:  "draft",
					Usage: "The article is a draft",
				},

				&cli.StringFlag{
					Name:  "title",
					Usage: "Article title",
				},
				&cli.StringFlag{
					Name:  "author",
					Usage: "Article author",
				},
				&cli.StringFlag{
					Name:  "cover",
					Usage: "Article cover path",
				},
				&cli.StringFlag{
					Name:  "date",
					Usage: "The date and time on which the article was created (2006-01-02 15:04:05)",
				},
				&cli.StringFlag{
					Name:  "file",
					Usage: "The path of where the article will be stored",
				},

				&cli.StringSliceFlag{
					Name:  "tag",
					Usage: "Adds a tag to the article",
				},
			},
			Action: func(c *cli.Context) error {
				command.New(c)
				return nil
			},
		},
	}
	app.Run(os.Args)
	os.Exit(1)
}
