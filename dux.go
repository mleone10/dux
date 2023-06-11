// Package dux contains types which enable active monitoring and reloading of a program.
package dux

import (
	"context"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// TODO: Refactor this all to be testable...and add tests
// TODO: Add verbose mode (it should print a picture of a duck)

// A Watcher blocks until a trigger occurs or the context is canceled.
type Watcher interface {
	Watch(context.Context)
}

// A Closer is any type with a Close() method.
type Closer interface {
	Close()
}

// ExecEngine combines a command, arguments, and a watcher which triggers a reload of the command.
type ExecEngine struct {
	Cmd     string
	Args    []string
	Watcher Watcher
}

// Run executes ExecEngine.Cmd with ExecEngine.Args, then waits for ExecEngine.Watcher to unblock.  When it does, Run kills the command and reruns it.  This continues until the context is canceled.
// TODO: Stop Run when context is canceled
func (e ExecEngine) Run(ctx context.Context) {
	for {
		cancelCtx, cancelFn := context.WithCancel(ctx)
		runCommand(cancelCtx, e)
		e.Watcher.Watch(ctx)
		cancelFn()
	}
}

// runCommand kicks off the ExecEngine's Cmd and Args asynchronously, then stops it and any child processes when the context is canceled.
func runCommand(ctx context.Context, e ExecEngine) {
	// Don't use a CommandContext, as we want to do the killing ourselves.
	// Using a CommandContext _and_ killing with a syscall results in the first child process (cmd) being killed, then its children.
	// Since the parent isn't around to read the children's exit status, they become zombies.  Instead, manage it all manually.
	cmd := exec.Command(e.Cmd, e.Args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	go func() {
		// Issue a SIGKILL to the entire process group when the context is canceled.
		<-ctx.Done()
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}()

	go cmd.Run()
}

// A FuncEngine combines a function (which returns a Closer) with a Watcher which triggers that function's reload.
type FuncEngine struct {
	Func    func() Closer
	Watcher Watcher
}

// Run executes FuncEngine.Func, then waits for FuncEngine.Watcher to unblock.  When it does, Func's Close method is called and Func is rerun.
func (e FuncEngine) Run(ctx context.Context) {
	for {
		closeable := e.Func()
		e.Watcher.Watch(ctx)
		closeable.Close()
	}
}

// FileWatcher monitors a given file system, checking all files within at the given polling frequency.
type FileWatcher struct {
	FileSystem fs.FS
	PollFreq   time.Duration
}

// Watch monitors the FileWatcher file system for changes, blocking until either the context is canceled or a change is detected.
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

// TimeWatcher is a Watcher that triggers at a set interval.
type TimeWatcher struct {
	Delay time.Duration
}

// Watch blocks until TimeWatcher.Delay duration passes or the context is canceled.
func (t TimeWatcher) Watch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(t.Delay):
			return
		}
	}
}
