// Package dux contains types which enable active monitoring and reloading of a program.
package dux

import (
	"context"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

// TODO: Refactor this all to be testable...and add tests
// TODO: Add verbose mode

// A Watcher blocks until a trigger occurs or the context is canceled.
type Watcher interface {
	Watch(context.Context)
}

// An Engine combines a command, arguments, and a watcher which triggers a reload of the command.
type Engine struct {
	Cmd     string
	Args    []string
	Watcher Watcher
}

// Run executes e.Cmd with arguments e.Args, then waits for e.Watcher to unblock.  When it does, Run kills the command and reruns it.  This continues until the context is cancelled.
// TODO: Stop Run when context is canceled
func (e Engine) Run(ctx context.Context) {
	for {
		cancelCtx, cancelFn := context.WithCancel(ctx)
		runCommand(cancelCtx, e)
		e.Watcher.Watch(ctx)
		cancelFn()
	}
}

func runCommand(ctx context.Context, e Engine) {
	cmd := exec.CommandContext(ctx, e.Cmd, e.Args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	go cmd.Run()
}

// FileWatcher monitors a given file system, checking all files within at the given polling frequency.
type FileWatcher struct {
	FileSystem fs.FS
	PollFreq   time.Duration
}

// Watch monitors the FileWatcher file system for changes, blocking until either the context is cancelled or a change is detected.
func (fw FileWatcher) Watch(ctx context.Context) {
	if fw.FileSystem == nil {
		fw.FileSystem = os.DirFS(".")
	}

	if fw.PollFreq == time.Nanosecond*0 {
		fw.PollFreq = time.Second
	}

	blockUntilChange(fw, ctx)
}

func blockUntilChange(fw FileWatcher, ctx context.Context) {
	var wg sync.WaitGroup
	cancelCtx, cancelFn := context.WithCancel(ctx)

	// Create a watcher goroutine for each file in the file system
	fs.WalkDir(fw.FileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		wg.Add(1)

		go func() {
			defer wg.Done()

			fInfo, err := d.Info()
			if err != nil {
				return
			}

			for {
				select {
				case <-cancelCtx.Done():
					return
				case <-time.After(fw.PollFreq):
					nInfo, err := d.Info()
					if err != nil {
						log.Printf("%+v", err)
						return
					}

					if fInfo.ModTime() != nInfo.ModTime() {
						cancelFn()
						return
					}
				}
			}
		}()

		return nil
	})

	// TODO: Watch for new files
	// TODO: Watch for deleted files

	// Wait for all goroutines to finish
	wg.Wait()
}
