package workshop_scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cast"
	"gitlab.com/barry_nevio/goo-tools/Database"
)

type Props struct {
	NextPageLink              string
	IsLastPage                bool
	CurrentDoc                *goquery.Document
	CurrentIndiviualWSItemDoc *goquery.Document
	SQLVals                   []string
	DBConn                    Database.Connection
}

const SQL_FILE = "dump.sql"
const DEFAULT_INSERT_SQL = "INSERT INTO `escape_simulator_room` (`title`, `link`, `num_of_players`, `rating`, `cover_image`, `published_at_unix`) VALUES "
const DEFAULT_INSERT_PREPARED_VALUES_SQL = "(?, ?, ?, ?, ?, ?)"
const WRITE_DIRECT_TO_DB = true

// GetRating ... figures out the rating based on the image link
func GetRating(imgLink string) (rating int) {

	if strings.Contains(imgLink, "5-star") {
		return 5
	}

	if strings.Contains(imgLink, "4-star") {
		return 4
	}

	if strings.Contains(imgLink, "3-star") {
		return 3
	}

	if strings.Contains(imgLink, "2-star") {
		return 2
	}

	if strings.Contains(imgLink, "1-star") {
		return 1
	}

	return
}

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

	// Settdb conn
	err = p.setDBConn()

	if err != nil {
		return
	}

	pageNum := 1
	for p.setNextPageLink() || !p.IsLastPage {

		// A little output to the console
		fmt.Print("Working On Page: " + cast.ToString(pageNum) + " | Current Items: " + cast.ToString(len(p.SQLVals)) + "\n\n")

		// Set the SQL vales
		p.setSQLVals()

		// Set the new current doc.. if there is one to get
		if len(p.NextPageLink) > 0 {
			p.CurrentDoc, err = GetDoc(p.NextPageLink)
			if err != nil {
				// Write what we have anyway
				if !WRITE_DIRECT_TO_DB {
					p.writeSQLFile()
				}
				return
			}
		}

		// Increment the page number
		pageNum++

		/*if pageNum > 3 {
			break
		}*/
	}

	// Write directly to db?
	if !WRITE_DIRECT_TO_DB {
		p.writeSQLFile()
	}

	p.DBConn.Close()

	return
}

// setDBConn ...
func (p *Props) setDBConn() (err error) {
	// Open Settings File
	settingsReader, err := os.Open("./settings.json")
	if err != nil {
		return
	}
	defer settingsReader.Close()

	// Read settings file
	settingsBytes, err := ioutil.ReadAll(settingsReader)
	if err != nil {
		return
	}

	// Marshal
	dbSett := Database.Settings{}
	_ = json.Unmarshal(settingsBytes, &dbSett)

	// Close
	settingsReader.Close()

	// Add DB connection
	p.DBConn, err = Database.New(dbSett)

	return
}

// writeSQLFile ...
func (p *Props) writeSQLFile() {
	f, err := os.OpenFile(SQL_FILE, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(DEFAULT_INSERT_SQL + "\n"); err != nil {
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

		// Download the doc for this item
		p.CurrentIndiviualWSItemDoc, _ = GetDoc(link)

		// Get number of players
		players, _ := p.getNumOfPlayers()

		// Get rating
		rating := GetRating(s.Find(".fileRating").AttrOr("src", ""))

		// get cover image
		coverImage := p.getCoverImage()

		// Get date published unix
		datePublishedUnix := s.Find(".ugc").AttrOr("data-publishedfileid", "")

		// Clear this item doc
		p.CurrentIndiviualWSItemDoc = nil

		// Write directly to db?
		if WRITE_DIRECT_TO_DB {
			db, _ := p.DBConn.Get()
			_, _ = db.Exec(DEFAULT_INSERT_SQL+DEFAULT_INSERT_PREPARED_VALUES_SQL,
				title,
				link,
				players,
				rating,
				coverImage,
				datePublishedUnix,
			)
		} else {
			p.SQLVals = append(p.SQLVals, "('"+MysqlRealEscapeString(title)+"', '"+MysqlRealEscapeString(link)+"', '"+MysqlRealEscapeString(players)+"', "+cast.ToString(rating)+", '"+MysqlRealEscapeString(coverImage)+"', '"+MysqlRealEscapeString(datePublishedUnix)+"')")
		}

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
func (p *Props) getNumOfPlayers() (nOP string, err error) {

	nOP = "Not Specified"

	// Get the number of players
	p.CurrentIndiviualWSItemDoc.Find(".workshopTags").Each(func(i int, s *goquery.Selection) {
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

// getCoverImage ...
func (p *Props) getCoverImage() (coverImage string) {
	return p.CurrentIndiviualWSItemDoc.Find("#previewImage").AttrOr("src", "")
}
