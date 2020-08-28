# go-license

[![Build Status](https://travis-ci.com/spyzhov/go-license.svg?branch=master)](https://travis-ci.com/spyzhov/go-license)
[![Go Report Card](https://goreportcard.com/badge/github.com/spyzhov/go-license)](https://goreportcard.com/report/github.com/spyzhov/go-license)
[![GoDoc](https://godoc.org/github.com/spyzhov/go-license?status.svg)](https://godoc.org/github.com/spyzhov/go-license)
[![Coverage Status](https://coveralls.io/repos/github/spyzhov/go-license/badge.svg?branch=master)](https://coveralls.io/github/spyzhov/go-license?branch=master)

`go-license` is a tool to find all inherited licenses in the project, that uses `go modules`, based on `go.sum` file.

It could generate a report in any of the given formats: `Markdown`, `JSON`, `CSV`, `HTML`.

## Install

You can download a binary from the [Release page](https://github.com/spyzhov/go-license/releases), 
or install it from the source:

```
go get github.com/spyzhov/go-license
```

## Usage

`go-license` is a command line tool. 

```
Usage of go-license:
  -format string
        output format [csv, md, markdown, html, json] (default "json")
  -input string
        input path for go.sum file (default "go.sum")
  -output string
        output file name (default STDOUT)
  -v    verbose output
```

You can run it manually: 

```
go-license -format md > INHERITED_LICENSES.md
```

or, you can specify `go generate` command:

```go
// gen.go
package main
//go:generate go-license -format md -output INHERITED_LICENSES.md
```

## Examples

You can find [examples](examples) of running with the different output formats:
* `-format=md` or `-format=markdown` will generate [Markdown](examples/inherited_licenses.md) file;
* `-format=json` will generate [JSON](examples/inherited_licenses.json) file;
* `-format=csv` will generate [CSV](examples/inherited_licenses.csv) file;
* `-format=html` will generate [HTML](examples/inherited_licenses.html) file;

# License

MIT licensed. See the [LICENSE](LICENSE) file for details.
