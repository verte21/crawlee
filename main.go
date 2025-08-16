package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

var rawResultsFolderPath = "./raw-results"

func initFile(name string, folderPath string) (*os.File, error) {
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}

	fullPath := filepath.Join(rawResultsFolderPath, name+".txt")

	file, err := os.Create(fullPath)

	if err != nil {
		return nil, fmt.Errorf("error creating file: %v", err)
	}

	return file, nil
}

func extractDomain(fullURL string) (string, error) {
	u, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}

func extractSiteName(fullURL string) (string, error) {
	u, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}

	domain := u.Host

	domain = strings.TrimPrefix(domain, "www.")

	parts := strings.Split(domain, ".")
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", fmt.Errorf("could not extract site name from URL: %s", fullURL)
}

func initializeCollector(url string) *colly.Collector {
	allowedDomain, err := extractDomain(url)

	// log domain
	fmt.Printf("Allowed domain: %s\n", allowedDomain)

	if err != nil {
		fmt.Printf("Error extracting domain from URL %s: %v\n", url, err)
		return nil
	}

	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomain),
		colly.CacheDir("./cache"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       1 * time.Second,
	})

	return c

}

func isValidLink(url string, siteName string) bool {
	return strings.Contains(url, "http") &&
		strings.Contains(url, siteName) &&
		!strings.Contains(url, "mailto:") &&
		!strings.Contains(url, "tel:") &&
		!strings.Contains(url, "javascript:") &&
		!strings.Contains(url, "#") &&
		!strings.HasSuffix(url, ".pdf") &&
		!strings.HasSuffix(url, ".jpg") &&
		!strings.HasSuffix(url, ".png") &&
		!strings.HasSuffix(url, ".gif") &&
		!strings.HasSuffix(url, ".doc") &&
		!strings.HasSuffix(url, ".docx")
}

func scrapeWebsite(link string) {
	siteName, err := extractSiteName(link)
	if err != nil {
		fmt.Printf("Error extracting site name: %v\n", err)
		return
	}

	file, err := initFile(siteName, rawResultsFolderPath)
	if err != nil {
		fmt.Printf("Error initializing file: %v\n", err)
		return
	}

	c := initializeCollector(link)

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	linkCount := 0
	visitedLinks := make(map[string]bool)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)

		if isValidLink(absoluteURL, siteName) {
			if !visitedLinks[absoluteURL] {
				visitedLinks[absoluteURL] = true
				writer.WriteString(fmt.Sprintf("%s\n", absoluteURL))
			}

			c.Visit(absoluteURL)
		}

	})

	c.OnRequest(func(r *colly.Request) {
		linkCount++
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error visiting %s: %v\n", r.Request.URL, err)
	})

	c.Visit(link)

	fmt.Printf("Total links visited: %d\n", linkCount)
}

func readLinksFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var links []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			links = append(links, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return links, nil
}

func main() {
	links, err := readLinksFromFile("urls.txt")
	if err != nil {
		fmt.Printf("Error reading links from file: %v\n", err)
		return
	}

	var wg sync.WaitGroup

	for _, link := range links {
		wg.Add(1)

		go func(link string) {
			fmt.Printf("Starting scrape for: %s\n", link)

			defer wg.Done()
			scrapeWebsite(link)
		}(link)
	}

	wg.Wait()
	fmt.Println("Scraping completed for all websites.")

}
