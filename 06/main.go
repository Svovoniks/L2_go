package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
)

type match struct {
	line int
	from []int
	to   []int
}

type config struct {
	after      int
	before     int
	context    int
	count      bool
	ignoreCase bool
	invert     bool
	fixed      bool
	lineNum    bool
	inputFile  string
	pattern    string
}

func getConfig() *config {
	cfg := new(config)

	flag.IntVar(&cfg.after, "A", 0, "lines after")
	flag.IntVar(&cfg.before, "B", 0, "lines before")
	flag.IntVar(&cfg.context, "C", 0, "before + after")
	flag.BoolVar(&cfg.count, "c", false, "line count")
	flag.BoolVar(&cfg.ignoreCase, "i", false, "ignore case")
	flag.BoolVar(&cfg.invert, "v", false, "invert")
	flag.BoolVar(&cfg.fixed, "F", false, "exact match")
	flag.BoolVar(&cfg.lineNum, "h", false, "line num")
	flag.Parse()

	cfg.inputFile = flag.Arg(0)
	cfg.pattern = flag.Arg(1)

	if cfg.context != 0 {
		if cfg.before != 0 || cfg.after != 0 {
			fmt.Println("You can't specify both before/after and context")
			os.Exit(1)
		}

		cfg.before = cfg.context
		cfg.after = cfg.context
	}

	if cfg.pattern == "" {
		fmt.Println("No pattern specified")
		os.Exit(1)
	}

	return cfg
}

func getFileContent(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	lines := make([]string, 0)

	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	if err = sc.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func findAll(data []string, re *regexp.Regexp) (matches []*match) {
	for i := range data {
		lineMatches := re.FindAllStringIndex(data[i], -1)

		if lineMatches == nil {
			continue
		}

		match := new(match)
		match.line = i

		for matchIdx := range lineMatches {
			match.from = append(match.from, lineMatches[matchIdx][0])
			match.to = append(match.to, lineMatches[matchIdx][1])
		}

		matches = append(matches, match)
	}

	return matches
}

func findString(data []string, cfg *config) []*match {
	rgx := regexp.QuoteMeta(cfg.pattern)
	if cfg.fixed {
		rgx = "(?i)" + rgx
	}

	re := regexp.MustCompile(rgx)

	return findAll(data, re)
}

func findPattern(data []string, re *regexp.Regexp) []*match {
	return findAll(data, re)
}

func printInRange(data []string, start int, end int, lastLine int) int {
	if start < 0 {
		start = 0
	}

	if start <= lastLine {
		start = lastLine + 1
	}

	if end > len(data) {
		end = len(data)
	}

	if end <= lastLine {
		return lastLine
	}

	if start != lastLine+1 && lastLine >= 0 {
		fmt.Println("\u001B[34m--\u001B[0m")
	}

	for ; start < end; start++ {
		fmt.Println(data[start])
	}

	return end - 1

}

func sendResults(data []string, res []*match, cfg *config) {
	if cfg.count {
		fmt.Printf("Found match in %v line(s)\n", len(res))
		return
	}

	skippedLine := -1
	resIdx := 0

	for i := range data {
		skip := !cfg.invert

		if len(res) > resIdx {
			isMatch := i == res[resIdx].line
			skip = isMatch == cfg.invert
			if isMatch {
				resIdx += 1
			}
		}

		if skip {
			if skippedLine != -1 {
				skippedLine = 1
			}
			continue
		}

		if skippedLine == 1 {
			fmt.Println("\u001B[34m--\u001B[0m")
		}

		skippedLine = 0

		fmt.Println(data[i])
	}

}

func buildHighlitedString(str string, starts []int, ends []int) string {
	var buffer bytes.Buffer

	if len(starts) < 1 {
		return str
	}

	cur := 0

	for i := range starts {
		buffer.WriteString(str[cur:starts[i]])
		buffer.WriteString("\u001B[31m")
		buffer.WriteString(str[starts[i]:ends[i]])
		buffer.WriteString("\u001B[0m")
		cur = ends[i]
	}

	buffer.WriteString(str[cur:])

	return buffer.String()

}

func highlightResults(data []string, matches []*match) {
	for i := range matches {
		data[matches[i].line] = buildHighlitedString(data[matches[i].line], matches[i].from, matches[i].to)
	}
}

func getRegEx(cfg *config) *regexp.Regexp {
	rgx := cfg.pattern

	if cfg.fixed {
		rgx = regexp.QuoteMeta(cfg.pattern)
	}

	if cfg.ignoreCase {
		rgx = "(?i)" + rgx
	}

	re, err := regexp.Compile(rgx)
	if err != nil {
		fmt.Println("Invalid regex, if you want to look for this exact string use -F flag")
		os.Exit(1)
	}

	return re
}

func main() {
	cfg := getConfig()

	data, err := getFileContent(cfg.inputFile)
	if err != nil {
		fmt.Println("Couldn't read input file")
		os.Exit(1)
	}

	var matches []*match

	matches = findPattern(data, getRegEx(cfg))

	highlightResults(data, matches)

	sendResults(data, matches, cfg)
}
