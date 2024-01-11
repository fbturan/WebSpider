package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

var startURL = "https://prototurk.com"

func main() {
	db, _ := sql.Open("sqlite3", "web_spider.db")
	defer db.Close()
	
	deleteTableSQL := "DROP TABLE web_pages;"
	db.Exec(deleteTableSQL)

	createTable(db)
	crawl(startURL, db)
}

func createTable(db *sql.DB) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS web_pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT,
		title TEXT,
		content TEXT
	);
	`
	db.Exec(createTableSQL)
}

func crawl(url string, db *sql.DB) {

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed: Not get the page -", url)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("Failed: Not analyzed -", url)
		return
	}
	if fetchWebPages(url, db) == 0 {
		title, content := scrapePage(url)
		saveToDatabase(url, title, content, db)
		log.Printf("URL: %v\n", url)
		//incelemek için
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link,_ := s.Attr("href")

		if strings.HasPrefix(link, startURL) && fetchWebPages(link, db) == 0{
			crawl(link, db)
		}
	})
}

func fetchWebPages(targetURL string, db *sql.DB) int {
    var matchCount int
    err := db.QueryRow("SELECT COUNT(*) FROM web_pages WHERE url = ?", targetURL).Scan(&matchCount)
    if err != nil {
        log.Println("Failed to fetch web pages:", err)
    }
    return matchCount
}

func scrapePage(url string) (string, string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed: Not get the page-", url)
		return "", ""
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("Failed: Not analyze the page-", url)
		return "", ""
	}
	title := doc.Find("title").Text()
	content := doc.Find("body").Text()
	return title, content
}

func saveToDatabase(url string, title string, content string, db *sql.DB) {
	insertSQL := `
	INSERT INTO web_pages (url, title, content) VALUES (?, ?, ?);
	`
	_, err := db.Exec(insertSQL, url, title, content)
	if err != nil {
		log.Println("Failed: Not saved to database-", url)
	}
}
