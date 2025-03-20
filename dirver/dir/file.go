package dir

import (
	"github.com/culionbear/configure"
	"github.com/fsnotify/fsnotify"
	"log"
)

type Driver struct {
	path    string
	watcher *fsnotify.Watcher
}

func New(path string) (*Driver, error) {
	return &Driver{
		path: path,
	}, nil
}

func (d *Driver) Name() string {
	return "dir"
}

func (d *Driver) Listen(ctx *configure.Context) (err error) {
	d.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	err = d.watcher.Add(d.path)
	if err != nil {
		_ = d.watcher.Close()
		return err
	}
	go func() {
		for {
			select {
			case event, ok := <-d.watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-d.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	return nil
}

func (d *Driver) Get(key string, schemes configure.KV) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Close() {
	_ = d.watcher.Close()
}
