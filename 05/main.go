package main

import (
	"fmt"
	"slices"
	"strings"
)

type rusAnagramImprint [33]int

type anagramEntry struct {
	first string
	arr   []string
}

func getImprint(st string) (imprint rusAnagramImprint) {
	for _, rn := range st {
		if rn == 0x451 {
			imprint[32] += 1
		} else {
			imprint[rn-0x430] += 1
		}
	}
	return imprint
}

func GetAnagrams(ls []string) map[string][]string {
	mp := make(map[rusAnagramImprint]*anagramEntry)

	for i := range ls {
		str := strings.ToLower(ls[i])
		imprint := getImprint(str)
		entry := mp[imprint]

		if entry == nil {
			mp[imprint] = &anagramEntry{first: str, arr: []string{str}}
            continue
		}

		idx, found := slices.BinarySearch(entry.arr, str)

		if found {
			continue
		}

		entry.arr = slices.Insert(entry.arr, idx, str)
	}

	correctMp := make(map[string][]string)

	for _, v := range mp {
		if len(v.arr) < 2 {
			continue
		}

		correctMp[v.first] = v.arr
	}

	return correctMp
}

func main() {
	ls := []string{
		"пятка",
        "апятк",
		"пятак",
		"тяпка",
		"тяпка",
		"листок",
		"слиток",
		"столик",
		"столик",
		"столик",
		"слик",
	}
	for k, v := range GetAnagrams(ls) {
		fmt.Println(k, v)
	}
}
