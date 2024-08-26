package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

const helptext = `Usage: deadsniper [options] <link to sitemap.xml>

Options:
  -h | --help    print this help text
  -V | --version print the version number
  -s | --strict  allow only HTTP 200 response codes
  -t | --timeout set the request timeout in seconds (default 5)

Examples:
  deadsniper https://port19.xyz/sitemap.xml
  deadsniper -V
  deadsniper --strict https://port19.xyz/sitemap.xml
  deadsniper -t 1 -s https://port19.xyz/sitemap.xml`

var allowedStatusCodes = []int{
	http.StatusForbidden,       // Forbidden is very common. With or without a user agent
	http.StatusBadRequest,      // Bad Request -> e.g. gamefaqs.gamespot.com
	http.StatusTooManyRequests, // Too Many Requests -> e.g. geizhals.de
	http.StatusAccepted,        // Accepted -> e.g. DuckDuckGo
}
var exitCode int = 0
var timeout int = 5
var mutex sync.Mutex
var strict bool = false

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

func trapCode() {
	if exitCode != 1 {
		mutex.Lock() //Might be unneeded
		exitCode = 1
		mutex.Unlock()
	}
}

// contains checks if a slice contains a given int
func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// This checks a list of links. It also expects the corresponding toplevel, to make fixing easier
// This is an outside function to make it easily parallelizable
func isLinkAlive(url string, toplevel string) string {
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		trapCode()
		return fmt.Sprintf("❌: (rude/dead) %s -> %s\n", toplevel, url)
	} else if !strict && contains(allowedStatusCodes, resp.StatusCode) {
		return fmt.Sprintf("✓: (%d) %s -> %s\n", resp.StatusCode, toplevel, url)
	} else if resp.StatusCode != http.StatusOK {
		trapCode()
		return fmt.Sprintf("❌: (%d) %s -> %s\n", resp.StatusCode, toplevel, url)
	} else {
		return fmt.Sprintf("✓: (%d) %s -> %s\n", resp.StatusCode, toplevel, url)
	}
}

// Read man shift. This function kinda reimplements that posix shell builtin
func shift() {
	os.Args = os.Args[1:]
}

func main() {
	var body string
	// I swear, one more argument and I'll use the flag module
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println(helptext)
		return
	}
	if os.Args[1] == "-V" || os.Args[1] == "--version" {
		fmt.Println("v1.4 - 20240826")
		return
	}
	// Sorry Dijkstra
harmful:
	if os.Args[1] == "-s" || os.Args[1] == "--strict" {
		strict = true
		shift()
		goto harmful
	} else if os.Args[1] == "-t" || os.Args[1] == "--timeout" {
		var err error
		timeout, err = strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("timeout needs to be an integer", err)
		}
		shift()
		shift()
		goto harmful
	} else {
		body = reqWrap(os.Args[1]) // Assumption: timeouts on the website being deadlink-checked do not occur
	}

	// Parse all the requests from the sitemap file
	re := regexp.MustCompile(`<loc>(https?://.+?)</loc>`)
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
		re := regexp.MustCompile(`"(https?://.+?)"`)
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
