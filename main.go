package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getPlayers(link string) string {
	// Request the HTML page.
	res, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	//"Number of players:"
	// Could also be "tags"

	//return doc.Find(".workshopTags").First().Siblings().Find("a").First().Text()
	//strings.Contains(str, "abc")
	nOP := "Not Specified"
	doc.Find(".workshopTags").Each(func(i int, s *goquery.Selection) {
		s.Children().Each(func(i int, s *goquery.Selection) {
			title := s.Text()
			if strings.Contains(title, "Number of players") || strings.Contains(title, "Tags") {
				//nOP = title
				nOP = s.Siblings().Eq(0).Text()
			}
			//nOP = nOP + "|" + title
		})
	})

	return nOP
}

func main() {

	// Request the HTML page.
	res, err := http.Get("https://steamcommunity.com/workshop/browse/?appid=1435790&searchtext=&childpublishedfileid=0&browsesort=trend&section=readytouseitems&created_date_range_filter_start=0&created_date_range_filter_end=0&updated_date_range_filter_start=0&updated_date_range_filter_end=0")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find(".workshopItem").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		title := s.Find(".workshopItemTitle").Text()
		link := s.Find(".ugc").AttrOr("href", "NO LINK")
		players := getPlayers(link)
		fmt.Print(title + " - " + link + " - " + players + "\n\n")
	})
}
