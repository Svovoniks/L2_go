package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type sortItem struct {
	val          string
	sortKeyStr   string
	sortKeyFloat float64
	isNumber     bool
}

type config struct {
	inputFile        string
	outputFile       string
	column           int
	sep              string
	numSort          bool
	reverse          bool
	uniqueOutput     bool
	monthSort        bool
	ignoreTrailSpace bool
	checkSorted      bool
	withSuffix       bool
}

func sortStrData(arr []*sortItem) {
	for idx := range arr {
		arr[idx].sortKeyStr = strings.ToLower(arr[idx].sortKeyStr)
	}

	reverseASCII := func(a, b *sortItem) int {
		for i := range min(len(a.sortKeyStr), len(b.sortKeyStr)) {
			if a.sortKeyStr[i] != b.sortKeyStr[i] {
				return 0
			}

			if a.val[i] == b.val[i] {
				continue
			}

			if unicode.IsLower(rune(a.val[i])) {
				return -1
			}
			return 1

		}

		if len(a.val) < len(b.val) {
			return -1
		}

		return 1
	}

	slices.SortFunc(arr, func(a, b *sortItem) int {
		return strings.Compare(a.sortKeyStr, b.sortKeyStr)
	})

	slices.SortStableFunc(arr, reverseASCII)
}

func sortNumeric(arr []*sortItem) {
	var nums []*sortItem
	var NaNs []*sortItem

	for idx := range arr {
		if arr[idx].isNumber {
			nums = append(nums, arr[idx])
		} else {
			NaNs = append(NaNs, arr[idx])
		}
	}

	floatSort := func(a, b *sortItem) int {
		if a.sortKeyFloat < b.sortKeyFloat {
			return -1
		} else if a.sortKeyFloat > b.sortKeyFloat {
			return 1
		}
		return 0
	}

	slices.SortFunc(nums, floatSort)
	sortStrData(NaNs)

	gIdx := 0
	for idx := range nums {
		arr[gIdx] = nums[idx]
		gIdx++
	}

	for idx := range NaNs {
		arr[gIdx] = NaNs[idx]
		gIdx++
	}
}

