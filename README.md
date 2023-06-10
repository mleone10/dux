# dux
[![Go Reference](https://pkg.go.dev/badge/github.com/mleone10/dux.svg)](https://pkg.go.dev/github.com/mleone10/dux) [![Go Report Card](https://goreportcard.com/badge/github.com/mleone10/dux)](https://goreportcard.com/report/github.com/mleone10/dux)
> The tiny live reload tool

Dux monitors all files in a directory and reloads a given program if any of those files change.

## Usage
```
Usage of dux:
  -c string
    	command to execute
  -d string
    	directory to monitor for changes (default ".")
  -freq int
    	frequency at which the directory will be scanned, in seconds (default 1)
```
