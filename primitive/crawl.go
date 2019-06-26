package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

func getRequest(url string) string {
	var httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		fmt.Print("Network is down.")
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
		uris[i] = u[2]
	}
	return uris
}

func resolveAbsURL(baseURL string, uris []string) []*url.URL {
	base, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	var absURLs = make([]*url.URL, len(uris))
	for i, uri := range uris {
		u, err := url.Parse(uri)
		if err != nil {
			log.Fatal(err)
		}
		absURLs[i] = base.ResolveReference(u)
	}
	return absURLs
}

func main() {
	startURL := "https://www.nvshens.net/"
	htmlBody := getRequest(startURL)
	uris := getURLFromBody(htmlBody)
	absURLs := resolveAbsURL(startURL, uris)
	for i, v := range absURLs {
		fmt.Printf("%d: %s\n", i, v)
	}
}
