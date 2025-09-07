package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Details struct {
	Heading string `json:"Heading"`
	Content string `json:"Content"`
}

type Link struct {
	Name string `json:"Name"`
	Url  string `json:"Url"`
}

func sendReq(url string, wg *sync.WaitGroup) {
	defer wg.Done()
	var allDetails []Details
	var allLinks []Link
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("User-Agent", "random")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var temp Details
	var urlTemp Link
	mainHeading := doc.Find("#firstHeading").Text()
	if len(strings.TrimSpace(mainHeading)) == 0 {
		parts := strings.Split(url, "/")
		mainHeading = parts[len(parts)-1]
	}
	details := ""
	temp.Heading = mainHeading
	doc.Find("#mw-content-text > div.mw-content-ltr.mw-parser-output > *").Each(func(i int, s *goquery.Selection) {
		element := goquery.NodeName(s)
		if element == "div" && s.HasClass("mw-heading") {
			temp.Content = details
			allDetails = append(allDetails, temp)
			temp.Heading = strings.TrimSpace(s.Text())
			details = ""
		} else if element == "p" {
			s.Find("a").Each(func(i int, q *goquery.Selection) {
				urlTemp.Name = q.Text()
				urlTemp.Url = "https://en.wikipedia.org" + q.AttrOr("href", "")
				allLinks = append(allLinks, urlTemp)
			})
			details += strings.TrimSpace(s.Text())
		}
	})
	if len(details) > 0 {
		temp.Content = details
		allDetails = append(allDetails, temp)
	}
	saveJson(allDetails, allLinks)
}

func saveJson(allDetails []Details, allLinks []Link) {
	content_jsonData, err := json.Marshal(allDetails)
	if err != nil {
		log.Fatal(err)
	}
	link_jsonData, err := json.Marshal(allLinks)
	if err != nil {
		log.Fatal(err)
	}
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	heading := strings.ReplaceAll(allDetails[0].Heading, " ", "_")
	if len(heading) > 20 {
		heading = heading[:20]
	}

	contentFileName := fmt.Sprintf("%s_%s.json", heading, timestamp)
	linkFileName := fmt.Sprintf("%s_links_%s.json", heading, timestamp)
	contentFile, err := os.Create(contentFileName)
	if err != nil {
		log.Fatal(err)
	}
	linkFile, err := os.Create(linkFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer contentFile.Close()
	defer linkFile.Close()
	_, err = contentFile.Write(content_jsonData)
	if err != nil {
		log.Fatal(err)
	}
	_, err = linkFile.Write(link_jsonData)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var wg sync.WaitGroup
	start := time.Now()
	urls := []string{"https://en.wikipedia.org/wiki/Riot_Games", "https://en.wikipedia.org/wiki/Valorant_Champions_Tour", "https://en.wikipedia.org/wiki/Valorant", "https://en.wikipedia.org/wiki/Microsoft_Windows", "https://en.wikipedia.org/wiki/Graphical_user_interface", "https://en.wikipedia.org/wiki/Command-line_interface"}
	for _, url := range urls {
		wg.Add(1)
		go sendReq(url, &wg)
	}
	wg.Wait()
	fmt.Println("time taken to execute", time.Since(start))
}
