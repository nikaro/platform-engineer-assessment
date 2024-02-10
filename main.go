// This program, given any number of HTTP URLs through the "-u" command line argument,
// connect to each URL, extract all links from it, and depending on the “-o” option,
// output either one absolute URL per line, or a JSON hash where the key is the base domain,
// and the array is the relative paths.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

// Command line flags
var debug bool
var output string
var urls arrayVar

// Cutsom flag type to store multiple URLs
type arrayVar []url.URL

func (a *arrayVar) Set(s string) error {
	// Return an error if the URL is invalid
	if u, err := isHttpURL(s); err != nil {
		return err
	} else {
		*a = append(*a, *u)
	}
	return nil
}

func (a *arrayVar) String() string {
	// Reassemble the URL objects into strings
	var URLs []string
	for _, u := range *a {
		URLs = append(URLs, u.String())
	}
	return fmt.Sprintf("URLs: %s", URLs)
}

// Check if the string is a valid HTTP URL
func isHttpURL(s string) (*url.URL, error) {
	url, parseErr := url.Parse(s)
	if parseErr != nil {
		return nil, parseErr
	}
	if strings.Index(s, "http") != 0 {
		return nil, fmt.Errorf("URL must start with http:// or https://")
	}
	return url, nil
}

func extractLinks(url *url.URL, body io.Reader) (links []*url.URL, err error) {
	z := html.NewTokenizer(body)
	for z.Err() != io.EOF {
		tt := z.Next()
		slog.Debug("parsing", slog.String("url", url.String()), slog.String("token", tt.String()))
		switch tt {
		case html.ErrorToken:
			slog.Debug("parsing", slog.String("url", url.String()), slog.String("error", z.Err().Error()))
		case html.StartTagToken:
			t := z.Token()
			slog.Debug("parsing", slog.String("url", url.String()), slog.String("tag", t.Data))
			if t.Data == "a" {
				// Get the href attribute
				for _, a := range t.Attr {
					if a.Key == "href" {
						// Resolve relative URLs
						if a.Val[0] == '/' {
							a.Val = url.String() + a.Val
						}
						// Store the link
						if found, err := isHttpURL(a.Val); err == nil {
							slog.Debug("parsing", slog.String("url", url.String()), slog.String("found", found.String()))
							links = append(links, found)
						}
					}
				}
			}
		}
	}
	return links, nil
}

func main() {
	// Define the valid output formats
	var outputFormats = []string{"line", "json"}

	// Parse flags
	flag.Var(&urls, "u", "HTTP URL")
	flag.StringVar(&output, "o", "line", fmt.Sprintf("output format, could be: %s", outputFormats))
	flag.BoolVar(&debug, "d", false, "enable debug logging")
	flag.Parse()

	// Check if there are any URLs
	if len(urls) == 0 {
		fmt.Println("No URLs provided.")
		flag.Usage()
		os.Exit(2)
	}

	// Check if the output format is valid
	if !slices.Contains(outputFormats, output) {
		fmt.Println("Invalid output format.")
		flag.Usage()
		os.Exit(2)
	}

	// Configure logging level
	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	// Initialize links storage
	links := make(map[string][]string)

	// Connect to each URL
	for _, u := range urls {
		// Connect to the URL
		slog.Debug("connecting", slog.String("url", u.String()))
		resp, err := http.Get(u.String())
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer resp.Body.Close()
		slog.Debug("response", slog.String("url", u.String()), slog.Int("status", resp.StatusCode))

		// Extract all links from it
		if extracts, err := extractLinks(&u, resp.Body); err == nil {
			for _, link := range extracts {
				baseDomain := link.Scheme + "://" + link.Host
				links[baseDomain] = append(links[baseDomain], link.Path)
			}
		} else {
			slog.Error("extracting", slog.String("url", u.String()), slog.String("error", err.Error()))
		}
	}

	// Output either one absolute URL per line or a JSON hash
	switch output {
	case "line":
		for baseDomain, relativePaths := range links {
			for _, path := range relativePaths {
				fmt.Println(baseDomain + path)
			}
		}
	case "json":
		// Convert the map to JSON
		if jsonString, err := json.Marshal(links); err == nil {
			fmt.Println(string(jsonString))
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
