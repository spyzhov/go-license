package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type testExampleMap map[string]string

var examples = testExampleMap{
	"MIT": `github.com/spyzhov/ajson v0.2.1 h1:Za4nWtiEcTp5w2R3+vJx/L/uCQMSAHpZhPNA4zlMAkc=
github.com/spyzhov/ajson v0.2.1/go.mod h1:63V+CGM6f1Bu/p4nLIN8885ojBdt88TbLoSFzyqMuVA=`,
	"Unknown":    `gitlab.com/spyzhov/private v0.2.1 h1:Za4nWtiEcTp5w2R3+vJx/L/uCQMSAHpZhPNA4zlMAkc=`,
	"Apache-2.0": `gopkg.in/yaml.v2 v2.2.2/go.mod h1:hI93XBmqTisBFMUTm0b8Fm+jr3Dg1NNxqwp+5A1VGuI=`,
	"Apache-2.0, MIT": `gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c h1:dUUwHk2QECo/6vqA44rthZ8ie2QXMNeKRTHCNY2nXvo=
gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=`,
	"BSD-2-Clause": `gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 h1:qIbj1fsPNlZgppZ+VLlY7N33q108Sa+fhmuc+sWQYwY=
gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=`,
}

func TestMain_JSON(t *testing.T) {
	request := strings.Join([]string{
		examples["MIT"],
		examples["Unknown"],
		"broken",
		examples["BSD-2-Clause"],
		examples["Apache-2.0"],
		examples["Apache-2.0, MIT"],
		"",
	}, "\n")
	content, err := runTest(t, request, "json")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	value := make(map[string]map[string][]string)
	err = json.Unmarshal(content, &value)
	if err != nil {
		t.Fatalf("json.Unmarshal() unexpected error: %s", err)
	}
	expected := map[string]map[string][]string{
		"MIT": {
			"github.com/spyzhov/ajson": {"v0.2.1"},
			"gopkg.in/yaml.v3":         {"v3.0.0-20200313102051-9f266ea9e77c"},
		},
		"BSD-2-Clause": {
			"gopkg.in/check.v1": {"v0.0.0-20161208181325-20d25e280405", "v1.0.0-20180628173108-788fd7840127"},
		},
		"Apache-2.0": {
			"gopkg.in/yaml.v2": {"v2.2.2"},
			"gopkg.in/yaml.v3": {"v3.0.0-20200313102051-9f266ea9e77c"},
		},
		"Unknown": {
			"gitlab.com/spyzhov/private": {"v0.2.1"},
		},
	}
	if !reflect.DeepEqual(expected, value) {
		t.Errorf("reflect.DeepEqual() not equal: \nexpected: %v\n  actual: %v", expected, value)
	}
}

func TestMain_Markdown(t *testing.T) {
	request := strings.Join([]string{
		examples["MIT"],
	}, "\n")
	content, err := runTest(t, request, "markdown")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected := []byte(`# Inherited Licenses

## [MIT](https://opensource.org/licenses/MIT)

| Package | Versions |
| --- | --- |
| github.com/spyzhov/ajson | [v0.2.1](https://pkg.go.dev/github.com/spyzhov/ajson@v0.2.1?tab=licenses) |
`)
	if !reflect.DeepEqual(expected, content) {
		t.Errorf("reflect.DeepEqual() not equal: \nexpected: %s\n  actual: %s", expected, content)
	}
}

func TestMain_CSV(t *testing.T) {
	request := strings.Join([]string{
		examples["MIT"],
	}, "\n")
	content, err := runTest(t, request, "csv")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected := []byte(`License,Package,Versions
MIT,github.com/spyzhov/ajson,v0.2.1
`)
	if !reflect.DeepEqual(expected, content) {
		t.Errorf("reflect.DeepEqual() not equal: \nexpected: %s\n  actual: %s", expected, content)
	}
}

func TestMain_HTML(t *testing.T) {
	request := strings.Join([]string{
		examples["MIT"],
		examples["BSD-2-Clause"],
	}, "\n")
	content, err := runTest(t, request, "html")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected := []byte(`<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>Inherited Licenses</title>
</head>
<body>
<h1>Inherited Licenses</h1>
<hr />
<h2><a href="https://opensource.org/licenses/BSD-2-Clause" target="_blank">BSD-2-Clause</a></h2>
<table>
    <thead>
	<tr>
		<th>Package</th>
		<th>Versions</th>
	</tr>
    </thead>
    <tbody>
    <tr>
        <td valign="top" rowspan="2">gopkg.in/check.v1</td>
        <td><a href="https://pkg.go.dev/gopkg.in/check.v1@v0.0.0-20161208181325-20d25e280405?tab=licenses" target="_blank">v0.0.0-20161208181325-20d25e280405</a></td>
    </tr>
    <tr>
		<td><a href="https://pkg.go.dev/gopkg.in/check.v1@v1.0.0-20180628173108-788fd7840127?tab=licenses" target="_blank">v1.0.0-20180628173108-788fd7840127</a></td>
	</tr>
    </tbody>
</table>
<hr />
<h2><a href="https://opensource.org/licenses/MIT" target="_blank">MIT</a></h2>
<table>
    <thead>
	<tr>
		<th>Package</th>
		<th>Versions</th>
	</tr>
    </thead>
    <tbody>
    <tr>
		<td>github.com/spyzhov/ajson</td>
		<td><a href="https://pkg.go.dev/github.com/spyzhov/ajson@v0.2.1?tab=licenses" target="_blank">v0.2.1</a></td>
	</tr>
    </tbody>
</table>
`)
	if !reflect.DeepEqual(expected, content) {
		t.Errorf("reflect.DeepEqual() not equal: \nexpected: %s\n  actual: %s", expected, content)
	}
}

func runTest(t *testing.T, content string, format string) ([]byte, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "go-license")
	if err != nil {
		t.Fatalf("can't prepare dir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(dir) // clean up
	}()

	goSum := filepath.Join(dir, "go.sum")
	if err := ioutil.WriteFile(goSum, []byte(content), 0600); err != nil {
		return nil, err
	}
	output := filepath.Join(dir, "output")

	os.Args = []string{
		"go-license",
		"-format", format,
		"-input", goSum,
		"-output", output,
		"-v",
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	main()
	return ioutil.ReadFile(output)
}
