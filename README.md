# go-license

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
