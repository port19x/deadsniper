package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

const helptext = "Usage: deadsniper <link to sitemap.xml>"

var exitCode int = 0
var mutex sync.Mutex

func reqWrap(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("reqWrap: failed to make request: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("reqWrap: failed to read response body: %v", err)
	}
	body2 := string(body)
	return body2
}

func linkSlurp(url string) []string {
	body := reqWrap(url)
	re := regexp.MustCompile(`"(https://.+?)"`)
	matches := re.FindAllStringSubmatch(body, -1)

	var links []string
	for _, matchSlice := range matches {
		if len(matchSlice) > 1 {
			links = append(links, matchSlice[1]) // append the first submatch
		}
	}
	return links
}

func isLinkAlive(url string, toplevel string) string {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		if exitCode != 1 {
			mutex.Lock() //Might be unneeded
			exitCode = 1
			mutex.Unlock()
		}
		return fmt.Sprintf("❌: %s -> %s\n", toplevel, url)
	} else {
		return fmt.Sprintf("✓: %s -> %s\n", toplevel, url)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(helptext)
		return
	}
	url := os.Args[1]
	body2 := reqWrap(url)

	re := regexp.MustCompile(`<loc>(https://.+?)</loc>`)
	to_test := re.FindAllStringSubmatch(body2, -1)
	var body3 []string
	for _, matchSlice := range to_test {
		if len(matchSlice) > 1 {
			body3 = append(body3, matchSlice[1])
		}
	}

	results := make(chan string)
	var wg sync.WaitGroup

	// Further parallelizing linkSlurp does not help due to CPU and Network Limitations, I tried
	for _, toplevel := range body3 {
		labrat := linkSlurp(toplevel)
		for _, link := range labrat {
			wg.Add(1)
			go func(link string) {
				defer wg.Done()
				results <- isLinkAlive(link, toplevel)
			}(link)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Prints results as they come in
	for result := range results {
		fmt.Print(result)
	}

	os.Exit(exitCode)
}
