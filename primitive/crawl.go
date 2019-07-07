package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

var httpClient = &http.Client{
	// Timeout: time.Second * 10,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		// We use ABSURDLY large keys, and should probably not.
		TLSHandshakeTimeout: 60 * time.Second,
	},
}

func getRequest(url string) string {
	if !strings.HasPrefix(url, "http") {
		fmt.Print("had url: " + url)
		return ""
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		fmt.Print("I'm fail here.")
		log.Fatal(err)
		return ""
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print("IO error when reading response's body.")
		return ""
	}
	return string(content)
}

func getURLFromBody(body string) (uris []string) {
	hrefPattern := regexp.MustCompile(`<a\s+(.*?\s+)*?href=['"](.+?)['"](\s+.*?\s*)*?>.*?</a>`)
	hrefs := hrefPattern.FindAllStringSubmatch(body, -1)
	uris = make([]string, len(hrefs))
	for i, u := range hrefs {
		// fmt.Printf("found uri: %s\n", u[2])
		uris[i] = u[2]
	}
	return uris
}

func resolveAbsURL(baseURL string, uris []string) []string {
	base, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	var absURLs = make([]string, len(uris))
	for i, uri := range uris {
		u, err := url.Parse(uri)
		if err != nil {
			log.Fatal(err)
		}
		absURLs[i] = base.ResolveReference(u).String()
		// fmt.Printf("found url: %s\n", absURLs[i])
	}
	return absURLs
}

func crawlRoutine(url string, urlChan chan<- string) {
	defer func() {
		if r := recover(); r != nil {
			// 这里可以对异常进行一些处理和捕获
			fmt.Println("Recovered:", r)
		}
	}()

	htmlBody := getRequest(url)
	uris := getURLFromBody(htmlBody)
	absURLs := resolveAbsURL(url, uris)
	// fmt.Printf("go(%d): %s, put %d\n", GoID(), url, len(absURLs))
	for i, v := range absURLs {
		fmt.Printf("go(%d) put(%d): %s\n", GoID(), i, v)
		urlChan <- v
	}
	fmt.Printf("go(%d): end\n", GoID())
}

func main() {
	startURL := "https://www.nvshens.net/"
	seen := make(map[string]bool)
	urlChan := make(chan string, 10)
	urlChan <- startURL
	// n := 4
	for url := range urlChan {
		if !seen[url] {
			seen[url] = true
			go crawlRoutine(url, urlChan)
		}

	}
	// for i := 0; i < n; i++ {
	// 	go crawlRoutine(&urlChan)
	// }
	time.Sleep(10 * time.Second)
}
