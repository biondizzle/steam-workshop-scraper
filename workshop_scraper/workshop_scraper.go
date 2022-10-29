package workshop_scraper

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cast"
)

type Props struct {
	NextPageLink string
	IsLastPage   bool
	CurrentDoc   *goquery.Document
	SQLVals      []string
}

const SQL_FILE = "dump.sql"

// MysqlRealEscapeString ...
func MysqlRealEscapeString(value string) string {
	var sb strings.Builder
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch c {
		case '\\', 0, '\n', '\r', '\'', '"':
			sb.WriteByte('\\')
			sb.WriteByte(c)
		case '\032':
			sb.WriteByte('\\')
			sb.WriteByte('Z')
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

// GetDoc ... downloads the page and converts it to a goquery.Document
func GetDoc(link string) (doc *goquery.Document, err error) {
	// Request the HTML page.
	res, err := http.Get(link)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return doc, errors.New("Error Opening Link: " + cast.ToString(res.StatusCode) + " - " + res.Status)
	}

	// Load the HTML document
	return goquery.NewDocumentFromReader(res.Body)
}

// Scrape ... initilize scraping. This is the entry point
func Scrape(firstPageLink string) (err error) {
	p := Props{}
	p.CurrentDoc, err = GetDoc(firstPageLink)
	if err != nil {
		return
	}

	pageNum := 1
	for p.setNextPageLink() || !p.IsLastPage {

		// A little output to the console
		fmt.Print("Working On Page: " + cast.ToString(pageNum) + " | Current Items: " + cast.ToString(len(p.SQLVals)) + "\n\n")

		// Set the SQL vales
		p.setSQLVals()

		// Set the new current doc
		p.CurrentDoc, err = GetDoc(p.NextPageLink)
		if err != nil {
			return
		}

		// Increment the page number
		pageNum++

		/*if pageNum > 3 {
			break
		}*/
	}

	p.writeSQLFile()

	return
}

// writeSQLFile ...
func (p *Props) writeSQLFile() {
	f, err := os.OpenFile(SQL_FILE, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString("INSERT INTO `escape_simulator_room` (`title`, `link`, `num_of_players`) VALUES \n"); err != nil {
		panic(err)
	}

	if _, err = f.WriteString(strings.Join(p.SQLVals, ", \n")); err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		panic(err)
	}
}

// buildSQLString ...
func (p *Props) setSQLVals() {
	p.CurrentDoc.Find(".workshopItem").Each(func(i int, s *goquery.Selection) {
		// Get title
		title := s.Find(".workshopItemTitle").Text()

		// get link of workshop item
		link := s.Find(".ugc").AttrOr("href", "")

		// Get number of players
		players, _ := p.getNumOfPlayers(link)

		p.SQLVals = append(p.SQLVals, "('"+MysqlRealEscapeString(title)+"', '"+MysqlRealEscapeString(link)+"', '"+MysqlRealEscapeString(players)+"')")
	})
}

// setNextPageLink .. gets the link of the next page
func (p *Props) setNextPageLink() bool {

	// Already on last page, do nothing
	if p.IsLastPage {
		return false
	}

	p.NextPageLink = p.CurrentDoc.Find(".pagebtn").Eq(1).AttrOr("href", "")

	if len(p.NextPageLink) < 1 {
		p.IsLastPage = true
	}

	return true
}

// getNumOfPlayers ...
func (p *Props) getNumOfPlayers(link string) (nOP string, err error) {

	nOP = "Not Specified"

	// Link empty
	if len(link) < 1 {
		return
	}

	// DOwnload the doc
	doc, err := GetDoc(link)

	if err != nil {
		return
	}

	// Get the number of players
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

	return
}
