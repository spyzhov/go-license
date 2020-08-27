package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type packageMap map[string]map[string][]string

type request struct {
	name    string
	version string
}

var re = regexp.MustCompile(`<div id="#lic-0">([^<]*)</div>`)
var isVerbose bool
var licenseUrl = map[string]string{
	"AGPL-3.0":     "https://opensource.org/licenses/AGPL-3.0",
	"Apache-2.0":   "https://opensource.org/licenses/Apache-2.0",
	"Artistic-2.0": "https://opensource.org/licenses/Artistic-2.0",
	"BSD-2-Clause": "https://opensource.org/licenses/BSD-2-Clause",
	"BSD-3-Clause": "https://opensource.org/licenses/BSD-3-Clause",
	"BSL-1.0":      "https://opensource.org/licenses/BSL-1.0",
	"EPL-1.0":      "https://opensource.org/licenses/EPL-1.0",
	"EPL-2.0":      "https://opensource.org/licenses/EPL-2.0",
	"GPL-2.0":      "https://opensource.org/licenses/GPL-2.0",
	"GPL-3.0":      "https://opensource.org/licenses/GPL-3.0",
	"ISC":          "https://opensource.org/licenses/ISC",
	"LGPL-2.1":     "https://opensource.org/licenses/LGPL-2.1",
	"LGPL-3.0":     "https://opensource.org/licenses/LGPL-3.0",
	"MIT":          "https://opensource.org/licenses/MIT",
	"MPL-2.0":      "https://opensource.org/licenses/MPL-2.0",
	"NCSA":         "https://opensource.org/licenses/NCSA",
	"OSL-3.0":      "https://opensource.org/licenses/OSL-3.0",
	"Zlib":         "https://opensource.org/licenses/Zlib",
}

func main() {
	format := flag.String("format", "json", "output format [csv, md, markdown, html, json]")
	inFile := flag.String("input", "go.sum", "input path for go.sum file")
	outFile := flag.String("output", "", "output file name (default STDOUT)")
	flag.BoolVar(&isVerbose, "v", false, "verbose output")
	flag.Parse()

	validate(format, inFile)
	content, err := ioutil.ReadFile(*inFile)
	if err != nil {
		fail("fail to read %s: %s", *inFile, err.Error())
	}
	verbose("found file: %s, size: %d", *inFile, len(content))
	packages := parse(content)
	verbose("found %d unique inherited package", len(packages))
	output := os.Stdout
	if outFile != nil && *outFile != "" {
		output, err = os.Create(*outFile)
		if err != nil {
			fail("failed to create file %s: %s", *outFile, err.Error())
		}
		verbose("output to: %s", *outFile)
		defer func() {
			_ = output.Close()
		}()
	} else {
		verbose("output to: STDOUT")
	}

	wg := new(sync.WaitGroup)
	size := runtime.GOMAXPROCS(0)
	jobs := make(chan *request, size)
	for i := 0; i < size; i++ {
		wg.Add(1)
		go worker(i, wg, jobs, packages)
	}
	for name, versions := range packages {
		for version := range versions {
			jobs <- &request{
				name:    name,
				version: version,
			}
		}
	}
	close(jobs)
	wg.Wait()

	_, err = output.WriteString(packageMap(packages).Print(*format))
	if err != nil {
		_ = output.Close()
		fail("failed to write data: %s", err.Error())
	}
}

func verbose(message string, args ...interface{}) {
	if isVerbose {
		log.Printf(message, args...)
	}
}

func fail(message string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, message+"\n", args...)
	os.Exit(1)
}

func validate(format *string, inFile *string) {
	if format == nil || *format == "" {
		fail("please select valid -format")
		return
	}
	switch strings.ToLower(*format) {
	case "json", "md", "markdown", "csv", "html":
		verbose("valid format: %s", *format)
	default:
		fail("please select valid -format: csv, md, markdown, html, json")
	}
	if inFile == nil {
		fail("please select valid -input file")
		return
	}
	if path.Base(*inFile) != "go.sum" {
		*inFile = path.Join(*inFile, "go.sum")
	}
	info, err := os.Stat(*inFile)
	if os.IsNotExist(err) {
		fail("please select valid -input file: file not exists")
	}
	if info.IsDir() {
		fail("please select valid -input file: selected directory")
	}
}

func parse(content []byte) (result map[string]map[string][]string) {
	result = make(map[string]map[string][]string)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) != 3 {
			log.Printf("ERROR: wrong line: %s", line)
			continue
		}
		if _, ok := result[parts[0]]; !ok {
			result[parts[0]] = make(map[string][]string)
		}
		result[parts[0]][strings.TrimSuffix(parts[1], "/go.mod")] = make([]string, 0)
	}
	return result
}

func worker(index int, wg *sync.WaitGroup, jobs <-chan *request, result map[string]map[string][]string) {
	defer verbose("worker done: %d", index)
	defer wg.Done()
	verbose("worker start: %d", index)
	for req := range jobs {
		result[req.name][req.version] = strings.Split(get(req), ", ")
		verbose("for %s found licenses: %s", req, strings.Join(result[req.name][req.version], ", "))
	}
}

