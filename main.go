package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

const helptext = "Usage: deadsniper <link to sitemap.xml>"

var exitCode int = 0
var mutex sync.Mutex

// This is a low level wrapper for http get requests to return the request body as a string or error otherwise.
func reqWrap(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("reqWrap: failed to make request: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("reqWrap: failed to read response body: %v", err)
	}
	return string(body)
}

// This checks a list of links. It also expects the corresponding toplevel, to make fixing easier
// This is an outside function to make it easily parallelizable
func isLinkAlive(url string, toplevel string) string {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
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
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println(helptext)
		return
	}
	if os.Args[1] == "-V" || os.Args[1] == "--version" {
		fmt.Println("v1.1 - 20240822")
		return
	}
	body := reqWrap(os.Args[1]) // Assumption: timeouts on the website being deadlink-checked do not occur

	// Parse all the requests from the sitemap file
	re := regexp.MustCompile(`<loc>(https://.+?)</loc>`)
	fat_result := re.FindAllStringSubmatch(body, -1)
	var sites_in_sitemap []string
	for _, matchSlice := range fat_result {
			sites_in_sitemap = append(sites_in_sitemap, matchSlice[1])
	}

	// Prepare waitgroup and channel for async processing
	results := make(chan string)
	var wg sync.WaitGroup

	for _, toplevel := range sites_in_sitemap {
		// This takes a toplevel site and populates a list with the links on that site
		// Parallelizing this toplevel processing does not help due to CPU and Network Limitations, I tried
		body := reqWrap(toplevel) // Assumption: timeouts on the website being deadlink-checked do not occur
		re := regexp.MustCompile(`"(https://.+?)"`)
		matches := re.FindAllStringSubmatch(body, -1)
		var links []string
		for _, matchSlice := range matches {
		  links = append(links, matchSlice[1]) // append the first submatch
		}

		for _, link := range links {
			wg.Add(1)
			go func(link string) {
				results <- isLinkAlive(link, toplevel)
				wg.Done()
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
