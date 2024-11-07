package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"unicode"
)

type InvalidInput struct {
}

func (e *InvalidInput) Error() string {
	return "invalid input"
}

func ParseCount(runes []rune, startAt int) (count int, usedBefore int) {
	count = 0

	for ; startAt < len(runes); startAt++ {
		if !unicode.IsDigit(runes[startAt]) {
			break
		}

		count = count*10 + int(runes[startAt]-'0')
	}

	if count == 0 {
		return 1, startAt
	}

	return count, startAt
}

func HandleEscapeSeq(runes []rune, startAt int) (parsedRune rune, usedBefore int, err error) {
	if startAt >= len(runes) {
		return -1, -1, new(InvalidInput)
	}

	if unicode.IsDigit(runes[startAt]) {
		return runes[startAt], startAt + 1, nil
	}

	if runes[startAt] == '\\' {
		return runes[startAt], startAt + 1, nil
	}

	return -1, -1, new(InvalidInput)

}

func ParseRune(runes []rune, startAt int) (parsedRune rune, usedBefore int, err error) {
	if startAt >= len(runes) {
		return -1, -1, new(InvalidInput)
	}

	if runes[startAt] == '\\' {
		return HandleEscapeSeq(runes, startAt+1)
	}

	if unicode.IsDigit(runes[startAt]) {
		return -1, -1, new(InvalidInput)
	}

	return runes[startAt], startAt + 1, nil
}

func Unpack(str string) (*string, error) {
	var buffer bytes.Buffer
	runes := []rune(str)

	for curIdx := 0; curIdx < len(runes); {
		var err error
		var rn rune
		var count int

		rn, curIdx, err = ParseRune(runes, curIdx)
		if err != nil {
			return nil, err
		}

		count, curIdx = ParseCount(runes, curIdx)

		for range count {
			buffer.WriteRune(rn)
		}
	}

	st := buffer.String()

	return &st, nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Println("couldn't read input")
	}

	st := scanner.Text()

	res, err := Unpack(st)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println(string(*res))
}