func get(req *request) string {
	url := fmt.Sprintf("https://pkg.go.dev/%s?tab=licenses", req)
	verbose("GET: %s", url)
	start := time.Now()
	// #nosec G107
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("ERROR: failed to get licenses value for: %s -> %s", req, err.Error())
		return "Unknown"
	}
	verbose("GOT: %s as Status: %s, in a %s", url, resp.Status, time.Since(start))
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: failed to get licenses value for: %s -> %s", req, resp.Status)
		return "Unknown"
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: failed to read response body: %s -> %s", req, err.Error())
		return "Unknown"
	}
	result := re.Find(content)
	str := ""
	if len(result) > 0 {
		str = string(result[17 : len(result)-6])
	}
	if str == "" {
		str = "Unknown"
	}
	return str
}

func (r *request) String() string {
	return r.name + "@" + r.version
}

func (r packageMap) Print(format string) string {
	verbose("generate content: %s", strings.ToLower(format))
	switch strings.ToLower(format) {
	case "json":
		return r.json()
	case "md", "markdown":
		return r.markdown()
	case "csv":
		return r.csv()
	case "html":
		return r.html()
	}
	return ""
}

func (r packageMap) json() string {
	data, err := json.Marshal(r.reverse())
	if err != nil {
		fail("failed to generate JSON response: %s", err.Error())
	}
	return string(data)
}

func (r packageMap) markdown() string {
	result := "# Inherited Licenses\n"
	licenses := r.reverse()
	sorted := make([]string, 0, len(licenses))
	for name := range licenses {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	for _, license := range sorted {
		if url, ok := licenseUrl[license]; ok {
			result += fmt.Sprintf("\n## [%s](%s)\n\n", license, url)
		} else {
			result += fmt.Sprintf("\n## %s\n\n", license)
		}
		result += "| Package | Versions |\n| --- | --- |\n"
		packages := keys(licenses[license])
		for _, name := range packages {
			sort.Strings(licenses[license][name])
			result += fmt.Sprintf("| %s | ", name)
			for i, version := range licenses[license][name] {
				if i != 0 {
					result += " <br /> "
				}
				result += fmt.Sprintf("[%s](https://pkg.go.dev/%s@%s?tab=licenses)", version, name, version)
			}
			result += " |\n"
		}
	}

	return result
}

func (r packageMap) html() string {
	result := `<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>Inherited Licenses</title>
</head>
<body>
<h1>Inherited Licenses</h1>
<hr />
`
	licenses := r.reverse()
	sorted := make([]string, 0, len(licenses))
	for name := range licenses {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	for j, license := range sorted {
		if j != 0 {
			result += "<hr />\n"
		}
		if url, ok := licenseUrl[license]; ok {
			result += fmt.Sprintf("<h2><a href=\"%s\" target=\"_blank\">%s</a></h2>\n", url, license)
		} else {
			result += fmt.Sprintf("<h2>%s</h2>\n", license)
		}
		result += `<table>
    <thead>
	<tr>
		<th>Package</th>
		<th>Versions</th>
	</tr>
    </thead>
    <tbody>
`
		packages := keys(licenses[license])
		for _, name := range packages {
			sort.Strings(licenses[license][name])
			for i, version := range licenses[license][name] {
				if len(licenses[license][name]) == 1 {
					result += fmt.Sprintf(`    <tr>
		<td>%s</td>
		<td><a href="https://pkg.go.dev/%s@%s?tab=licenses" target="_blank">%s</a></td>
	</tr>
`, name, name, version, version)
				} else {
					if i == 0 {
						result += fmt.Sprintf(`    <tr>
        <td valign="top" rowspan="%d">%s</td>
        <td><a href="https://pkg.go.dev/%s@%s?tab=licenses" target="_blank">%s</a></td>
    </tr>
`, len(licenses[license][name]), name, name, version, version)
					} else {
						result += fmt.Sprintf(`    <tr>
		<td><a href="https://pkg.go.dev/%s@%s?tab=licenses" target="_blank">%s</a></td>
	</tr>
`, name, version, version)
					}
				}
			}
		}
		result += `    </tbody>
</table>
`
	}

	return result
}

func (r packageMap) csv() string {
	buf := bytes.Buffer{}
	writer := csv.NewWriter(&buf)
	err := writer.Write([]string{
		"License",
		"Package",
		"Versions",
	})
	if err != nil {
		fail(err.Error())
	}

	licenses := r.reverse()
	sorted := make([]string, 0, len(licenses))
	for name := range licenses {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	for _, license := range sorted {
		packages := keys(licenses[license])
		for _, name := range packages {
			sort.Strings(licenses[license][name])
			err = writer.Write([]string{
				license,
				name,
				strings.Join(licenses[license][name], ", "),
			})
			if err != nil {
				log.Printf("ERROR: faled to write csv line: %s", err.Error())
			}
		}
	}
	writer.Flush()
	return buf.String()
}

func (r packageMap) reverse() map[string]map[string][]string {
	result := make(map[string]map[string][]string)
	for name, versions := range r {
		for version, licenses := range versions {
			for _, license := range licenses {
				if _, ok := result[license]; !ok {
					result[license] = make(map[string][]string)
				}
				if _, ok := result[license][name]; !ok {
					result[license][name] = make([]string, 0)
				}
				result[license][name] = append(result[license][name], version)
			}
		}
	}
	return result
}

func keys(array map[string][]string) []string {
	list := make([]string, 0, len(array))
	for name := range array {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}
