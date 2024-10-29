package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getLinks(curDir string, page string, linkMap map[string]bool) (string, []string) {
	rx := regexp.MustCompile(`href=["'].*?["']`)

	var newLinks []string
	replFunc := func(linkStr string) string {
		linkStr = linkStr[6 : len(linkStr)-1]
		if linkStr[0] != '/' {
			return linkStr
		}

		if strings.Contains(linkStr, "#") {
			linkStr = strings.Split(linkStr, "#")[0]
		}

		if _, ok := linkMap[linkStr]; ok {
			return curDir + linkStr
		}
		fmt.Println(linkStr)

		linkMap[linkStr] = true
		newLinks = append(newLinks, linkStr)

		return "href=\"" + curDir + linkStr + "\""
	}

	page = rx.ReplaceAllStringFunc(page, replFunc)

	return page, newLinks
}

func DownloadPage(link string) (string, error) {
	rsp, err := http.Get(link)
	if err != nil {
		return "", err
	}

	fmt.Printf("Downloading: %v\n", link)

	byt, err1 := io.ReadAll(rsp.Body)
	if err1 != nil {
		return "", err1
	}

	return string(byt), nil
}

func storePage(curDir string, page string, path string) {
	fmt.Println("storing path", path)

	dirPath := filepath.Join(curDir, filepath.Dir(path))
	gotIndex := false

	if stats, err := os.Stat(dirPath); err == nil && !stats.IsDir() {
		os.Rename(dirPath, dirPath+"tmp.file")
		gotIndex = true
	}

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		panic(err.Error())
	}

	if gotIndex {
		os.Rename(dirPath+"tmp.file", filepath.Join(dirPath, "index.html"))
	}

	filePath := filepath.Join(curDir, path)

	if stats, err := os.Stat(filePath); err == nil && stats.IsDir() {
		filePath = filepath.Join(filePath, "index.html")
	}

	file, err := os.Create(filePath)

	if err != nil {
		panic("Couldn't store a page" + err.Error())
	}
	defer file.Close()

	file.WriteString(page)
}

func main() {
	domain := "https://pkg.go.dev"

	curDir, err := os.Getwd()
	if err != nil {
		return
	}
	curDir = filepath.Join(curDir, "web")

	fmt.Printf("Downloading into %v", curDir)

	indexPage, err1 := DownloadPage(domain)
	if err1 != nil {
		return
	}

	var linkMap = map[string]bool{
		"/": true,
	}

	indexPage, links := getLinks(curDir, indexPage, linkMap)

	storePage(curDir, indexPage, "index.html")

	for len(links) > 0 {
		var newLinks []string

		for i := range links {
			page, err := DownloadPage(domain + links[i])

			if err != nil {
				fmt.Printf("Couldn't download: %v\n", domain+links[i])
				continue
			}

			convertedPage, pageLinks := getLinks(curDir, page, linkMap)
			newLinks = append(newLinks, pageLinks...)

			storePage(curDir, convertedPage, links[i])
		}

		links = newLinks
	}
}