func getSortData(filename string) ([]string, error) {
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

func sendSorted(data []*sortItem, cfg *config) {
	cur := 0
	end := len(data)
	incr := 1

	if cfg.reverse {
		end = 1
		cur = len(data) - 1
		incr = -1
	}

	var stream io.Writer = os.Stdout

	if cfg.outputFile != "" {
		file, err := os.Create(cfg.outputFile)

		if err != nil {
			fmt.Println("Couldn't open output file")
			os.Exit(1)
		}

		stream = file
		defer file.Close()
	}

	var prev string
	var startFlag = true

	for ; incr*cur < end; cur += incr {
		if cfg.uniqueOutput && !startFlag && prev == data[cur].val {
			continue
		}

		fmt.Fprintln(stream, data[cur].val)
		startFlag = false
		prev = data[cur].val
	}
}

func intoSortItems(data []string, cfg *config) []*sortItem {
	var getKey func(string) string

	if cfg.column < 0 {
		getKey = func(s string) string {
			return s
		}
	} else {
		getKey = func(s string) string {
			if len(s) == 0 {
				return s
			}

			spl := strings.Split(s, cfg.sep)
			if len(spl) > cfg.column {
				return spl[cfg.column]
			}

			return ""
		}
	}

	res := make([]*sortItem, len(data))
	for i := range data {
		res[i] = &sortItem{
			val:        data[i],
			sortKeyStr: getKey(data[i]),
		}
	}

	if cfg.numSort || cfg.withSuffix {
		tryToNumber(res, cfg.withSuffix)
	}

	if cfg.monthSort {
		tryToMonth(res)
	}

	return res
}

func tryToNumber(data []*sortItem, withSuffix bool) {
	var getNum func(string) (float64, bool)

	var defaultGetNum = func(s string) (float64, bool) {
		num, err := strconv.ParseFloat(s, 64)
		return num, err == nil
	}

	suffixMap := map[rune]float64{
		'k': 1 << 10,
		'm': 1 << 20,
		'g': 1 << 30,
		't': 1 << 40,
		'p': 1 << 50,
	}

	if withSuffix {
		getNum = func(s string) (float64, bool) {
			if num, isNum := defaultGetNum(s); isNum {
				return num, isNum
			}

			if len(s) < 2 {
				return 0, false
			}

			var num float64 = 0
			var isNum = false

			if parsed, err := strconv.ParseFloat(s[:len(s)-1], 64); err == nil {
				num = parsed
				isNum = true
			}

			if isNum {
				mult, ok := suffixMap[unicode.ToLower(rune(s[len(s)-1]))]
				num *= mult
				isNum = ok
			}

			return num, isNum
		}
	} else {
		getNum = defaultGetNum
	}

	for i := range data {
		data[i].sortKeyFloat, data[i].isNumber = getNum(data[i].sortKeyStr)
	}

}

func tryToMonth(data []*sortItem) {
	months := []string{
		"january",
		"february",
		"march",
		"april",
		"may",
		"june",
		"july",
		"august",
		"september",
		"october",
		"november",
		"december",
	}

	for dataIdx := range data {
		if len(data[dataIdx].sortKeyStr) < 3 {
			continue
		}

		for monthIdx := range months {
			if strings.HasPrefix(months[monthIdx], strings.ToLower(data[dataIdx].sortKeyStr)) {
				data[dataIdx].sortKeyFloat = float64(monthIdx)
				data[dataIdx].isNumber = true
				break
			}
		}
	}
}

func sortData(data []*sortItem, cfg *config) {
	if cfg.withSuffix || cfg.numSort || cfg.monthSort {
		sortNumeric(data)
		return
	}

	sortStrData(data)
}

func boolToInt(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

func getConfig() *config {
	cfg := new(config)

	flag.IntVar(&cfg.column, "k", -1, "Use values of given column for sorting")
	flag.StringVar(&cfg.sep, "s", " ", "Use as separator for columns (Default is single space)")
	flag.BoolVar(&cfg.numSort, "n", false, "Sort by numeric value")
	flag.BoolVar(&cfg.reverse, "r", false, "Sort in reverse")
	flag.BoolVar(&cfg.uniqueOutput, "u", false, "Only display unique lines")
	flag.BoolVar(&cfg.monthSort, "M", false, "Sort by month")
	flag.BoolVar(&cfg.ignoreTrailSpace, "b", false, "Ignore trail space")
	flag.BoolVar(&cfg.withSuffix, "h", false, "Sort by numeric value considering suffix")
	flag.BoolVar(&cfg.checkSorted, "c", false, "Check if sorted")
	flag.Parse()

	cfg.inputFile = flag.Arg(0)
	cfg.outputFile = flag.Arg(1)

	if boolToInt(cfg.numSort)+boolToInt(cfg.withSuffix)+boolToInt(cfg.monthSort) > 1 {
		fmt.Println("You can only choose one sorting methond")
		os.Exit(1)
	}

	return cfg
}

func checkSorted(data []*sortItem, cfg *config) int {
	cmpr := func(i1 *sortItem, i2 *sortItem) bool { return i1.sortKeyStr < i2.sortKeyStr }

	if cfg.monthSort || cfg.numSort || cfg.withSuffix {
		cmpr = func(i1 *sortItem, i2 *sortItem) bool { return i1.sortKeyFloat < i2.sortKeyFloat }
	}

	for idx := 0; idx < len(data)-1; idx++ {
		if cmpr(data[idx], data[idx+1]) == cfg.reverse {
			return idx + 2
		}
	}

	return -1
}

func sendCheck(upTo int) {
	if upTo == -1 {
		fmt.Println("File is sorted")
		return
	}
	fmt.Printf("File is sorted up to line: %v\n", upTo)
}

func main() {
	cfg := getConfig()

	data, err := getSortData(cfg.inputFile)
	if err != nil {
		fmt.Println("Couldn't read input file")
		os.Exit(1)
	}

	if cfg.ignoreTrailSpace {
		for i := range data {
			data[i] = strings.Trim(data[i], " ")
		}
	}

	sortItems := intoSortItems(data, cfg)

	if cfg.checkSorted {
		sendCheck(checkSorted(sortItems, cfg))
		return
	}

	sortData(sortItems, cfg)

	sendSorted(sortItems, cfg)
}
