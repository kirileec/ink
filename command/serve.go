package command

import (
	"github.com/linxlib/logs"
	"ink/article"
	"os"
	"path/filepath"

	"github.com/facebookgo/symwalk"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/taadis/oper"
	"ink/ink.go"
)

var watcher *fsnotify.Watcher
var conn *websocket.Conn

func Watch() {
	// Listen watched file change event
	if watcher != nil {
		watcher.Close()
	}
	watcher, _ = fsnotify.NewWatcher()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op == fsnotify.Write {
					// Handle when file change
					logs.Info(event.Name)
					article.ParseGlobalConfigWrap(article.RootPath, true)
					Build()
					if conn != nil {
						if err := conn.WriteMessage(websocket.TextMessage, []byte("change")); err != nil {
							logs.Warn(err)
						}
					}
				}
			case err := <-watcher.Errors:
				logs.Warn(err)
			}
		}
	}()
	var dirs = []string{
		filepath.Join(article.RootPath, "source"),
		filepath.Join(article.ThemePath, "bundle"),
	}
	var files = []string{
		filepath.Join(article.RootPath, "config.yml"),
		filepath.Join(article.ThemePath),
	}
	for _, source := range dirs {
		symwalk.Walk(source, func(path string, f os.FileInfo, err error) error {
			if f.IsDir() {
				if err := watcher.Add(path); err != nil {
					logs.Warn(err.Error())
				}
			}
			return nil
		})
	}
	for _, source := range files {
		if err := watcher.Add(source); err != nil {
			logs.Warn(err.Error())
		}
	}
}

func Websocket(ctx *ink.Context) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	if c, err := upgrader.Upgrade(ctx.Res, ctx.Req, nil); err != nil {
		logs.Warn(err)
	} else {
		conn = c
	}
	ctx.Stop()
}

func Serve(watch bool) {
	Build()
	if watch {
		Watch()
	}
	previewWeb := ink.New()
	previewWeb.Get("/live", Websocket)
	previewWeb.Get("*", ink.Static(filepath.Join(article.RootPath, article.Global.Build.Output)))

	uri := "http://localhost:" + article.Global.Build.Port + "/"
	logs.Info("Access " + uri + " to open preview")
	oper.Access(uri)
	previewWeb.Listen(":" + article.Global.Build.Port)
}
