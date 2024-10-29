package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type field interface {
	getString(strs []string) []string
}

type rangeFiled struct {
	start int
	end   int
}

func (r *rangeFiled) getString(strs []string) (res []string) {
	start := r.start
	if start == -1 {
		start = 0
	}

	end := r.end
	if end == -1 {
		end = len(strs)
	}

	for ; start < end && start < len(strs); start++ {
		res = append(res, strs[start])
	}

	return res
}

func parseRangeField(str string) (*rangeFiled, error) {
	parts := strings.Split(str, "-")
	err := errors.New("Invalid range field")

	if len(parts) != 2 {
		return nil, err
	}

	num1 := -1
	if parts[0] != "" {
		num, err1 := strconv.Atoi(parts[0])
		if err1 != nil {
			return nil, err
		}
		num1 = num
	}

	num2 := -1
	if parts[1] != "" {
		num, err2 := strconv.Atoi(parts[1])
		if err2 != nil {
			return nil, err
		}
		num2 = num
	}

	if num1 <= 0 && num2 <= 0 {
		return nil, err
	}

	return &rangeFiled{start: num1 - 1, end: num2}, nil
}

type singleFiled struct {
	idx int
}

func (r *singleFiled) getString(strs []string) (res []string) {
	if r.idx < len(strs) {
		res = append(res, strs[r.idx])
	}
	return res
}

func parseSingleField(str string) (*singleFiled, error) {
	if num, err := strconv.ParseInt(str, 10, 64); err == nil && num > 0 {
		return &singleFiled{idx: int(num) - 1}, nil
	}

	return nil, errors.New("Invalid single field")
}

type config struct {
	fields    []field
	delimeter string
	separated bool
}

func parseFields(fieldsStr string) ([]field, error) {
	ls := strings.Split(fieldsStr, ",")
	fields := []field{}

	for i := range ls {
		if field, err := parseSingleField(ls[i]); err == nil {
			fields = append(fields, field)
			continue
		}

		if field, err := parseRangeField(ls[i]); err == nil {
			fields = append(fields, field)
			continue
		}

		return nil, errors.New("Couldn't parse fileds")
	}

	return fields, nil
}

func getConfig() *config {
	cfg := new(config)

	var fieldsStr string

	flag.StringVar(&fieldsStr, "f", "", "fields")
	flag.StringVar(&cfg.delimeter, "d", "	", "delimeter")
	flag.BoolVar(&cfg.separated, "s", false, "separated")
	flag.Parse()

	fields, err := parseFields(fieldsStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg.fields = fields

	return cfg
}

func getLines() (lines []string) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines

}

func applyCut(line string, cfg *config) (string, bool) {
	parts := strings.Split(line, cfg.delimeter)
	if len(parts) < 2 {
		return line, false
	}

	res := []string{}

	for flIdx := range cfg.fields {
		strs := cfg.fields[flIdx].getString(parts)
		res = append(res, strs...)
	}

	return strings.Join(res, cfg.delimeter), true
}

func sendResults(lines []string, cfg *config) {
	for i := range lines {
		str, succ := applyCut(lines[i], cfg)
		if succ || !cfg.separated {
			fmt.Println(str)
		}
	}
}

func main() {
	cfg := getConfig()
	lines := getLines()

	sendResults(lines, cfg)
}
