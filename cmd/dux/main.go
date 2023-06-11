package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mleone10/dux"
)

func main() {
	fDir := flag.String("d", ".", "directory to monitor for changes")
	fCmd := flag.String("c", "", "linux shell command to execute")
	fFreq := flag.Int("freq", 1, "frequency at which the directory will be scanned, in seconds")
	flag.Parse()

	if *fCmd == "" {
		fmt.Println("Missing required command parameter ('-c').  Try again with a '-c' argument containing a valid shell command.")
		os.Exit(1)
	}

	cmdSlice := strings.Fields(*fCmd)
	fSys := os.DirFS(*fDir)

	dux.ExecEngine{
		Cmd:  cmdSlice[0],
		Args: cmdSlice[1:],
		Watcher: dux.FileWatcher{
			FileSystem: fSys,
			PollFreq:   time.Second * time.Duration(*fFreq),
		},
	}.Run(context.Background())
}
