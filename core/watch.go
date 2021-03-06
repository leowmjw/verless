// Package watch provides verless' ability to watch a project
// and react to changes in a verless project.
package core

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
)

// watchContext provides all components required for serving an already built project.
type watchContext struct {
	Path        string
	IgnorePaths []string
	ChangedCh   chan<- string
	StopCh      <-chan bool
}

// watch watches a verless project for changes and writes the changed
// files to the passed watchContext.ChangedCh channel.
// To stop the watcher just close the watchContext.StopCh channel.
// watchContext.IgnorePath can be used to ignore a path inside the given watchContext.Path.
func watch(ctx watchContext) error {
	w := watcher.New()
	w.FilterOps(watcher.Write)

	go func() {
	watcherLoop:
		for {

			select {
			case event, ok := <-w.Event:
				if !ok {
					return
				}
				if filepath.Ext(event.Path) == "" {
					continue
				}

				for _, ignorePath := range ctx.IgnorePaths {
					p, err := filepath.Abs(ignorePath)
					if err != nil {
						log.Println(err)
						continue
					}
					if strings.HasPrefix(event.Path, p) {
						continue watcherLoop
					}
				}

				ctx.ChangedCh <- event.Path

			case err, ok := <-w.Error:
				if !ok {
					return
				}
				log.Println("watcher error:", err)

			case _, ok := <-ctx.StopCh:
				if !ok {
					w.Close()
				}
			}
		}
	}()

	if err := w.AddRecursive(ctx.Path); err != nil {
		return err
	}

	if err := w.Ignore(ctx.IgnorePaths...); err != nil {
		return err
	}

	var err error

	go func() {
		err = w.Start(time.Millisecond * 100)
	}()

	return err
}
