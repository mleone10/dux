package dux_test

import (
	"testing"
)

func TestFileWatcher(t *testing.T) {
	// fs := fstest.MapFS{
	// 	"fileName": &fstest.MapFile{
	// 		ModTime: time.Now(),
	// 	},
	// }
	// var changed bool

	// go func() {
	// 	dux.FileWatcher{fs, time.Millisecond * 200}.Watch(context.Background())
	// 	changed = true
	// }()

	// fs["fileName"].ModTime = time.Now().Add(time.Hour * -1)

	// time.Sleep(time.Millisecond * 500)
	// if !changed {
	// 	t.Errorf("Expected modTime to be detected, but it was not")
	// }
}
