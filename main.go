package main

import (
	"./workshop_scraper"
)

func main() {

	workshop_scraper.Scrape("https://steamcommunity.com/workshop/browse/?appid=1435790&browsesort=trend&section=readytouseitems&requiredtags=2+or+more+players&created_date_range_filter_start=0&created_date_range_filter_end=0&updated_date_range_filter_start=0&updated_date_range_filter_end=0&actualsort=trend&p=1&days=-1")

}
